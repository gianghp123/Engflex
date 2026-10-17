import type { TurnFeedback } from "@engflex/contracts";
import { cn } from "cn";
import { ChevronDown, Sparkles } from "lucide-react";
import { useState } from "react";
import { Badge } from "#/components/ui/badge";
import {
	Collapsible,
	CollapsibleContent,
	CollapsibleTrigger,
} from "#/components/ui/collapsible";
import { TurnDiagnosticsCard } from "#/features/voice/components/transcript/turn-diagnostics-card";
import { AnalyzeTrigger } from "#/features/voice/components/transcript/turn-feedback";
import { m } from "#/paraglide/messages";

/** The one line that earns a collapsed row its place: the first fix-up, so a
 * reader can scan a past session without opening every turn. Null when the
 * analysis found nothing to correct. */
function summaryLine(feedback: TurnFeedback): string | null {
	const span = feedback.spans[0];
	return span ? `${span.text} → ${span.correction}` : null;
}

/** Review-surface analysis: an inline disclosure under the turn it belongs to.
 * The live room keeps the dialog, and not by taste — while a modal is open the
 * transcript behind it is inert, so arriving analyses are serialized by the
 * interaction itself. A review page has no interaction to serialize: every
 * analyzed turn mounts at the same instant, so one dialog each would stack.
 * A collapsed list cannot. */
export function TurnFeedbackDisclosure({
	feedback,
	utterance,
	pending,
	failed,
	onAnalyze,
}: {
	feedback?: TurnFeedback;
	utterance: string;
	pending: boolean;
	failed: boolean;
	onAnalyze: () => void;
}) {
	const [open, setOpen] = useState(false);

	if (!feedback) {
		return (
			<AnalyzeTrigger pending={pending} failed={failed} onAnalyze={onAnalyze} />
		);
	}

	const summary = summaryLine(feedback);

	return (
		<Collapsible open={open} onOpenChange={setOpen} className="w-full">
			<CollapsibleTrigger className="flex w-full items-center gap-2 rounded-xl border-2 border-border bg-muted px-3 py-2 text-left text-xs font-medium text-foreground transition-colors hover:bg-muted/70">
				<Sparkles className="size-3.5 shrink-0" />
				<span className="shrink-0">{m["voice.room.feedback.title"]()}</span>
				{summary ? (
					<span className="min-w-0 flex-1 truncate font-normal text-muted-foreground">
						{summary}
					</span>
				) : (
					<span className="flex-1" />
				)}
				{feedback.spans.length > 1 ? (
					<Badge variant="secondary" aria-hidden="true">
						{feedback.spans.length}
					</Badge>
				) : null}
				<ChevronDown
					aria-hidden="true"
					className={cn(
						"size-4 shrink-0 text-muted-foreground transition-transform",
						open && "rotate-180",
					)}
				/>
			</CollapsibleTrigger>
			<CollapsibleContent className="pt-2">
				<TurnDiagnosticsCard
					feedback={feedback}
					utterance={utterance}
					pending={pending}
					onReanalyze={onAnalyze}
					framed={false}
				/>
			</CollapsibleContent>
		</Collapsible>
	);
}
