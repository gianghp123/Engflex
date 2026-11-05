import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { IeltsConfirmModal } from "#/features/exams/components/ielts/ielts-confirm-modal";
import { IeltsHeader } from "#/features/exams/components/ielts/ielts-header";
import { IeltsPassagePane } from "#/features/exams/components/ielts/ielts-passage-pane";
import { IeltsQuestionPalette } from "#/features/exams/components/ielts/ielts-question-palette";
import { IeltsQuestionPane } from "#/features/exams/components/ielts/ielts-question-pane";
import { IeltsResultModal } from "#/features/exams/components/ielts/ielts-result-modal";
import { IeltsTabBar } from "#/features/exams/components/ielts/ielts-tab-bar";
import { GT_IELTS_TEST_1 } from "#/features/exams/ielts/mock-data";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/exams/ielts/")({
	component: IeltsExamPage,
});

const STORAGE_ANSWERS = "engflex_ielts_answers";
const STORAGE_BOOKMARKS = "engflex_ielts_bookmarks";
const STORAGE_TIME = "engflex_ielts_time_left";

function IeltsExamPage() {
	const navigate = useNavigate();
	const test = GT_IELTS_TEST_1;

	// Initial State from LocalStorage
	const [answers, setAnswers] = useState<Record<string, string>>(() => {
		if (typeof window === "undefined") return {};
		try {
			const saved = localStorage.getItem(STORAGE_ANSWERS);
			return saved ? JSON.parse(saved) : {};
		} catch {
			return {};
		}
	});

	const [bookmarks, setBookmarks] = useState<Set<string>>(() => {
		if (typeof window === "undefined") return new Set();
		try {
			const saved = localStorage.getItem(STORAGE_BOOKMARKS);
			return saved ? new Set(JSON.parse(saved)) : new Set();
		} catch {
			return new Set();
		}
	});

	const [timeLeft, setTimeLeft] = useState<number>(() => {
		if (typeof window === "undefined") return test.durationSeconds;
		try {
			const saved = localStorage.getItem(STORAGE_TIME);
			return saved ? Number(saved) : test.durationSeconds;
		} catch {
			return test.durationSeconds;
		}
	});

	const [activeSection, setActiveSection] = useState<number>(1);
	const [activeQuestionId, setActiveQuestionId] = useState<string | null>(null);
	const [highlightMode, setHighlightMode] = useState<boolean>(false);
	const [isPaletteCollapsed, setIsPaletteCollapsed] = useState<boolean>(false);
	const [actionModal, setActionModal] = useState<
		"submit" | "exit" | "reset" | null
	>(null);
	const [resultModalOpen, setResultModalOpen] = useState<boolean>(false);

	// Persist answers
	useEffect(() => {
		try {
			localStorage.setItem(STORAGE_ANSWERS, JSON.stringify(answers));
		} catch {
			// Ignore localStorage write errors
		}
	}, [answers]);

	// Persist bookmarks
	useEffect(() => {
		try {
			localStorage.setItem(
				STORAGE_BOOKMARKS,
				JSON.stringify(Array.from(bookmarks)),
			);
		} catch {
			// Ignore localStorage write errors
		}
	}, [bookmarks]);

	// Timer countdown & persistence
	useEffect(() => {
		if (resultModalOpen) return;

		const timer = setInterval(() => {
			setTimeLeft((prev) => {
				if (prev <= 1) {
					clearInterval(timer);
					setResultModalOpen(true);
					return 0;
				}
				const next = prev - 1;
				try {
					localStorage.setItem(STORAGE_TIME, String(next));
				} catch {
					// Ignore
				}
				return next;
			});
		}, 1000);

		return () => clearInterval(timer);
	}, [resultModalOpen]);

	// Handler: Answer input
	const handleAnswerChange = (questionId: string, value: string) => {
		setAnswers((prev) => ({
			...prev,
			[questionId]: value,
		}));
	};

	// Handler: Bookmark toggle
	const handleToggleBookmark = (questionId: string) => {
		setBookmarks((prev) => {
			const next = new Set(prev);
			if (next.has(questionId)) {
				next.delete(questionId);
			} else {
				next.add(questionId);
			}
			return next;
		});
	};

	// Handler: Palette question selection
	const handleSelectQuestion = (
		questionId: string,
		targetSectionNumber: number,
	) => {
		if (activeSection !== targetSectionNumber) {
			setActiveSection(targetSectionNumber);
		}
		setActiveQuestionId(questionId);

		setTimeout(() => {
			document.getElementById(questionId)?.scrollIntoView({
				behavior: "smooth",
				block: "center",
			});
		}, 100);
	};

	// Handler: Reset test
	const handleReset = () => {
		setAnswers({});
		setBookmarks(new Set());
		setTimeLeft(test.durationSeconds);
		setActiveSection(1);
		setActiveQuestionId(null);
		setResultModalOpen(false);
		try {
			localStorage.removeItem(STORAGE_ANSWERS);
			localStorage.removeItem(STORAGE_BOOKMARKS);
			localStorage.removeItem(STORAGE_TIME);
		} catch {
			// Ignore
		}
		toast.success(m["exams.ielts.take.saved"]());
	};

	// Handler: Exit test
	const handleExit = () => {
		void navigate({ to: APP_ROUTES.HOME });
	};

	// Current Section
	const currentSection =
		test.sections.find((s) => s.sectionNumber === activeSection) ??
		test.sections[0];

	// Tab bar items meta
	const tabSectionsMeta = test.sections.map((sec) => {
		const secQuestions = sec.groups.flatMap((g) => g.questions);
		const answered = secQuestions.filter(
			(q) => answers[q.id] != null && answers[q.id].trim() !== "",
		).length;
		return {
			sectionNumber: sec.sectionNumber,
			label: m["exams.ielts.take.section"]({ number: sec.sectionNumber }),
			totalQuestions: secQuestions.length,
			answeredCount: answered,
		};
	});

	// Palette groups meta
	const paletteGroups = test.sections.map((sec) => ({
		sectionNumber: sec.sectionNumber,
		title: m["exams.ielts.take.section"]({ number: sec.sectionNumber }),
		questions: sec.groups.flatMap((g) =>
			g.questions.map((q) => ({
				id: q.id,
				questionNumber: q.number,
			})),
		),
	}));

	const answeredTotal = Object.values(answers).filter(
		(v) => v != null && v.trim() !== "",
	).length;

	return (
		<div className="fixed inset-0 z-50 flex flex-col overflow-hidden bg-background text-foreground">
			{/* Top Header */}
			<IeltsHeader
				testTitle={test.title}
				timeLeft={timeLeft}
				highlightMode={highlightMode}
				onToggleHighlight={() => setHighlightMode((prev) => !prev)}
				onReset={() => setActionModal("reset")}
				onExit={() => setActionModal("exit")}
				onSubmit={() => setActionModal("submit")}
			/>

			{/* Section Tab Bar */}
			<IeltsTabBar
				sections={tabSectionsMeta}
				activeSection={activeSection}
				onSelectSection={(secNum) => {
					setActiveSection(secNum);
					setActiveQuestionId(null);
				}}
			/>

			{/* Status Bar */}
			<div
				className="flex items-center justify-between gap-2 border-b border-border/70 px-4 py-1.5 text-xs text-muted-foreground bg-muted/10 select-none"
				aria-live="polite"
			>
				<span className="flex items-center gap-1.5 font-medium">
					<span className="inline-block size-2 rounded-full bg-emerald-500 animate-pulse" />
					{m["exams.ielts.take.saved"]()}
				</span>
				<span>
					{m["exams.ielts.take.section"]({ number: activeSection })} •{" "}
					{currentSection.passages.map((p) => p.title).join(" & ")}
				</span>
			</div>

			{/* Main Split Layout */}
			<div className="flex min-h-0 flex-1 overflow-hidden">
				<main className="grid grid-cols-1 lg:grid-cols-2 min-w-0 flex-1 overflow-hidden">
					{/* Left: Reading Passage */}
					<IeltsPassagePane
						key={`passage-${activeSection}`}
						passages={currentSection.passages}
						highlightMode={highlightMode}
						onClearHighlights={() => {
							const marks = document.querySelectorAll("mark");
							for (const m of marks) {
								const text = document.createTextNode(m.textContent ?? "");
								m.parentNode?.replaceChild(text, m);
							}
						}}
					/>

					{/* Right: Questions Pane */}
					<IeltsQuestionPane
						key={`questions-${activeSection}`}
						groups={currentSection.groups}
						answers={answers}
						bookmarks={bookmarks}
						activeQuestionId={activeQuestionId}
						onAnswerChange={handleAnswerChange}
						onToggleBookmark={handleToggleBookmark}
					/>
				</main>

				{/* Right Sidebar: Palette */}
				<aside className="shrink-0">
					<IeltsQuestionPalette
						isCollapsed={isPaletteCollapsed}
						onToggleCollapse={() => setIsPaletteCollapsed((prev) => !prev)}
						groups={paletteGroups}
						activeQuestionId={activeQuestionId}
						answers={answers}
						bookmarks={bookmarks}
						onSelectQuestion={handleSelectQuestion}
					/>
				</aside>
			</div>

			{/* Confirm Action Modal */}
			<IeltsConfirmModal
				action={actionModal}
				answeredCount={answeredTotal}
				totalCount={test.totalQuestions}
				onClose={() => setActionModal(null)}
				onConfirm={() => {
					if (actionModal === "submit") {
						setResultModalOpen(true);
					} else if (actionModal === "exit") {
						handleExit();
					} else if (actionModal === "reset") {
						handleReset();
					}
				}}
			/>

			{/* Result Review Modal */}
			<IeltsResultModal
				open={resultModalOpen}
				test={test}
				answers={answers}
				timeSpentSeconds={test.durationSeconds - timeLeft}
				onClose={() => setResultModalOpen(false)}
				onRetake={handleReset}
			/>
		</div>
	);
}
