import type { TurnFeedback } from "@engflex/contracts";
import { Sparkles } from "lucide-react";
import { useState } from "react";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { Dialog, DialogContent, DialogTitle } from "#/components/ui/dialog";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#/components/ui/tooltip";
import { TurnDiagnosticsCard } from "#/features/voice/components/transcript/turn-diagnostics-card";
import { m } from "#/paraglide/messages";

/** The analyze control a transcript panel needs per learner turn. The live
 * panels wire it to the analyze mutation; the DEV preview injects a noop so
 * the same container renders with fixtures and zero network. */
export type AnalyzeControl = {
	pending: boolean;
	failed: boolean;
	onAnalyze: (position: number) => void;
};

/** The analyze affordance for a turn with no analysis yet. The live dialog and
 * the review disclosure both render it, so the two presentations cannot
 * drift. The language control lives on the surface that owns the transcript,
 * not here: an affordance per turn would put one picker on every bubble. */
export function AnalyzeTrigger({
	pending,
	failed,
	onAnalyze,
}: {
	pending: boolean;
	failed: boolean;
	onAnalyze: () => void;
}) {
	return (
		<div className="flex flex-col items-start gap-1">
			<Tooltip>
				<TooltipTrigger asChild>
					<SubmitButton
						type="button"
						variant="outline"
						size="icon-sm"
						pending={pending}
						aria-label={m["voice.room.transcript.analyze"]()}
						onClick={onAnalyze}
					>
						<Sparkles />
					</SubmitButton>
				</TooltipTrigger>
				<TooltipContent>
					{pending
						? m["voice.room.transcript.analyzing"]()
						: m["voice.room.transcript.analyze"]()}
				</TooltipContent>
			</Tooltip>
			{failed ? (
				<p className="text-xs text-destructive">
					{m["voice.room.transcript.analyzeFailed"]()}
				</p>
			) : null}
		</div>
	);
}

/** Analyze affordance and result for one learner turn, as a dialog. Live-room
 * only: a modal makes the transcript behind it inert, so arriving analyses are
 * serialized by the interaction itself. The review surfaces mount every
 * analyzed turn at once and use TurnFeedbackDisclosure instead.
 *
 * Pending and error state come from the mutation, not local useState: a local
 * flag would clear the spinner before the async request resolves. `open` is
 * written only from clicks — the learner's own Analyze, or the "view analysis"
 * trigger — never from an effect watching arriving data (react-effects.md: a
 * user action belongs in an event handler). */
export function TurnFeedbackPanel({
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

	if (feedback) {
		const label = m["voice.room.feedback.viewAnalysis"]();
		return (
			<>
				<Tooltip>
					<TooltipTrigger asChild>
						<Button
							type="button"
							variant="outline"
							size="icon-sm"
							aria-label={label}
							onClick={() => setOpen(true)}
						>
							<Sparkles />
						</Button>
					</TooltipTrigger>
					<TooltipContent>{label}</TooltipContent>
				</Tooltip>
				<Dialog open={open} onOpenChange={setOpen}>
					<DialogContent className="max-h-[calc(100dvh-2rem)] overflow-y-auto sm:max-w-2xl">
						<DialogTitle className="sr-only">
							{m["voice.room.feedback.title"]()}
						</DialogTitle>
						<TurnDiagnosticsCard
							feedback={feedback}
							utterance={utterance}
							pending={pending}
							onReanalyze={onAnalyze}
							framed={false}
						/>
					</DialogContent>
				</Dialog>
			</>
		);
	}
	return (
		<AnalyzeTrigger
			pending={pending}
			failed={failed}
			onAnalyze={() => {
				setOpen(true);
				onAnalyze();
			}}
		/>
	);
}
