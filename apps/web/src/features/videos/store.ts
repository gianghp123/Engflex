import { browserPreferences } from "#/lib/browser-preferences";

export const VIDEO_SPEEDS = [0.75, 1, 1.25, 1.5];
const defaults = {
	speed: 1,
	showTranscript: true,
	showMedia: true,
	showPhonetic: false,
	showTranslation: true,
	autoPause: true,
	largeVideo: false,
};
export type VideoPreferences = typeof defaults;
function parse(raw: unknown): VideoPreferences {
	const value = (raw ?? {}) as Partial<VideoPreferences>;
	return Object.fromEntries(
		Object.entries(defaults).map(([key, fallback]) => [
			key,
			key === "speed"
				? VIDEO_SPEEDS.includes(value.speed ?? 0)
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
	parse,
);
export const shadowingPreferences = browserPreferences(
	"engflex:videos:shadowing:v1",
	defaults,
	parse,
);
