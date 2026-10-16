import { Mic } from "lucide-react";
import { SubmitButton } from "#/components/common/submit-button";
import { m } from "#/paraglide/messages";

/** The two voice modes the hub offers; per-mode copy lives in MODE_COPY. */
export const VOICE_MODES = ["spontaneous", "structured"] as const;
export type VoiceMode = (typeof VOICE_MODES)[number];

const MODE_COPY: Record<
	VoiceMode,
	{
		badge: () => string;
		title: () => string;
		description: () => string;
		meta: () => string;
		features: Array<() => string>;
		cta: () => string;
	}
> = {
	spontaneous: {
		badge: () => m["voice.modes.option.spontaneous.badge"](),
		title: () => m["voice.modes.option.spontaneous.title"](),
		description: () => m["voice.modes.option.spontaneous.description"](),
		meta: () => m["voice.modes.option.spontaneous.meta"](),
		features: [
			() => m["voice.modes.option.spontaneous.feature0"](),
			() => m["voice.modes.option.spontaneous.feature1"](),
			() => m["voice.modes.option.spontaneous.feature2"](),
		],
		cta: () => m["voice.modes.option.spontaneous.cta"](),
	},
	structured: {
		badge: () => m["voice.modes.option.structured.badge"](),
		title: () => m["voice.modes.option.structured.title"](),
		description: () => m["voice.modes.option.structured.description"](),
		meta: () => m["voice.modes.option.structured.meta"](),
		features: [
			() => m["voice.modes.option.structured.feature0"](),
			() => m["voice.modes.option.structured.feature1"](),
			() => m["voice.modes.option.structured.feature2"](),
		],
		cta: () => m["voice.modes.option.structured.cta"](),
	},
};

export function ModeCard({
	mode,
	onStart,
	pending = false,
}: {
	mode: VoiceMode;
	onStart: () => void;
	pending?: boolean;
}) {
	const copy = MODE_COPY[mode];
	const primary = mode === "spontaneous";

	return (
		<div className="surface-card flex h-full flex-col gap-4 p-5">
			<span className="chip w-fit">
				<span aria-hidden="true" className="size-1.5 rounded-full bg-primary" />
				{copy.badge()}
			</span>
			<h2 className="text-xl font-bold text-foreground">{copy.title()}</h2>
			<p className="text-[15px] font-medium text-muted-foreground">
				{copy.description()}
			</p>
			<div className="flex flex-wrap gap-1.5">
				{copy.features.map((feature) => (
					<span key={feature()} className="chip">
						{feature()}
					</span>
				))}
			</div>
			<p className="rounded-lg border-2 border-border bg-accent p-3 text-[15px] font-medium text-foreground">
				{copy.meta()}
			</p>
			<div className="mt-auto pt-1">
				<SubmitButton
					type="button"
					variant={primary ? "default" : "outline"}
					className={
						primary ? "btn btn-primary w-full" : "btn btn-outline w-full"
					}
					onClick={onStart}
					pending={pending}
				>
					{primary ? <Mic data-icon="inline-start" /> : null}
					{copy.cta()}
				</SubmitButton>
			</div>
		</div>
	);
}
