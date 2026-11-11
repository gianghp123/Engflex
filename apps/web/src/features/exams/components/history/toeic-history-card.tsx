import type { AttemptStatus, ExamAttemptSummary } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { Clock } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { SubmitButton } from "#/components/common/submit-button";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { formatExamDate, formatExamDuration } from "#/features/exams/dates";
import { m } from "#/paraglide/messages";

const STATUS_LABELS: Record<AttemptStatus, () => string> = {
	in_progress: () => m["exams.toeic.history.status.inProgress"](),
	completed: () => m["exams.toeic.history.status.completed"](),
	abandoned: () => m["exams.toeic.history.status.abandoned"](),
};
export function ToeicHistoryCard({
	item,
	onCancel,
	onContinue,
	pending,
	locked,
}: {
	item: ExamAttemptSummary;
	onCancel: (item: ExamAttemptSummary) => void;
	onContinue: (id: string) => void;
	pending: boolean;
	locked: boolean;
}) {
	const result = item.result;
	return (
		<div className="flex flex-col gap-4 rounded-2xl border border-border/80 bg-card p-5 shadow-xs sm:flex-row sm:items-center sm:justify-between">
			<div className="flex min-w-0 flex-col gap-2">
				<div className="flex flex-wrap items-center gap-2">
					<h3 className="truncate text-base font-bold">
						{item.examTest?.title ?? m["exams.toeic.history.missingTest"]()}
					</h3>
					<Badge variant="outline">
						{item.mode === null
							? m["exams.attempt.unavailable"]()
							: item.mode === "exam"
								? m["exams.toeic.hub.history.modeExam"]()
								: m["exams.toeic.hub.history.modePractice"]()}
					</Badge>
					<Badge variant="secondary">{STATUS_LABELS[item.status]()}</Badge>
					{item.mode === "practice" && (
						<span className="text-xs text-muted-foreground">
							{m["exams.toeic.history.selectedParts"]({
								parts: item.selectedParts.join(", "),
							})}
						</span>
					)}
				</div>
				<div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
					{item.status === "completed" && (
						<span className="font-bold text-foreground">
							{result?.correctCount != null && result.accuracyPct != null
								? m["exams.toeic.history.result"]({
										correct: result.correctCount,
										total: result.questionCount,
										accuracy: Math.round(result.accuracyPct),
									})
								: m["exams.attempt.unavailable"]()}
						</span>
					)}
					<span className="inline-flex items-center gap-1">
						<Clock className="size-3.5" aria-hidden="true" />
						{formatExamDuration(item.durationSec)}
					</span>
					<span>{formatExamDate(item.startedAt ?? item.createdAt)}</span>
				</div>
			</div>
			<div className="flex shrink-0 items-center gap-2 self-end sm:self-center">
				{item.status === "completed" && (
					<Button asChild variant="outline">
						<Link to={APP_ROUTES.EXAMS.TOEIC.RESULT} params={{ id: item.id }}>
							{m["exams.toeic.hub.history.viewResult"]()}
						</Link>
					</Button>
				)}
				{item.status === "in_progress" && (
					<>
						<Button
							type="button"
							variant="destructive"
							disabled={locked}
							onClick={() => {
								if (!locked) onCancel(item);
							}}
						>
							{m["exams.toeic.hub.history.cancel"]()}
						</Button>
						<SubmitButton
							pending={pending}
							disabled={locked && !pending}
							onClick={() => onContinue(item.id)}
						>
							{m["exams.toeic.hub.history.continue"]()}
						</SubmitButton>
					</>
				)}
			</div>
		</div>
	);
}
