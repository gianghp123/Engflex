import { ThumbsDown, ThumbsUp } from "lucide-react";
import { useState } from "react";
import { MoMascot } from "#/components/common/mo-mascot";
import { SubmitButton } from "#/components/common/submit-button";
import { TipCard } from "#/components/common/tip-card";
import { cn } from "#/lib/utils";
import { m } from "#/paraglide/messages";

const PACES = ["slow", "ok", "fast"] as const;

/** House answer to dynamic copy: an explicit record keyed by option,
 *  never an interpolated message id. */
const PACE_COPY = {
	slow: () => m["voice.room.survey.paceSlow"](),
	ok: () => m["voice.room.survey.paceOk"](),
	fast: () => m["voice.room.survey.paceFast"](),
} as const;

/**
 * Post-session 2-tap verdict. MOCK ONLY: no API, no persistence — submit
 * swaps in a thank-you (same frame, no layout jump), exactly as a saved
 * survey would. The real endpoint arrives with the feedback pipeline.
 * Rendered only above a non-empty transcript, so empty sessions never
 * see it.
 */
export function SurveyCard() {
	const [understood, setUnderstood] = useState<boolean | null>(null);
	const [pace, setPace] = useState<(typeof PACES)[number]>("ok");
	const [done, setDone] = useState(false);
	if (done) {
		return (
			<TipCard
				title={m["voice.room.survey.title"]()}
				body={m["voice.room.survey.done"]()}
				variant="heart"
			/>
		);
	}
	return (
		<div className="surface-card flex flex-col gap-4 p-5">
			<div className="flex items-center gap-3">
				<MoMascot variant="heart" size={40} />

				<div className="flex flex-col gap-1">
					<h3 className="text-base font-bold tracking-tight text-foreground">
						{m["voice.room.survey.title"]()}
					</h3>
				</div>
			</div>
			<ol className="flex list-decimal flex-col gap-4 pl-5 marker:font-bold marker:text-primary">
				<li>
					<div className="flex flex-col gap-2">
						<p className="text-sm font-bold text-foreground">
							{m["voice.room.survey.helpLabel"]()}
						</p>
						<div className="flex gap-2">
							<button
								type="button"
								aria-pressed={understood === true}
								onClick={() => setUnderstood(true)}
								className={cn(
									"chip flex-1 justify-center",
									understood === true && "chip-selected",
								)}
							>
								<ThumbsUp data-icon="inline-start" />
								{m["voice.room.survey.yes"]()}
							</button>
							<button
								type="button"
								aria-pressed={understood === false}
								onClick={() => setUnderstood(false)}
								className={cn(
									"chip flex-1 justify-center",
									understood === false && "chip-selected",
								)}
							>
								<ThumbsDown data-icon="inline-start" />
								{m["voice.room.survey.no"]()}
							</button>
						</div>
					</div>
				</li>
				<li>
					<div className="flex flex-col gap-2">
						<p className="text-sm font-bold text-foreground">
							{m["voice.room.survey.paceLabel"]()}
						</p>
						<div className="grid grid-cols-3 gap-2">
							{PACES.map((option) => (
								<button
									key={option}
									type="button"
									aria-pressed={pace === option}
									onClick={() => setPace(option)}
									className={cn(
										"chip justify-center",
										pace === option && "chip-selected",
									)}
								>
									{PACE_COPY[option]()}
								</button>
							))}
						</div>
					</div>
				</li>
			</ol>
			<SubmitButton
				type="button"
				className="btn btn-primary w-full"
				disabled={understood === null}
				onClick={() => setDone(true)}
			>
				{m["voice.room.survey.submit"]()}
			</SubmitButton>
		</div>
	);
}
