import type { ExamAttempt } from "@engflex/contracts";
import { useQueryClient } from "@tanstack/react-query";
import { useHydrated, useNavigate } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import {
	examKeys,
	useCancelExamAttempt,
	usePatchExamAttempt,
} from "#/features/exams/queries";
import { m } from "#/paraglide/messages";
import { IeltsAudio } from "./ielts-audio";
import { IeltsConfirmModal } from "./ielts-confirm-modal";
import { IeltsHeader } from "./ielts-header";
import { IeltsPassagePane } from "./ielts-passage-pane";
import { IeltsQuestionPalette } from "./ielts-question-palette";
import { IeltsQuestionPane } from "./ielts-question-pane";
import { IeltsTabBar } from "./ielts-tab-bar";

export function IeltsSession({
	initialAttempt: attempt,
}: {
	initialAttempt: ExamAttempt;
}) {
	const navigate = useNavigate();
	const hydrated = useHydrated();
	const queryClient = useQueryClient();
	const patch = usePatchExamAttempt();
	const cancel = useCancelExamAttempt();
	const content = attempt.examTest?.ieltsContent;
	const sections = content?.sections ?? [];
	const [activeSection, setActiveSection] = useState(
		sections[0]?.sectionNumber ?? 1,
	);
	const [activeQuestionId, setActiveQuestionId] = useState<string | null>(null);
	const [highlightMode, setHighlightMode] = useState(false);
	const [highlightVersion, setHighlightVersion] = useState(0);
	const [collapsed, setCollapsed] = useState(false);
	const [action, setAction] = useState<"submit" | "exit" | "reset" | null>(
		null,
	);
	const [busy, setBusy] = useState(false);
	const [saveFailed, setSaveFailed] = useState(false);
	const [saved, setSaved] = useState(false);
	const changes = useRef(0);
	const [answers, setAnswers] = useState<Record<string, string>>(() =>
		Object.fromEntries(
			(attempt.ieltsResult?.answers ?? []).map((answer) => [
				answer.questionId,
				answer.selectedKey ??
					answer.textValue ??
					answer.selectedKeys?.join(",") ??
					"",
			]),
		),
	);
	const [bookmarks, setBookmarks] = useState(
		() =>
			new Set(
				(attempt.ieltsResult?.answers ?? [])
					.filter((answer) => answer.flagged)
					.map((answer) => answer.questionId),
			),
	);
	const [timeLeft, setTimeLeft] = useState<number | null>(() =>
		attempt.deadlineAt && attempt.startedAt
			? Math.max(
					0,
					Math.ceil(
						(new Date(attempt.deadlineAt).getTime() -
							new Date(attempt.startedAt).getTime()) /
							1000,
					),
				)
			: null,
	);
	const answersRef = useRef(answers);
	const bookmarksRef = useRef(bookmarks);
	const revision = useRef(attempt.ieltsResult?.revision ?? 1);
	const debounce = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
	const queue = useRef<Promise<unknown>>(Promise.resolve());
	const locked = useRef(false);
	const submitRef = useRef<() => Promise<void>>(async () => {});

	useEffect(() => () => clearTimeout(debounce.current), []);
	useEffect(() => {
		if (!attempt.deadlineAt || attempt.mode !== "exam") return;
		const deadline = new Date(attempt.deadlineAt).getTime();
		const tick = () => {
			const remaining = Math.max(0, Math.ceil((deadline - Date.now()) / 1000));
			setTimeLeft(remaining);
			if (!remaining) {
				clearInterval(timer);
				void submitRef.current();
			}
		};
		const timer = setInterval(tick, 1000);
		tick();
		return () => clearInterval(timer);
	}, [attempt.deadlineAt, attempt.mode]);

	const save = (submit = false) => {
		clearTimeout(debounce.current);
		setSaved(false);
		const savedVersion = changes.current;
		const snapshot = answersRef.current;
		const flags = bookmarksRef.current;
		const payload = sections.flatMap((section) =>
			section.questionGroups.flatMap((group) =>
				group.questions.map((question) => {
					const value = snapshot[question.id] || null;
					const multi = group.type === "multiple_choice_multiple";
					return {
						questionId: question.id,
						selectedKey: question.options?.length && !multi ? value : null,
						selectedKeys: multi && value ? value.split(",") : [],
						textValue: question.options?.length ? null : value,
						flagged: flags.has(question.id),
					};
				}),
			),
		);
		const pending = queue.current
			.catch(() => {})
			.then(async () => {
				const updated = await patch.mutateAsync({
					attemptId: attempt.id,
					req: { revision: revision.current, submit, answers: payload },
				});
				revision.current = updated.ieltsResult?.revision ?? revision.current;
				queryClient.setQueryData(examKeys.attempt(attempt.id), updated);
				setSaveFailed(false);
				setSaved(changes.current === savedVersion);
			});
		queue.current = pending;
		return pending;
	};
	const scheduleSave = () => {
		changes.current += 1;
		setSaved(false);
		clearTimeout(debounce.current);
		debounce.current = setTimeout(() => {
			void save().catch(() => {
				setSaveFailed(true);
				toast.error(m["exams.ielts.take.saveFailed"]());
			});
		}, 1000);
	};
	const perform = async (operation: "submit" | "exit" | "reset") => {
		if (locked.current) return;
		locked.current = true;
		setBusy(true);
		try {
			if (operation === "reset") {
				changes.current += 1;
				setSaved(false);
				answersRef.current = {};
				bookmarksRef.current = new Set();
				setAnswers({});
				setBookmarks(new Set());
				setHighlightVersion((version) => version + 1);
			}
			await save(operation === "submit");
			if (operation === "exit") await cancel.mutateAsync(attempt.id);
			setAction(null);
			if (operation !== "reset") {
				await navigate(
					operation === "submit"
						? {
								to: APP_ROUTES.EXAMS.IELTS.RESULT,
								params: { id: attempt.id },
								replace: true,
							}
						: { to: APP_ROUTES.EXAMS.IELTS.HISTORY, replace: true },
				);
			}
		} catch {
			setSaveFailed(true);
			toast.error(m["exams.ielts.take.saveFailed"]());
		} finally {
			locked.current = false;
			setBusy(false);
		}
	};
	submitRef.current = () => perform("submit");
	const current =
		sections.find((section) => section.sectionNumber === activeSection) ??
		sections[0];
	const palette = sections.map((section) => ({
		sectionNumber: section.sectionNumber,
		title: section.title,
		questions: section.questionGroups.flatMap((group) =>
			group.questions.map((question) => ({
				id: question.id,
				questionNumber: question.questionNumber,
			})),
		),
	}));
	const tabs = palette.map((section) => ({
		sectionNumber: section.sectionNumber,
		label: m["exams.ielts.take.section"]({ number: section.sectionNumber }),
		totalQuestions: section.questions.length,
		answeredCount: section.questions.filter((question) =>
			answers[question.id]?.trim(),
		).length,
	}));
	const total = palette.reduce(
		(count, section) => count + section.questions.length,
		0,
	);
	const questionProps = {
		groups: current?.questionGroups ?? [],
		answers,
		bookmarks,
		activeQuestionId,
		disabled: busy || !hydrated,
		onAnswerChange: (id: string, value: string) => {
			answersRef.current = { ...answersRef.current, [id]: value };
			setAnswers(answersRef.current);
			scheduleSave();
		},
		onToggleBookmark: (id: string) => {
			const next = new Set(bookmarksRef.current);
			if (next.has(id)) next.delete(id);
			else next.add(id);
			bookmarksRef.current = next;
			setBookmarks(next);
			scheduleSave();
		},
	};
	const confirmation = (
		<IeltsConfirmModal
			action={action}
			answeredCount={
				Object.values(answers).filter((value) => value.trim()).length
			}
			totalCount={total}
			pending={busy}
			onClose={() => {
				if (!busy) setAction(null);
			}}
			onConfirm={() => {
				if (action) void perform(action);
			}}
		/>
	);

	return (
		<div className="flex h-[calc(100dvh-3.5rem)] w-full flex-col overflow-hidden bg-background">
			<IeltsHeader
				testTitle={attempt.examTest?.title ?? m["exams.ielts.hub.title"]()}
				timeLeft={timeLeft}
				highlightMode={highlightMode}
				onToggleHighlight={() => setHighlightMode((value) => !value)}
				onSubmit={() => setAction("submit")}
				onExit={() => setAction("exit")}
				onReset={() => setAction("reset")}
				pending={busy || !hydrated}
			/>
			{saveFailed && (
				<p
					role="alert"
					className="shrink-0 bg-destructive/10 p-2 text-center text-sm text-destructive"
				>
					{m["exams.ielts.take.saveFailed"]()}
				</p>
			)}
			{!saveFailed && saved && (
				<output className="sr-only">{m["exams.ielts.take.saved"]()}</output>
			)}
			{current?.audioAssetId && (
				<div className="shrink-0 px-4 py-2">
					<IeltsAudio
						key={current.audioAssetId}
						identity={current.audioAssetId}
					/>
				</div>
			)}
			<div
				inert={!hydrated || busy}
				className={`grid min-h-0 flex-1 ${current?.passages.length ? "grid-cols-1 grid-rows-2 md:grid-cols-2 md:grid-rows-1" : "grid-cols-1"}`}
			>
				{Boolean(current?.passages.length) && (
					<IeltsPassagePane
						key={`${activeSection}-${highlightVersion}`}
						passages={current?.passages ?? []}
						highlightMode={highlightMode}
					/>
				)}
				<IeltsQuestionPane {...questionProps} />
			</div>
			<div
				className="shrink-0 border-t border-border"
				inert={!hydrated || busy}
			>
				<IeltsTabBar
					sections={tabs}
					activeSection={activeSection}
					onSelectSection={setActiveSection}
				/>
				<IeltsQuestionPalette
					groups={palette}
					answers={answers}
					bookmarks={bookmarks}
					activeQuestionId={activeQuestionId}
					isCollapsed={collapsed}
					onToggleCollapse={() => setCollapsed((value) => !value)}
					onSelectQuestion={(id, section) => {
						setActiveSection(section);
						setActiveQuestionId(id);
					}}
				/>
			</div>
			{confirmation}
		</div>
	);
}
