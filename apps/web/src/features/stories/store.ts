import { browserPreferences } from "#/lib/browser-preferences";

const defaults = {
	fontSize: 18,
	fontFamily: "sans",
	lineHeight: 1.75,
	showTranslation: true,
};
export const readerPreferences = browserPreferences(
	"engflex:stories:reader:v1",
	defaults,
	(raw) => {
		const value = (raw ?? {}) as Partial<typeof defaults>;
		return {
			fontSize: [16, 18, 20, 24].includes(value.fontSize ?? 0)
				? (value.fontSize ?? 18)
				: 18,
			fontFamily: value.fontFamily === "serif" ? "serif" : "sans",
			lineHeight: [1.5, 1.75, 2].includes(value.lineHeight ?? 0)
				? (value.lineHeight ?? 1.75)
				: 1.75,
			showTranslation:
				typeof value.showTranslation === "boolean"
					? value.showTranslation
					: true,
		};
	},
);
