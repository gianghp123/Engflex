import type { ExamReviewAttempt } from "@engflex/contracts";
import { CardShell } from "#/components/common/card-shell";
import { ProgressRing, StatCard } from "#/components/common/stat-card";
import { formatExamDuration } from "#/features/exams/dates";
import { type AnswersById, partStatistics } from "#/features/exams/review";
import { m } from "#/paraglide/messages";

export function ToeicResultHeaderCard({
	attempt,
	answers,
}: {
	attempt: ExamReviewAttempt;
	answers: AnswersById;
}) {
	const parts = attempt.examTest?.content?.parts ?? [];
	const result = attempt.result;
	const total = parts.reduce(
		(sum, part) => sum + partStatistics(part, answers).totalQuestions,
		0,
	);
	const sections = (["listening", "reading"] as const).map((section) => {
		const scores = parts
			.filter((part) => part.section === section)
			.map((part) => partStatistics(part, answers));
		return {
			section,
			scaledScore: result?.sections?.[section]?.scaledScore,
			correct: scores.reduce((sum, part) => sum + part.correctCount, 0),
			total: scores.reduce((sum, part) => sum + part.totalQuestions, 0),
		};
	});
	return (
		<div className="flex flex-col gap-4">
			<h2 className="font-bold">{attempt.examTest?.title}</h2>
			<div className="grid gap-4 md:grid-cols-3">
				<StatCard
					title={m["exams.toeic.result.accuracy"]()}
					value={
						result?.accuracyPct != null
							? m["exams.toeic.result.percent"]({
									value: Math.round(result.accuracyPct),
								})
							: m["exams.attempt.unavailable"]()
					}
					aside={
						result?.accuracyPct != null ? (
							<ProgressRing value={result.accuracyPct} />
						) : undefined
					}
				>
					<p className="text-xs text-muted-foreground">
						{result?.correctCount != null
							? m["exams.toeic.result.correctCount"]({
									correct: result.correctCount,
									total,
								})
							: m["exams.attempt.unavailable"]()}
					</p>
					<p className="text-xs text-muted-foreground">
						{m["exams.toeic.result.totalScore"]()}:{" "}
						{attempt.score != null
							? m["exams.toeic.result.estimatedScore"]({
									score: attempt.score,
									max: 990,
								})
							: m["exams.toeic.result.scaledUnavailable"]()}
					</p>
				</StatCard>
				{sections.map((section) => (
					<StatCard
						key={section.section}
						title={
							section.section === "listening"
								? m["exams.toeic.test.breakdown.listening"]()
								: m["exams.toeic.test.breakdown.reading"]()
						}
						value={
							section.total
								? m["exams.toeic.result.fraction"]({
										correct: section.correct,
										total: section.total,
									})
								: m["exams.toeic.result.notTaken"]()
						}
					>
						<p className="text-xs text-muted-foreground">
							{section.scaledScore != null
								? m["exams.toeic.result.estimatedScore"]({
										score: section.scaledScore,
										max: 495,
									})
								: m["exams.toeic.result.scaledUnavailable"]()}
						</p>
					</StatCard>
				))}
			</div>
			<CardShell className="p-4 text-sm text-muted-foreground">
				{m["exams.toeic.result.timeSpent"]({
					duration: formatExamDuration(attempt.durationSec),
				})}
			</CardShell>
		</div>
	);
}
