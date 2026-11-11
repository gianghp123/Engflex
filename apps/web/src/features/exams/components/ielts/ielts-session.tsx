import type { ExamAttempt } from "@engflex/contracts";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import type {
	IeltsPassage,
	IeltsQuestionGroup,
	IeltsQuestionType,
} from "#/features/exams/ielts/mock-data";
import {
	useCancelExamAttempt,
	usePatchExamAttempt,
} from "#/features/exams/queries";
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
	const cancel = useCancelExamAttempt();

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

	const revisionRef = useRef<number>(attempt.ieltsResult?.revision ?? 1);
	const answersRef = useRef(answers);
	answersRef.current = answers;
	const bookmarksRef = useRef(bookmarks);
	bookmarksRef.current = bookmarks;
	const handleSubmitRef = useRef<() => Promise<void>>(() => Promise.resolve());

	useEffect(() => {
		if (attempt.mode !== "exam" || !attempt.deadlineAt) return;
		const deadline = new Date(attempt.deadlineAt).getTime();
		const timer = setInterval(() => {
			const remaining = Math.max(0, Math.ceil((deadline - Date.now()) / 1000));
			setTimeLeft(remaining);
			if (remaining === 0) {
				clearInterval(timer);
				void handleSubmitRef.current();
			}
		}, 1000);
		return () => clearInterval(timer);
	}, [attempt.deadlineAt, attempt.mode]);

	const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(
		undefined,
	);
	const saveQueue = useRef<Promise<unknown>>(Promise.resolve());

	useEffect(() => () => clearTimeout(debounceTimerRef.current), []);

	const saveToServer = (submit = false) => {
		clearTimeout(debounceTimerRef.current);
		const currentAnswers = answersRef.current;
		const currentBookmarks = bookmarksRef.current;
		const answerPayload = sections.flatMap((section) =>
			section.questionGroups.flatMap((group) =>
				group.questions.map((question) => {
					const value = currentAnswers[question.id] || null;
					const hasOptions = Boolean(question.options?.length);
					return {
						questionId: question.id,
						selectedKey: hasOptions ? value : null,
						textValue: hasOptions ? null : value,
						flagged: currentBookmarks.has(question.id),
					};
				}),
			),
		);

		const pending = saveQueue.current
			.catch(() => {})
			.then(async () => {
				const updated = await patch.mutateAsync({
					attemptId: attempt.id,
					req: {
						revision: revisionRef.current,
						submit,
						answers: answerPayload,
					},
				});
				revisionRef.current =
					updated.ieltsResult?.revision ?? revisionRef.current;
			});
		saveQueue.current = pending;
		return pending;
	};

	const scheduleSave = () => {
		clearTimeout(debounceTimerRef.current);
		debounceTimerRef.current = setTimeout(() => {
			void saveToServer()
				.then(() => toast.success(m["exams.ielts.take.saved"]()))
				.catch(() => toast.error(m["exams.ielts.take.saveFailed"]()));
		}, 2000);
	};

	const handleAnswerChange = (questionId: string, answer: string) => {
		setAnswers((prev) => ({ ...prev, [questionId]: answer }));
		scheduleSave();
	};

	const handleToggleBookmark = (questionId: string) => {
		setBookmarks((prev) => {
			const next = new Set(prev);
			if (next.has(questionId)) next.delete(questionId);
			else next.add(questionId);
			return next;
		});
		scheduleSave();
	};

	const handleSubmit = async () => {
		setActionModal(null);
		try {
			await saveToServer(true);

			void navigate({
				to: APP_ROUTES.EXAMS.IELTS.RESULT,
				params: { id: attempt.id },
				replace: true,
			});
		} catch {
			toast.error(m["exams.ielts.take.saveFailed"]());
		}
	};
	handleSubmitRef.current = handleSubmit;

	const handleExit = async () => {
		setActionModal(null);
		try {
			await saveToServer();
			await cancel.mutateAsync(attempt.id);
			void navigate({
				to: APP_ROUTES.EXAMS.IELTS.HISTORY,
				replace: true,
			});
		} catch {
			toast.error(m["exams.ielts.take.saveFailed"]());
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
			optionKeys: q.options?.map((opt) => opt.key),
			correctAnswer: "",
			explanation: "",
		})),
	}));

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
						answers={answers}
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
						actionModal === "submit" ? () => void handleSubmit() : handleExit
					}
					onClose={() => setActionModal(null)}
				/>
			)}
		</div>
	);
}
