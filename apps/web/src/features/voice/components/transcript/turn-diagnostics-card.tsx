import type {
	AnalysisSpan,
	SpanStatus,
	TurnFeedback,
} from "@engflex/contracts";
import { cn } from "cn";
import { ArrowRight, Lightbulb, RefreshCw } from "lucide-react";
import { SubmitButton } from "#/components/common/submit-button";
import { Badge } from "#/components/ui/badge";
import { m } from "#/paraglide/messages";

const SPAN_TONE: Record<SpanStatus, string> = {
	incorrect: "text-ai-coral",
	awkward: "text-warning",
};

/** nth (1-based) occurrence of text in utterance; null when absent. */
function locateSpan(
	utterance: string,
	text: string,
	occurrence: number,
): { start: number; end: number } | null {
	let index = -1;
	for (let i = 0; i < occurrence; i++) {
		index = utterance.indexOf(text, index + 1);
		if (index < 0) {
			return null;
		}
	}
	return { start: index, end: index + text.length };
}

/** Utterance rendered with every locatable span fully highlighted. */
function HighlightedUtterance({
	utterance,
	spans,
}: {
	utterance: string;
	spans: AnalysisSpan[];
}) {
	const located = spans
		.map((span) => ({
			span,
			loc: locateSpan(utterance, span.text, span.occurrence),
		}))
		.filter(
			(
				entry,
			): entry is { span: AnalysisSpan; loc: { start: number; end: number } } =>
				entry.loc !== null,
		)
		.sort((a, b) => a.loc.start - b.loc.start);
	if (located.length === 0) {
		return <>{utterance}</>;
	}
	const parts: React.ReactNode[] = [];
	let cursor = 0;
	located.forEach(({ span, loc }) => {
		if (loc.start < cursor) {
			return;
		}
		parts.push(
			<span key={`plain-${cursor}`}>{utterance.slice(cursor, loc.start)}</span>,
		);
		parts.push(
			<span
				key={`span-${span.text}-${span.occurrence}`}
				className={cn(
					"inline-flex items-center py-0.5 font-medium",
					SPAN_TONE[span.status],
				)}
				title={`${span.correction} — ${span.reason}`}
			>
				{utterance.slice(loc.start, loc.end)}
			</span>,
		);
		cursor = loc.end;
	});
	parts.push(<span key="plain-tail">{utterance.slice(cursor)}</span>);
	return <>{parts}</>;
}

/** Corrected sentence with each located fix highlighted emerald. */
function HighlightedRepair({
	corrected,
	spans,
}: {
	corrected: string;
	spans: AnalysisSpan[];
}) {
	const located = spans
		.map((span) => ({
			span,
			loc: locateSpan(corrected, span.correction, 1),
		}))
		.filter(
			(
				entry,
			): entry is { span: AnalysisSpan; loc: { start: number; end: number } } =>
				entry.loc !== null,
		)
		.sort((a, b) => a.loc.start - b.loc.start);
	if (located.length === 0) {
		return <>{corrected}</>;
	}
	const parts: React.ReactNode[] = [];
	let cursor = 0;
	located.forEach(({ loc }) => {
		if (loc.start < cursor) {
			return;
		}
		parts.push(
			<span key={`plain-${cursor}`}>{corrected.slice(cursor, loc.start)}</span>,
		);
		parts.push(
			<span key={`fix-${loc.start}`} className="font-semibold text-accuracy">
				{corrected.slice(loc.start, loc.end)}
			</span>,
		);
		cursor = loc.end;
	});
	parts.push(<span key="plain-tail">{corrected.slice(cursor)}</span>);
	return <>{parts}</>;
}

