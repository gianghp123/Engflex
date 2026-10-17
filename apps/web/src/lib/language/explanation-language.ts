import { getLocale, locales } from "#/paraglide/runtime";

/**
 * The learner's explanation language: the language the LLM writes its
 * definitions, tips and reasons in, plus the per-surface record of whether
 * that surface's first-use pick has been answered.
 *
 * Device-local by design — the value rides every LLM request explicitly, so
 * the server never has to read storage. See the design note on why this is not
 * a profile field.
 */
const PREFIX = "engflex:explanation-language";

/** Read one namespaced preference; null when unset or storage is unavailable. */
function readPreference(key: string): string | null {
	if (typeof window === "undefined") return null;
	try {
		return window.localStorage.getItem(`${PREFIX}:${key}`);
	} catch {
		return null;
	}
}

/** Write one namespaced preference. A failure (private mode, quota) leaves the
 *  session working — the choice just does not persist. */
function writePreference(key: string, value: string): void {
	if (typeof window === "undefined") return;
	try {
		window.localStorage.setItem(`${PREFIX}:${key}`, value);
	} catch {
		// Intentionally ignored.
	}
}

function isOffered(code: string | null): code is string {
	return code !== null && (locales as readonly string[]).includes(code);
}

/** The learner's chosen language, or null when they never made one. */
export function storedExplanationLanguage(): string | null {
	const raw = readPreference("value");
	return isOffered(raw) ? raw : null;
}

/**
 * The code every LLM request sends: the explicit choice, else the UI locale.
 * The server resolves the code and treats an empty or unusable value as its
 * default brief, so this only has to be right for the languages the picker
 * offers.
 */
export function explanationLanguage(): string {
	return storedExplanationLanguage() ?? getLocale();
}

export function setExplanationLanguage(code: string): void {
	writePreference("value", code);
}

/**
 * Whether `key`'s first-use pick has been answered. The key names the surface
 * ("vocabulary", "writing", "analyze", …): per surface, not per browser, so
 * each one introduces the control on its own first use rather than a screen
 * answering for a screen the learner has never seen.
 */
export function explanationLanguageAsked(key: string): boolean {
	return readPreference(`asked:${key}`) === "1";
}

export function markExplanationLanguageAsked(key: string): void {
	writePreference(`asked:${key}`, "1");
}
