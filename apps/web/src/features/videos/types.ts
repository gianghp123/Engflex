import type {
	VideoCategoryResponse,
	VideoExerciseResponse,
} from "@engflex/contracts";

export interface Video extends VideoExerciseResponse {
	views?: number;
	isPro?: boolean;
	source?: "youtube";
	practice?: { dictation: boolean; shadowing: boolean };
}

export interface VideoSection {
	category: VideoCategoryResponse;
	total: number;
	videos: Video[];
}
