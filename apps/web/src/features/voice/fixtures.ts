import {
	AudioWaveform,
	Clock3,
	Gauge,
	type LucideIcon,
	Users,
} from "lucide-react";

/** mock-only: no contract type yet — voice mode selector cards. */
export interface ModeCard {
	id: "spontaneous" | "structured";
	icon: LucideIcon;
	metaIcon: LucideIcon;
}

/** Verbatim from the voice mode-selector mock. */
export const MODE_CARDS: ModeCard[] = [
	{
		id: "spontaneous",
		icon: AudioWaveform,
		metaIcon: Clock3,
	},
	{
		id: "structured",
		icon: Users,
		metaIcon: Gauge,
	},
];