export function TurnDiagnosticsCard({
	feedback,
	utterance,
	onReanalyze,
	pending,
	framed = true,
}: {
	feedback: TurnFeedback;
	utterance: string;
	onReanalyze: () => void;
	pending: boolean;
	/** Render without the outer card frame (inside a dialog, which frames). */
	framed?: boolean;
}) {
	const relevanceLabel = {
		relevant: m["voice.room.feedback.topic.relevant"](),
		partially_relevant: m["voice.room.feedback.topic.partiallyRelevant"](),
		off_topic: m["voice.room.feedback.topic.offTopic"](),
		not_applicable: m["voice.room.feedback.topic.notApplicable"](),
	}[feedback.relevance.status];
	// `SpanStatus` is exactly "incorrect" | "awkward", so this lookup is
	// exhaustive and tsc enforces it if a third status is ever added. Built as a
	// static object rather than an interpolated key so Paraglide's generated
	// types can check the message ids.
	const spanStatusLabel = {
		incorrect: m["voice.room.feedback.span.incorrect"](),
		awkward: m["voice.room.feedback.span.awkward"](),
	} as const;
	const alternatives: Array<{ text: string; reason?: string; kind: string }> =
		[];
	const languageAlt = feedback.alternatives.language;
	if (languageAlt != null) {
		alternatives.push({
			text: languageAlt.text,
			reason: languageAlt.reason ?? undefined,
			kind: m["voice.room.feedback.kind.language"](),
		});
	}
	const contextualAlt = feedback.alternatives.contextual;
	if (contextualAlt != null) {
		alternatives.push({
			text: contextualAlt.text,
			reason: contextualAlt.reason ?? undefined,
			kind: m["voice.room.feedback.kind.contextual"](),
		});
	}

	// The repair block only means something when the corrected sentence
	// differs; on a clean turn it would echo the utterance under a title.
	const showRepair =
		feedback.corrected.trim().toLowerCase() !== utterance.trim().toLowerCase();

	return (
		<div
			className={
				framed ? "surface-card flex flex-col gap-4 p-4" : "flex flex-col gap-4"
			}
		>
			<div className="flex items-center justify-between gap-2">
				<div className="flex items-center gap-2">
					<span className="h-2.5 w-2.5 rounded-full bg-primary" />
					<span className="text-sm font-semibold text-foreground">
						{m["voice.room.feedback.title"]()}
					</span>
				</div>
				<SubmitButton
					type="button"
					variant="outline"
					size="sm"
					pending={pending}
					loadingLabel={m["voice.room.transcript.analyzing"]()}
					onClick={onReanalyze}
				>
					<RefreshCw data-icon="inline-start" />
					{m["voice.room.transcript.reanalyze"]()}
				</SubmitButton>
			</div>
			<div className="flex flex-col gap-0.5 rounded-xl border-2 border-border bg-muted p-4">
				<span className="text-xs text-primary">
					{m["voice.room.feedback.utterance"]()}
				</span>
				<p className="text-base font-semibold leading-relaxed text-foreground">
					<HighlightedUtterance utterance={utterance} spans={feedback.spans} />
				</p>
			</div>

			{feedback.spans.map((span) => (
				<div
					key={`${span.text}-${span.occurrence}`}
					className="flex flex-col gap-0.5 rounded-xl border-2 border-border bg-muted p-4"
				>
					<div className="flex justify-between">
						<span className="text-xs text-primary">
							{m["voice.room.feedback.grammarCorrection"]()}
						</span>
						<Badge variant="outline">{spanStatusLabel[span.status]}</Badge>
					</div>
					<div className="flex flex-wrap items-center gap-2">
						<span className="text-sm text-ai-coral line-through opacity-85">
							{span.text}
						</span>
						<ArrowRight size={16} className="text-muted-foreground" />
						<span className="py-0.5 text-base font-semibold text-primary">
							{span.correction}
						</span>
					</div>
					<p className="text-sm">{span.reason}</p>
				</div>
			))}

			{showRepair ? (
				<div className="flex flex-col gap-2 rounded-xl border-2 border-border bg-accuracy-tint/50 p-4">
					<div className="flex items-center justify-between">
						<span className="text-xs font-medium text-accuracy">
							{m["voice.room.feedback.directRepair"]()}
						</span>
					</div>
					<p className="text-base font-medium leading-relaxed text-foreground">
						<HighlightedRepair
							corrected={feedback.corrected}
							spans={feedback.spans}
						/>
					</p>
				</div>
			) : null}

			<div className="flex items-start gap-2 rounded-xl border-2 border-border bg-muted p-4">
				<div className="flex flex-col gap-0.5 w-full">
					<div className="flex items-center gap-2 justify-between">
						<span className="text-xs font-semibold text-accent-violet">
							{m["voice.room.feedback.topic.coherence"]()}
						</span>
						<Badge variant="outline">{relevanceLabel}</Badge>
					</div>
					{feedback.relevance.reason ? (
						<p className="text-sm">{feedback.relevance.reason}</p>
					) : feedback.relevance.status === "relevant" ? (
						<p className="text-sm">
							{m["voice.room.feedback.topic.default"]()}
						</p>
					) : null}
				</div>
			</div>

			{alternatives.length > 0 ? (
				<div className="flex flex-col gap-2">
					<div className="flex items-center justify-between">
						<span className="text-sm font-semibold text-foreground">
							{m["voice.room.feedback.alternatives"]()}
						</span>
					</div>
					<div className="grid grid-cols-1 gap-2 md:grid-cols-2">
						{alternatives.map((alt) => (
							<div
								key={alt.kind}
								className="surface-card flex flex-col gap-1.5 p-4"
							>
								<Badge variant="outline" className="text-primary self-start">
									{alt.kind}
								</Badge>
								<span className="text-sm font-semibold text-foreground">
									“{alt.text}”
								</span>
								{alt.reason ? (
									<span className="text-sm text-muted-foreground">
										{alt.reason}
									</span>
								) : null}
							</div>
						))}
					</div>
				</div>
			) : null}

			<div className="flex items-center gap-2">
				<div className="tile bg-secondary text-primary">
					<Lightbulb size={17} />
				</div>
				<p className="text-sm">
					<span className="font-semibold text-foreground">
						{m["voice.room.feedback.coachTip"]()}{" "}
					</span>
					{feedback.tip}
				</p>
			</div>
		</div>
	);
}
