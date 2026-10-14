import type { Turn } from "@engflex/contracts";
import {
	AudioWaveform,
	Clock3,
	Gauge,
	type LucideIcon,
	Users,
} from "lucide-react";
import architectureReviews from "#/assets/topics/architecture-reviews.png";
import crossFunctionalSync from "#/assets/topics/cross-functional-sync.png";
import jobInterviews from "#/assets/topics/job-interviews.png";
import productPitch from "#/assets/topics/product-pitch.png";

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

export const TOPIC_IMAGES: Record<string, string> = {
	"job-interviews": jobInterviews,
	"architecture-reviews": architectureReviews,
	"cross-functional-sync": crossFunctionalSync,
	"product-pitch": productPitch,
};

/**
 * Mock end-of-session review, shown when a finished conversation has no
 * persisted turns (e.g. the mic never captured anything). Throwaway: delete
 * once real turns flow end to end.
 */
export const REVIEW_MOCK_TURNS: Turn[] = [
	{
		id: "mock-turn-1",
		position: 1,
		role: "ai",
		text: "Hi! What would you like to talk about today?",
		wasInterrupted: false,
		createdAt: "2026-09-26T15:00:00Z",
	},
	{
		id: "mock-turn-2",
		position: 2,
		role: "user",
		text: "I want to practice ordering coffee.",
		wasInterrupted: false,
		feedback: {
			payload: {
				corrected: "I want to practice ordering coffee.",
				spans: [],
				relevance: { status: "relevant", reason: undefined },
				alternatives: {
					language: {
						text: "I'd like to practice ordering coffee.",
						reason: "“I'd like to” sounds more natural than “I want to”.",
					},
					contextual: undefined,
				},
				tip: "Good opener — “I'd like to” sounds a little more polite.",
			},
		},
		createdAt: "2026-09-26T15:00:12Z",
	},
	{
		id: "mock-turn-3",
		position: 3,
		role: "ai",
		text: "Great choice. How did the review go — what went well?",
		wasInterrupted: false,
		createdAt: "2026-09-26T15:00:20Z",
	},
	{
		id: "mock-turn-4",
		position: 4,
		role: "user",
		text: "We shipped on time, but I struggled to explain the delay.",
		wasInterrupted: false,
		createdAt: "2026-09-26T15:00:35Z",
	},
];
