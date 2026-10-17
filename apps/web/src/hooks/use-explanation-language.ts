import { useCallback, useRef, useState } from "react";
import {
	explanationLanguage,
	explanationLanguageAsked,
	markExplanationLanguageAsked,
	setExplanationLanguage,
} from "#/lib/explanation-language";

/**
 * The language control and its first-use ask, scoped to one surface.
 *
 * `key` names that surface ("vocabulary", "writing", "analyze", …) and is the
 * only thing that scopes the ask: the first time the learner acts on a surface
 * the picker opens instead of the request, and the interrupted action runs
 * once a language is settled. The chosen value is shared across surfaces and
 * is not held here at all — `language()` reads it at request time.
 *
 * Nothing mirrors state through an effect (react-effects.md): the gate is the
 * learner's own click, so the click handler is where the state is set.
 */
export function useExplanationLanguage(key: string) {
	const [asking, setAsking] = useState(false);
	const pending = useRef<(() => void) | null>(null);

	/** Run `action`, asking for a language first when this surface has not
	 *  been asked yet. Call this from the surface's click handler. */
	const withLanguage = useCallback(
		(action: () => void) => {
			if (explanationLanguageAsked(key)) {
				action();
				return;
			}
			pending.current = action;
			setAsking(true);
		},
		[key],
	);

	const runPending = useCallback(() => {
		const run = pending.current;
		pending.current = null;
		run?.();
	}, []);

	/** Settle on `code`, remember the answer, and let the action proceed. */
	const confirm = useCallback(
		(code: string, askAgain: boolean) => {
			setExplanationLanguage(code);
			if (!askAgain) markExplanationLanguageAsked(key);
			setAsking(false);
			runPending();
		},
		[key, runPending],
	);

	/** Dismissed without choosing: the resolution order already has an answer,
	 *  so the interrupted action still proceeds. */
	const dismiss = useCallback(() => {
		markExplanationLanguageAsked(key);
		setAsking(false);
		runPending();
	}, [key, runPending]);

	/** The always-available way back in, from the surface's own trigger. */
	const reopen = useCallback(() => setAsking(true), []);

	return {
		asking,
		withLanguage,
		confirm,
		dismiss,
		reopen,
		language: explanationLanguage,
	};
}
