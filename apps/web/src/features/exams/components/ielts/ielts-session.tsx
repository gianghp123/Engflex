import type { ExamAnswer, ExamAttempt } from "@engflex/contracts";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import type {
	IeltsPassage,
	IeltsQuestionGroup,
	IeltsQuestionType,
} from "#/features/exams/ielts/mock-data";
import { usePatchExamAttempt } from "#/features/exams/queries";
import { getExamAttempt } from "#/features/exams/service";
import { ApiError } from "#/lib/api/api";
import { m } from "#/paraglide/messages";
import { IeltsConfirmModal } from "./ielts-confirm-modal";
import { IeltsHeader } from "./ielts-header";
import { IeltsPassagePane } from "./ielts-passage-pane";
import {
	type IeltsPaletteGroup,
	IeltsQuestionPalette,
} from "./ielts-question-palette";
import { IeltsQuestionPane } from "./ielts-question-pane";
import { IeltsTabBar } from "./ielts-tab-bar";

export function IeltsSession({
	initialAttempt,
}: {
	initialAttempt: ExamAttempt;
}) {
	const navigate = useNavigate();
	const patch = usePatchExamAttempt();

	const attempt = initialAttempt;
	const ieltsContent = attempt.examTest?.ieltsContent;

	const sections = ieltsContent?.sections ?? [];
	const [activeSection, setActiveSection] = useState<number>(
		sections[0]?.sectionNumber ?? 1,
	);
	const [activeQuestionId, setActiveQuestionId] = useState<string | null>(null);
	const [highlightMode, setHighlightMode] = useState<boolean>(false);
	const [isPaletteCollapsed, setIsPaletteCollapsed] = useState<boolean>(false);
	const [actionModal, setActionModal] = useState<"submit" | "exit" | null>(
		null,
	);

	const [answers, setAnswers] = useState<Record<string, string>>(() => {
		const map: Record<string, string> = {};
		if (attempt.ieltsResult?.answers) {
			for (const a of attempt.ieltsResult.answers) {
				if (a.selectedKey) map[a.questionId] = a.selectedKey;
				else if (a.textValue) map[a.questionId] = a.textValue;
				else if (a.selectedKeys && a.selectedKeys.length > 0)
					map[a.questionId] = a.selectedKeys.join(",");
			}
		}
		return map;
	});

	const [bookmarks, setBookmarks] = useState<Set<string>>(() => {
		const set = new Set<string>();
		if (attempt.ieltsResult?.answers) {
			for (const a of attempt.ieltsResult.answers) {
				if (a.flagged) set.add(a.questionId);
			}
		}
		return set;
	});

	const [timeLeft, setTimeLeft] = useState<number>(() => {
		if (attempt.deadlineAt) {
			const diff = Math.floor(
				(new Date(attempt.deadlineAt).getTime() - Date.now()) / 1000,
			);
			return Math.max(0, diff);
		}
		return 3600;
	});

	const revision = useRef(attempt.ieltsResult?.revision ?? 1);
	const drafts = useRef(new Map<string, ExamAnswer>());
	const timer = useRef<ReturnType<typeof setTimeout>>(undefined);
	const queue = useRef<Promise<void>>(Promise.resolve());
	const closed = useRef(false);
	const finalizing = useRef(false);
	const expire = useRef<() => Promise<void>>(async () => {});
	const questionById = new Map(
		sections.flatMap((section) =>
			section.questionGroups.flatMap((group) =>
				group.questions.map((question) => [question.id, question] as const),
			),
		),
	);

	useEffect(() => {
		if (attempt.mode !== "exam" || !attempt.deadlineAt) return;
		const deadline = Date.parse(attempt.deadlineAt);
		const tick = () => {
			const remaining = Math.max(0, Math.ceil((deadline - Date.now()) / 1000));
			setTimeLeft(remaining);
			if (remaining === 0 && !finalizing.current && !closed.current)
				void expire.current();
		};
		const interval = setInterval(tick, 1000);
		window.addEventListener("focus", tick);
		return () => {
			clearInterval(interval);
			window.removeEventListener("focus", tick);
			clearTimeout(timer.current);
		};
	}, [attempt.deadlineAt, attempt.mode]);
	useEffect(() => () => clearTimeout(timer.current), []);

	function saveToServer(submit = false) {
		clearTimeout(timer.current);
		const job = queue.current.then(async () => {
			if (closed.current || (!submit && drafts.current.size === 0)) return;
			const batch = [...drafts.current.values()];
			const write = () =>
				patch.mutateAsync({
					attemptId: attempt.id,
					req: { revision: revision.current, submit, answers: batch },
				});
			let updated: Awaited<ReturnType<typeof write>>;
			try {
				updated = await write();
			} catch (error) {
				if (!(error instanceof ApiError && error.status === 409)) throw error;
				const fresh = await getExamAttempt(attempt.id);
				if (fresh.status === "completed") {
					updated = fresh;
				} else {
					revision.current = fresh.ieltsResult?.revision ?? revision.current;
					updated = await write();
				}
			}
			revision.current = updated.ieltsResult?.revision ?? revision.current;
			for (const sent of batch) {
				if (drafts.current.get(sent.questionId) === sent)
					drafts.current.delete(sent.questionId);
			}
			if (updated.status === "completed") {
				closed.current = true;
				await navigate({
					to: APP_ROUTES.EXAMS.IELTS.RESULT,
					params: { id: attempt.id },
					replace: true,
				});
			}
		});
		queue.current = job.catch(() => {});
		return job;
	}

	function scheduleSave(questionId: string, value: string, flagged: boolean) {
		if (finalizing.current || closed.current) return;
		const choice = Boolean(questionById.get(questionId)?.options?.length);
		drafts.current.set(questionId, {
			questionId,
			selectedKey: choice ? value || null : null,
			textValue: choice ? null : value || null,
			flagged,
		});
		clearTimeout(timer.current);
		timer.current = setTimeout(() => {
			void saveToServer().then(
				() => {
					if (!closed.current && !drafts.current.size)
						toast.success(m["exams.ielts.take.saved"]());
				},
				() => toast.error(m["exams.ielts.take.saveFailed"]()),
			);
		}, 500);
	}

	const handleAnswerChange = (questionId: string, displayed: string) => {
		if (finalizing.current || closed.current) return;
		const choice = questionById
			.get(questionId)
			?.options?.find(
				(option) => `${option.key}. ${option.text}` === displayed,
			);
		const value = choice?.key ?? displayed;
		setAnswers((prev) => ({ ...prev, [questionId]: value }));
		scheduleSave(questionId, value, bookmarks.has(questionId));
	};

	const handleToggleBookmark = (questionId: string) => {
		if (finalizing.current || closed.current) return;
		const next = new Set(bookmarks);
		const flagged = !next.delete(questionId);
		if (flagged) next.add(questionId);
		setBookmarks(next);
		scheduleSave(questionId, answers[questionId] ?? "", flagged);
	};

	const handleSubmit = async () => {
		if (finalizing.current || closed.current) return;
		finalizing.current = true;
		setActionModal(null);
		try {
			await saveToServer(true);
		} catch {
			toast.error(m["exams.ielts.take.saveFailed"]());
		} finally {
			finalizing.current = false;
		}
	};
	expire.current = handleSubmit;

	const handleExit = async () => {
		if (finalizing.current || closed.current) return;
		finalizing.current = true;
		setActionModal(null);
		try {
			await saveToServer();
			if (!closed.current)
				await navigate({ to: APP_ROUTES.EXAMS.IELTS.HISTORY, replace: true });
		} catch {
			toast.error(m["exams.ielts.take.saveFailed"]());
		} finally {
			finalizing.current = false;
		}
	};

	const currentSection =
		sections.find((s) => s.sectionNumber === activeSection) ?? sections[0];

	const adaptedPassages: IeltsPassage[] = (currentSection?.passages ?? []).map(
		(p) => ({
			id: p.id,
			title: p.title ?? "",
			paragraphs: p.content.map((c) => c.text),
		}),
	);

	const adaptedGroups: IeltsQuestionGroup[] = (
		currentSection?.questionGroups ?? []
	).map((g) => ({
		id: g.id,
		title: g.title ?? undefined,
		instruction: g.instructions ?? g.title ?? "",
		passageRefId: currentSection?.passages[0]?.id ?? "",
		questions: g.questions.map((q) => ({
			id: q.id,
			number: q.questionNumber,
			type: (q.options && q.options.length > 0
				? "matching"
				: "fill_blank") as IeltsQuestionType,
			prompt: q.stem ?? undefined,
			options: q.options?.map((opt) => `${opt.key}. ${opt.text}`),
			correctAnswer: "",
			explanation: "",
		})),
	}));

	const displayedAnswers = Object.fromEntries(
		Object.entries(answers).map(([id, value]) => {
			const option = questionById
				.get(id)
				?.options?.find((choice) => choice.key === value);
			return [id, option ? `${option.key}. ${option.text}` : value];
		}),
	);

	const paletteGroups: IeltsPaletteGroup[] = sections.map((s) => ({
		sectionNumber: s.sectionNumber,
		title: s.title || `Section ${s.sectionNumber}`,
		questions: s.questionGroups.flatMap((g) =>
			g.questions.map((q) => ({ id: q.id, questionNumber: q.questionNumber })),
		),
	}));

	const totalQuestions = paletteGroups.reduce(
		(sum, g) => sum + g.questions.length,
		0,
	);
	const answeredCount = Object.keys(answers).filter(
		(k) => answers[k]?.trim() !== "",
	).length;

	const tabSections = sections.map((s) => {
		const sQuestions = s.questionGroups.flatMap((g) => g.questions);
		const answered = sQuestions.filter(
			(q) => answers[q.id] && answers[q.id].trim() !== "",
		).length;
		return {
			sectionNumber: s.sectionNumber,
			label: m["exams.ielts.take.section"]({ number: s.sectionNumber }),
			totalQuestions: sQuestions.length,
			answeredCount: answered,
		};
	});

	return (
		<div className="flex h-screen w-screen flex-col overflow-hidden bg-background">
			<IeltsHeader
				testTitle={attempt.examTest?.title ?? "IELTS Test"}
				timeLeft={timeLeft}
				highlightMode={highlightMode}
				onToggleHighlight={() => setHighlightMode((prev) => !prev)}
				onSubmit={() => setActionModal("submit")}
				onExit={() => setActionModal("exit")}
				onReset={() => {}}
			/>

			<div className="flex flex-1 overflow-hidden">
				<div className="flex-1 overflow-hidden border-r border-border/80">
					<IeltsPassagePane
						passages={adaptedPassages}
						highlightMode={highlightMode}
						onClearHighlights={() => {}}
					/>
				</div>

				<div className="flex-1 overflow-hidden">
					<IeltsQuestionPane
						groups={adaptedGroups}
						answers={displayedAnswers}
						bookmarks={bookmarks}
						activeQuestionId={activeQuestionId}
						onAnswerChange={handleAnswerChange}
						onToggleBookmark={handleToggleBookmark}
					/>
				</div>
			</div>

			<div className="shrink-0 border-t border-border/80 bg-background/95 backdrop-blur">
				<IeltsTabBar
					sections={tabSections}
					activeSection={activeSection}
					onSelectSection={setActiveSection}
				/>
				<IeltsQuestionPalette
					groups={paletteGroups}
					answers={answers}
					bookmarks={bookmarks}
					activeQuestionId={activeQuestionId}
					isCollapsed={isPaletteCollapsed}
					onToggleCollapse={() => setIsPaletteCollapsed((prev) => !prev)}
					onSelectQuestion={(qId, secNum) => {
						setActiveQuestionId(qId);
						if (secNum !== activeSection) {
							setActiveSection(secNum);
						}
					}}
				/>
			</div>

			{actionModal && (
				<IeltsConfirmModal
					action={actionModal}
					answeredCount={answeredCount}
					totalCount={totalQuestions}
					onConfirm={
						actionModal === "submit"
							? () => void handleSubmit()
							: () => void handleExit()
					}
					onClose={() => setActionModal(null)}
				/>
			)}
		</div>
	);
}
