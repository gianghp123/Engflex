import { browserPreferences } from "#/lib/browser-preferences";

export const VIDEO_SPEEDS = [0.75, 1, 1.25, 1.5];
export const DICTATION_SPEEDS = [0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2];
const defaults = {
	speed: 1,
	showTranscript: true,
	showMedia: true,
	showPhonetic: false,
	showTranslation: true,
	autoPause: true,
	largeVideo: false,
	showCorrect: true,
	soundFeedback: true,
};
export type VideoPreferences = typeof defaults;
function parse(raw: unknown, speeds = VIDEO_SPEEDS): VideoPreferences {
	const value = (raw ?? {}) as Partial<VideoPreferences>;
	return Object.fromEntries(
		Object.entries(defaults).map(([key, fallback]) => [
			key,
			key === "speed"
				? speeds.includes(value.speed ?? 0)
					? value.speed
					: 1
				: typeof value[key as keyof VideoPreferences] === "boolean"
					? value[key as keyof VideoPreferences]
					: fallback,
		]),
	) as VideoPreferences;
}
export const dictationPreferences = browserPreferences(
	"engflex:videos:dictation:v1",
	defaults,
	(raw) => parse(raw, DICTATION_SPEEDS),
);
export const shadowingPreferences = browserPreferences(
	"engflex:videos:shadowing:v1",
	defaults,
	parse,
);
