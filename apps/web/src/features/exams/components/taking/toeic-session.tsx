import { useAuth } from "@clerk/tanstack-react-start";
import type {
	ExamAnswer,
	ExamAttempt,
	ExamAttemptSummary,
	ExamReviewAttempt,
} from "@engflex/contracts";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useBlocker, useNavigate } from "@tanstack/react-router";
import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { CardShell } from "#/components/common/card-shell";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import { clearAudioPosition } from "#/features/exams/audio-position";
import { toeicPartLabel } from "#/features/exams/parts";
import {
	examAttemptQueryOptions,
	examKeys,
	useCancelExamAttempt,
	usePatchExamAttempt,
} from "#/features/exams/queries";
import { getExamAttempt } from "#/features/exams/service";
import { ExamSessionController } from "#/features/exams/session-controller";
import { ApiError } from "#/lib/api/api";
import { m } from "#/paraglide/messages";
import { SessionItem } from "./session-item";
import { ToeicExamHeader } from "./toeic-exam-header";
import { SessionConfirmModal } from "./toeic-exam-modals";
import { ToeicExamTabBar } from "./toeic-exam-tab-bar";
import { ToeicQuestionPalette } from "./toeic-question-palette";

export function SessionPending() {
	return (
		<div className="fixed inset-0 z-50 flex flex-col gap-4 bg-background p-4">
			<Skeleton className="h-14 w-full" />
			<Skeleton className="h-11 w-full" />
			<div className="grid flex-1 gap-4 md:grid-cols-[1fr_18rem]">
				<CardShell
					pending
					skeleton={<Skeleton className="h-full min-h-96 w-full" />}
				/>
				<CardShell
					pending
					skeleton={<Skeleton className="h-full min-h-96 w-full" />}
				/>
			</div>
		</div>
	);
}

