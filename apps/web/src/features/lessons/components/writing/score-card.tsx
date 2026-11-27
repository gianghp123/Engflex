import type {
	Span,
	TaskVerdict,
	WritingScoreResponse,
} from "@engflex/contracts";
import { cn } from "cn";
import { Copy } from "lucide-react";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";

/** Post-score panel for writing feedback v2: summary strip, task verdicts,
 *  grammar/phrasing rows, expressions with copy, suggested rewrite, one tip.
 *  Spans render as a list (no invented offsets — the API returns text, not
 *  positions). The learner paragraph is never repeated here. Status color
 *  follows the house tone language (`accuracy` good, `warning` partial,
 *  `ai-coral` needs work); each correction strikes the original in coral
 *  and sets the fix in accuracy, mirroring the voice transcript card. */
export function ScoreCard({
	score,
	continueLabel,
	onContinue,
}: {
	score: WritingScoreResponse;
	continueLabel: string;
	onContinue: () => void;
}) {
	const verdictMarker = (status: string) =>
		status === "covered"
			? "✅"
			: status === "partial"
				? "⚠️"
				: status === "missing"
					? "❌"
					: "";
	const verdictLabel = (status: string) =>
		status === "covered"
			? m["lessons.writing.score.covered"]()
			: status === "partial"
				? m["lessons.writing.score.partial"]()
				: status === "missing"
					? m["lessons.writing.score.missing"]()
					: status;
	const verdictTone = (status: string) =>
		status === "covered"
			? "text-accuracy"
			: status === "partial"
				? "text-warning"
				: status === "missing"
					? "text-ai-coral"
					: "text-muted-foreground";
	return (
		<div className="surface-card flex flex-col gap-4 p-5">
			<div className="flex flex-wrap items-center justify-between gap-2">
				<span className="text-[11px] font-bold text-muted-foreground">
					{m["lessons.writing.score.title"]()}
				</span>
				<span className="text-xs font-bold text-foreground">
					{m["lessons.writing.score.points"]({
						covered: score.summary.pointsCovered,
						total: score.summary.pointsTotal,
					})}
					{" · "}
					{m["lessons.writing.score.words"]({
						count: score.summary.wordCount,
						min: score.summary.minWords,
						max: score.summary.maxWords,
					})}
				</span>
			</div>
			<div className="flex flex-col gap-2 rounded-[16px] border-2 border-border bg-accent p-3">
				<span className="text-xs font-bold text-muted-foreground">
					{m["lessons.writing.score.taskTitle"]()}
				</span>
				<ul className="flex flex-col gap-1.5">
					{score.task.verdicts.map((verdict) => (
						<TaskVerdictRow
							key={verdict.item}
							verdict={verdict}
							marker={verdictMarker(verdict.status)}
							label={verdictLabel(verdict.status)}
							tone={verdictTone(verdict.status)}
						/>
					))}
				</ul>
				<p className="text-[15px] font-medium text-foreground">
					<span aria-hidden="true">
						{score.task.answersPrompt ? "✅ " : "❌ "}
					</span>
					<span className="text-xs font-bold text-muted-foreground">
						{m["lessons.writing.score.answersPrompt"]()}:{" "}
					</span>
					{score.task.answersNote}
				</p>
			</div>
			<SpanSection
				title={m["lessons.writing.score.grammarTitle"]()}
				spans={score.grammar}
				empty={m["lessons.writing.score.grammarEmpty"]()}
			/>
			<SpanSection
				title={m["lessons.writing.score.phrasingTitle"]()}
				spans={score.phrasing}
				empty={m["lessons.writing.score.phrasingEmpty"]()}
			/>
			{score.expressions.length > 0 ? (
				<div className="flex flex-col gap-2">
					<span className="text-xs font-bold text-accent-violet">
						{m["lessons.writing.score.expressionsTitle"]()}
					</span>
					<ul className="flex flex-col gap-1.5">
						{score.expressions.map((phrase) => (
							<li
								key={phrase}
								className="flex items-center justify-between gap-2 text-[15px] font-medium text-foreground"
							>
								<span>{phrase}</span>
								<Button
									type="button"
									variant="ghost"
									size="sm"
									className="btn btn-outline shrink-0 px-2"
									onClick={() => {
										void navigator.clipboard?.writeText(phrase).catch(() => {});
									}}
								>
									<Copy data-icon="inline-start" />
									{m["common.actions.copy"]()}
								</Button>
							</li>
						))}
					</ul>
				</div>
			) : null}
			<div className="flex flex-col gap-2">
				<span className="text-xs font-bold text-accent-violet">
					{m["lessons.writing.score.suggestedTitle"]()}
				</span>
				<p className="text-[15px] leading-[24px] font-medium text-foreground">
					{score.suggested}
				</p>
			</div>
			<div className="flex flex-col gap-2 rounded-[16px] border-2 border-border bg-accent p-3">
				<span className="text-xs font-bold text-accent-violet">
					{m["lessons.writing.score.tipTitle"]()}
				</span>
				<p className="text-sm text-muted-foreground">{score.tip}</p>
			</div>
			<div className="flex justify-end">
				<Button type="button" className="btn btn-primary" onClick={onContinue}>
					{continueLabel}
				</Button>
			</div>
		</div>
	);
}

function TaskVerdictRow({
	verdict,
	marker,
	label,
	tone,
}: {
	verdict: TaskVerdict;
	marker: string;
	label: string;
	tone: string;
}) {
	return (
		<li className="flex items-start gap-2 text-[15px] font-medium text-foreground">
			<span className={cn("shrink-0 text-xs font-bold", tone)}>
				{marker ? <span aria-hidden="true">{marker} </span> : null}
				<span>{label}</span>
			</span>
			<span className="flex flex-col gap-0.5">
				<span>{verdict.item}</span>
				{verdict.evidence ? (
					<span className="text-xs font-normal text-muted-foreground">
						{verdict.evidence}
					</span>
				) : null}
			</span>
		</li>
	);
}

function SpanSection({
	title,
	spans,
	empty,
}: {
	title: string;
	spans: Span[];
	empty: string;
}) {
	return (
		<div className="flex flex-col gap-2">
			<span className="text-xs font-bold text-accent-violet">{title}</span>
			{spans.length > 0 ? (
				<ul className="flex flex-col gap-1.5">
					{spans.map((span) => (
						<li
							key={`${span.text}-${span.occurrence}`}
							className="flex flex-col gap-0.5 text-[15px] font-medium text-foreground"
						>
							{span.errorType ? (
								<Badge variant="outline" className="self-start">
									{span.errorType}
								</Badge>
							) : null}
							<span>
								<span className="text-ai-coral line-through opacity-85">
									{span.text}
								</span>{" "}
								→{" "}
								<span className="font-semibold text-accuracy">
									{span.correction}
								</span>
							</span>
							<span className="text-xs font-normal text-muted-foreground">
								{span.reason}
							</span>
						</li>
					))}
				</ul>
			) : (
				<p className="text-sm font-medium text-accuracy">{empty}</p>
			)}
		</div>
	);
}
