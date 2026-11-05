import { Award, CheckCircle2, RotateCcw, XCircle } from "lucide-react";
import { useState } from "react";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import {
	calculateIeltsBandScore,
	type IeltsExamTest,
	isAnswerCorrect,
} from "#/features/exams/ielts/mock-data";
import { m } from "#/paraglide/messages";

function formatDuration(totalSeconds: number): string {
	const mm = Math.floor(totalSeconds / 60);
	const ss = totalSeconds % 60;
	return `${mm}m ${ss}s`;
}

function formatCorrectAnswer(answer: string | string[]): string {
	if (Array.isArray(answer)) {
		return answer.join(" / ");
	}
	return answer;
}

export function IeltsResultModal({
	open,
	test,
	answers,
	timeSpentSeconds,
	onClose,
	onRetake,
}: {
	open: boolean;
	test: IeltsExamTest;
	answers: Record<string, string>;
	timeSpentSeconds: number;
	onClose: () => void;
	onRetake: () => void;
}) {
	const [filter, setFilter] = useState<"all" | "correct" | "incorrect">("all");

	// Gather all 40 questions in order
	const allQuestions = test.sections.flatMap((s) =>
		s.groups.flatMap((g) => g.questions),
	);

	const results = allQuestions.map((q) => {
		const userAns = answers[q.id] ?? "";
		const correct = isAnswerCorrect(userAns, q.correctAnswer);
		return {
			question: q,
			userAnswer: userAns,
			isCorrect: correct,
		};
	});

	const correctCount = results.filter((r) => r.isCorrect).length;
	const totalCount = allQuestions.length;
	const bandScore = calculateIeltsBandScore(correctCount);
	const accuracy = Math.round((correctCount / totalCount) * 100);

	const filteredResults = results.filter((r) => {
		if (filter === "correct") return r.isCorrect;
		if (filter === "incorrect") return !r.isCorrect;
		return true;
	});

	return (
		<Dialog open={open} onOpenChange={(val) => !val && onClose()}>
			<DialogContent className="max-w-3xl max-h-[90vh] flex flex-col p-0 gap-0 overflow-hidden">
				<DialogHeader className="p-6 pb-4 border-b border-border/80 bg-muted/20 shrink-0">
					<div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
						<div className="space-y-1">
							<DialogTitle className="text-xl font-bold flex items-center gap-2">
								<Award className="size-5 text-amber-500" />
								<span>{m["exams.ielts.result.title"]()}</span>
							</DialogTitle>
							<DialogDescription className="text-xs">
								{test.title} • {formatDuration(timeSpentSeconds)}
							</DialogDescription>
						</div>

						{/* Band Score Pill */}
						<div className="flex items-center gap-3">
							<div className="flex flex-col items-center justify-center rounded-xl bg-gradient-to-br from-amber-500 to-orange-600 px-4 py-2 text-white shadow-md">
								<span className="text-[10px] font-bold uppercase tracking-wider opacity-90">
									{m["exams.ielts.result.bandScore"]()}
								</span>
								<span className="text-2xl font-black">
									{bandScore.toFixed(1)}
								</span>
							</div>
						</div>
					</div>

					{/* Summary Stats Grid */}
					<div className="grid grid-cols-3 gap-3 mt-4 pt-3 border-t border-border/60">
						<div className="rounded-lg border border-border/70 bg-card p-2.5 text-center">
							<div className="text-xs text-muted-foreground font-medium">
								{m["exams.ielts.result.correctCount"]({
									correct: correctCount,
									total: totalCount,
								})}
							</div>
							<div className="text-lg font-bold text-emerald-600">
								{correctCount} / {totalCount}
							</div>
						</div>

						<div className="rounded-lg border border-border/70 bg-card p-2.5 text-center">
							<div className="text-xs text-muted-foreground font-medium">
								{m["exams.ielts.result.accuracy"]({ percent: accuracy })}
							</div>
							<div className="text-lg font-bold text-sky-600">{accuracy}%</div>
						</div>

						<div className="rounded-lg border border-border/70 bg-card p-2.5 text-center">
							<div className="text-xs text-muted-foreground font-medium">
								{m["exams.ielts.result.timeSpent"]({
									duration: formatDuration(timeSpentSeconds),
								})}
							</div>
							<div className="text-lg font-bold text-foreground">
								{formatDuration(timeSpentSeconds)}
							</div>
						</div>
					</div>
				</DialogHeader>

				{/* Filter Tabs */}
				<div className="flex items-center gap-2 px-6 py-2.5 border-b border-border/70 bg-background shrink-0">
					<button
						type="button"
						onClick={() => setFilter("all")}
						className={`rounded-md px-3 py-1 text-xs font-semibold transition-all cursor-pointer ${
							filter === "all"
								? "bg-foreground text-background"
								: "text-muted-foreground hover:bg-muted hover:text-foreground"
						}`}
					>
						{m["exams.ielts.result.filterAll"]()} ({totalCount})
					</button>

					<button
						type="button"
						onClick={() => setFilter("correct")}
						className={`rounded-md px-3 py-1 text-xs font-semibold transition-all cursor-pointer ${
							filter === "correct"
								? "bg-emerald-600 text-white"
								: "text-muted-foreground hover:bg-muted hover:text-foreground"
						}`}
					>
						{m["exams.ielts.result.filterCorrect"]()} ({correctCount})
					</button>

					<button
						type="button"
						onClick={() => setFilter("incorrect")}
						className={`rounded-md px-3 py-1 text-xs font-semibold transition-all cursor-pointer ${
							filter === "incorrect"
								? "bg-rose-600 text-white"
								: "text-muted-foreground hover:bg-muted hover:text-foreground"
						}`}
					>
						{m["exams.ielts.result.filterIncorrect"]()} (
						{totalCount - correctCount})
					</button>
				</div>

				{/* Questions Review List */}
				<div className="flex-1 overflow-y-auto p-6 space-y-3.5 select-text">
					<div className="flex items-center justify-between pb-1">
						<h3 className="text-sm font-bold text-foreground">
							{m["exams.ielts.result.reviewTitle"]()}
						</h3>
					</div>
					{filteredResults.map(({ question, userAnswer, isCorrect }) => (
						<div
							key={question.id}
							className={`rounded-lg border p-4 space-y-2 transition-all ${
								isCorrect
									? "border-emerald-200 bg-emerald-50/20 dark:border-emerald-900/40 dark:bg-emerald-950/10"
									: "border-rose-200 bg-rose-50/20 dark:border-rose-900/40 dark:bg-rose-950/10"
							}`}
						>
							<div className="flex items-start justify-between gap-3">
								<div className="flex items-center gap-2">
									<span className="flex size-6 shrink-0 items-center justify-center rounded-md bg-muted font-mono text-xs font-bold text-foreground">
										{question.number}
									</span>
									{isCorrect ? (
										<span className="inline-flex items-center gap-1 text-xs font-bold text-emerald-600 dark:text-emerald-400">
											<CheckCircle2 className="size-4" />
											<span>{m["exams.ielts.result.filterCorrect"]()}</span>
										</span>
									) : (
										<span className="inline-flex items-center gap-1 text-xs font-bold text-rose-600 dark:text-rose-400">
											<XCircle className="size-4" />
											<span>{m["exams.ielts.result.filterIncorrect"]()}</span>
										</span>
									)}
								</div>
							</div>

							{question.prompt && (
								<p className="text-xs md:text-sm text-foreground font-medium">
									{question.prompt}
								</p>
							)}

							<div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs pt-1">
								<div className="rounded-md border border-border/70 bg-background/80 p-2">
									<span className="text-muted-foreground block text-[11px] mb-0.5">
										{m["exams.ielts.result.yourAnswer"]({ answer: "" })}
									</span>
									<span
										className={`font-semibold ${
											isCorrect ? "text-emerald-600" : "text-rose-600"
										}`}
									>
										{userAnswer || "(Chưa làm)"}
									</span>
								</div>

								<div className="rounded-md border border-border/70 bg-background/80 p-2">
									<span className="text-muted-foreground block text-[11px] mb-0.5">
										{m["exams.ielts.result.correctAnswer"]({ answer: "" })}
									</span>
									<span className="font-semibold text-emerald-700 dark:text-emerald-400">
										{formatCorrectAnswer(question.correctAnswer)}
									</span>
								</div>
							</div>

							{question.explanation && (
								<div className="text-xs text-muted-foreground pt-1 border-t border-border/50">
									<span className="font-semibold text-foreground">
										{m["exams.ielts.result.explanation"]()}{" "}
									</span>
									<span>{question.explanation}</span>
								</div>
							)}
						</div>
					))}
				</div>

				{/* Footer */}
				<div className="flex items-center justify-between border-t border-border/80 p-4 bg-muted/20 shrink-0">
					<Button
						variant="outline"
						size="sm"
						onClick={onRetake}
						className="gap-1.5 text-xs font-semibold"
					>
						<RotateCcw className="size-3.5" />
						<span>{m["exams.ielts.result.retakeButton"]()}</span>
					</Button>

					<Button size="sm" onClick={onClose} className="text-xs font-semibold">
						{m["exams.ielts.result.closeButton"]()}
					</Button>
				</div>
			</DialogContent>
		</Dialog>
	);
}