export function ToeicSession({
	initialAttempt,
}: {
	initialAttempt: ExamAttempt;
}) {
	const { id } = initialAttempt;
	const { userId } = useAuth();
	const audioOwner = useRef(userId);
	audioOwner.current = userId;
	const queryClient = useQueryClient();
	const navigate = useNavigate();
	const patch = usePatchExamAttempt();
	const cancel = useCancelExamAttempt();
	const [navigationError, setNavigationError] = useState(false);
	const destination = useRef<"completed" | "abandoned" | null>(null);
	const departing = useRef(false);
	function openClosed(status: "completed" | "abandoned") {
		destination.current = status;
		departing.current = true;
		setNavigationError(false);
		const test = initialAttempt.examTest;
		const navigation =
			status === "completed"
				? navigate({
						to: APP_ROUTES.EXAMS.TOEIC.RESULT,
						params: { id },
						replace: true,
					})
				: test
					? navigate({
							to: APP_ROUTES.EXAMS.TOEIC.TEST,
							params: { categoryId: test.categoryId, testId: test.id },
							replace: true,
						})
					: navigate({ to: APP_ROUTES.EXAMS.TOEIC.HISTORY, replace: true });
		void navigation.catch(() => setNavigationError(true));
	}
	const audio = useRef<HTMLAudioElement | null>(null);
	const mutationFns = useRef({
		patch: patch.mutateAsync,
		cancel: cancel.mutateAsync,
	});
	const { data: attempt } = useQuery({
		...examAttemptQueryOptions(id),
		initialData: initialAttempt,
		// The writer refreshes before recovering errors. Background refetch must
		// not race a PATCH or silently replace another tab's revision.
		staleTime: Infinity,
		refetchOnWindowFocus: false,
		refetchOnReconnect: false,
		refetchOnMount: false,
	});
	const [controller] = useState(
		() =>
			new ExamSessionController({
				read: () =>
					queryClient.getQueryData<ExamAttempt | ExamReviewAttempt>(
						examKeys.attempt(id),
					) ?? initialAttempt,
				accept: (value) =>
					queryClient.setQueryData(examKeys.attempt(id), value),
				get: async () => {
					try {
						return await getExamAttempt(id);
					} catch (error) {
						if (
							error instanceof ApiError &&
							[401, 404, 409].includes(error.status)
						) {
							departing.current = true;
							if (error.status === 401)
								window.location.assign(APP_ROUTES.AUTH.SIGN_IN);
							else if (error.status === 404) window.location.reload();
							else
								await navigate({
									to: APP_ROUTES.EXAMS.TOEIC.HISTORY,
									replace: true,
								});
						}
						throw error;
					}
				},
				patch: (revision, answers, submit) =>
					mutationFns.current.patch({
						attemptId: id,
						req: { revision, answers, submit },
					}),
				cancel: () => mutationFns.current.cancel(id),
				closed: (
					value: ExamAttempt | ExamReviewAttempt | ExamAttemptSummary,
				) => {
					departing.current = true;
					audio.current?.pause();
					if (audioOwner.current) clearAudioPosition(audioOwner.current, id);
					void queryClient.invalidateQueries({
						queryKey: [...examKeys.all, "attempts"],
					});
					void queryClient.invalidateQueries({ queryKey: examKeys.stats });
					openClosed(value.status === "completed" ? "completed" : "abandoned");
				},
			}),
	);
	const state = useSyncExternalStore(
		controller.subscribe,
		controller.getSnapshot,
		controller.getSnapshot,
	);
	const [activePart, setActivePart] = useState(
		initialAttempt.examTest?.content?.parts[0]?.partNumber ?? 1,
	);
	const [focusedQuestion, setFocusedQuestion] = useState<string | null>(null);
	const [collapsed, setCollapsed] = useState(false);
	const [speed, setSpeed] = useState(1);
	const [action, setAction] = useState<"submit" | "cancel" | null>(null);
	// Start with a server/client-stable value; the clock subscription takes
	// over after hydration and on visibility changes.
	const [now, setNow] = useState<number | null>(null);
	const deadline = attempt.deadlineAt ? Date.parse(attempt.deadlineAt) : null;
	const expired = deadline !== null && now !== null && now >= deadline;
	const terminalBusy = state.terminal;
	const locked = state.terminal || expired || attempt.status !== "in_progress";
	const blocker = useBlocker({
		shouldBlockFn: () => !departing.current,
		withResolver: true,
		enableBeforeUnload: () => !departing.current,
	});
	useEffect(() => {
		// Scroll after a newly selected part mounts, once per navigation intent.
		// Clock ticks and autosave responses must not pull the reader back.
		if (focusedQuestion)
			document
				.getElementById(focusedQuestion)
				?.scrollIntoView({ behavior: "smooth", block: "start" });
	}, [focusedQuestion]);
	useEffect(() => {
		controller.activate();
		return () => controller.dispose();
	}, [controller]);
	useEffect(() => {
		const tick = () => setNow(Date.now());
		const interval = setInterval(tick, 1000);
		document.addEventListener("visibilitychange", tick);
		return () => {
			clearInterval(interval);
			document.removeEventListener("visibilitychange", tick);
		};
	}, []);
	useEffect(() => {
		if (deadline === null) return;
		// External deadline timer, independent of modal state or background tab throttling.
		const finish = () => {
			if (Date.now() >= deadline) {
				audio.current?.pause();
				void controller.run("submit");
			}
		};
		const timer = setTimeout(finish, Math.max(0, deadline - Date.now()));
		document.addEventListener("visibilitychange", finish);
		return () => {
			clearTimeout(timer);
			document.removeEventListener("visibilitychange", finish);
		};
	}, [controller, deadline]);
	const content = attempt.examTest?.content;
	if (!content || !attempt.result)
		throw new Error("Invalid exam session content");
	const answers = new Map<string, ExamAnswer>(
		attempt.result.answers.map((answer) => [answer.questionId, answer]),
	);
	for (const [questionId, draft] of state.drafts)
		answers.set(questionId, draft.answer);
	const parts = content.parts;
	const groups = parts.map((part) => ({
		partNum: part.partNumber,
		title: `P${part.partNumber} — ${toeicPartLabel(part.partNumber)}`,
		questions: part.items.flatMap((item) =>
			item.questions.map((question) => ({
				id: question.id,
				questionNumber: question.questionNumber,
			})),
		),
	}));
	const answered = Object.fromEntries(
		[...answers]
			.filter(([, answer]) => answer.selectedKey !== null)
			.map(([key, answer]) => [key, answer.selectedKey as string]),
	);
	const bookmarks = new Set(
		[...answers].filter(([, answer]) => answer.flagged).map(([key]) => key),
	);
	const visible = parts.find((part) => part.partNumber === activePart);
	const seconds =
		now === null
			? 0
			: deadline === null
				? Math.max(
						0,
						Math.floor(
							(now - Date.parse(attempt.startedAt ?? attempt.createdAt)) / 1000,
						),
					)
				: Math.max(0, Math.ceil((deadline - now) / 1000));
	const currentAction =
		action ?? (blocker.status === "blocked" ? "cancel" : null);
	return (
		<div className="fixed inset-0 z-50 flex flex-col overflow-hidden bg-background text-foreground">
			<ToeicExamHeader
				testTitle={attempt.examTest?.title ?? content.title}
				timeLeft={seconds}
				practice={attempt.mode === "practice"}
				speed={speed}
				pending={terminalBusy}
				locked={locked}
				onCycleSpeed={() =>
					setSpeed((value) => (value === 1 ? 1.25 : value === 1.25 ? 1.5 : 1))
				}
				onSubmitClick={() => setAction("submit")}
				onExitClick={() => setAction("cancel")}
			/>
			<ToeicExamTabBar
				activePart={activePart}
				parts={groups.map((group) => ({
					partNum: group.partNum,
					label: `P${group.partNum}`,
					section: group.partNum <= 4 ? "listening" : "reading",
					totalQuestions: group.questions.length,
					answeredCount: group.questions.filter(
						(question) => answered[question.id] != null,
					).length,
				}))}
				onSelectPart={(part) => {
					audio.current?.pause();
					setActivePart(part);
					setFocusedQuestion(null);
				}}
			/>
			<div
				className="flex items-center justify-between gap-2 border-b px-4 py-2 text-xs"
				aria-live="polite"
			>
				<span role={state.error ? "alert" : "status"}>
					{state.error
						? m["exams.toeic.take.saveFailed"]()
						: state.busy || state.drafts.size
							? m["exams.toeic.take.saving"]()
							: m["exams.toeic.take.saved"]()}
				</span>
				{(state.error || expired) && (
					<SubmitButton
						size="sm"
						pending={!!state.busy}
						onClick={() => {
							void controller.run(expired ? "submit" : "save");
						}}
					>
						{m["exams.toeic.take.retry"]()}
					</SubmitButton>
				)}
			</div>
			<div className="flex min-h-0 flex-1 overflow-hidden">
				<main className="min-w-0 flex-1 overflow-y-auto p-3 md:p-5">
					<fieldset disabled={locked} className="flex min-w-0 flex-col gap-5">
						{visible?.items.map((item) => (
							<SessionItem
								key={item.id}
								item={item}
								assets={content.assets}
								answers={answers}
								speed={speed}
								locked={locked}
								onEdit={(answer) => {
									if (!locked) controller.edit(answer);
								}}
								onPlay={(next) => {
									if (locked) {
										next.pause();
										return;
									}
									if (audio.current !== next) audio.current?.pause();
									audio.current = next;
								}}
							/>
						))}
					</fieldset>
				</main>
				<aside className="shrink-0">
					<ToeicQuestionPalette
						isCollapsed={collapsed}
						onToggleCollapse={() => setCollapsed((value) => !value)}
						groups={groups}
						activeQuestionId={focusedQuestion}
						answers={answered}
						bookmarks={bookmarks}
						onSelectQuestion={(questionId) => {
							const group = groups.find((group) =>
								group.questions.some((question) => question.id === questionId),
							);
							if (group) {
								if (activePart !== group.partNum) audio.current?.pause();
								setActivePart(group.partNum);
								setFocusedQuestion(questionId);
								document
									.getElementById(questionId)
									?.scrollIntoView({ behavior: "smooth", block: "start" });
							}
						}}
					/>
				</aside>
			</div>
			{navigationError && (
				<div className="border-t bg-background p-4" role="alert">
					<p>{m["exams.toeic.take.navigationFailed"]()}</p>
					<Button
						onClick={() => {
							if (destination.current) openClosed(destination.current);
						}}
					>
						{m["exams.toeic.take.retry"]()}
					</Button>
				</div>
			)}

			<SessionConfirmModal
				action={navigationError ? null : currentAction}
				answered={Object.keys(answered).length}
				total={answers.size}
				pending={terminalBusy}
				error={state.error}
				onClose={() => {
					setAction(null);
					blocker.reset?.();
				}}
				onConfirm={() => {
					if (currentAction) {
						audio.current?.pause();
						void controller.run(currentAction);
					}
				}}
			/>
		</div>
	);
}
