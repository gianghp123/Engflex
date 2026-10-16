import { useEffect, useState } from "react";

/** localStorage flag behind a guide's "don't show this tip again". */
function isGuideDismissed(key: string): boolean {
	try {
		return window.localStorage.getItem(key) != null;
	} catch {
		return false;
	}
}

/**
 * One instruction guide's visibility: opens on mount unless dismissed with
 * "don't show again", persists that choice under `key`, and reopens on demand.
 * `autoOpen: false` keeps the guide closed and trigger-only — for surfaces
 * that show the help button but never introduce the tip themselves.
 *
 * localStorage (theme-toggle precedent) — a client-only preference with no
 * SSR need, so reads happen in an effect and SSR and first paint agree.
 */
export function useInstructionGuide(key: string, autoOpen = true) {
	const [open, setOpen] = useState(false);

	useEffect(() => {
		if (autoOpen && !isGuideDismissed(key)) {
			setOpen(true);
		}
	}, [key, autoOpen]);

	const close = (dontShowAgain: boolean) => {
		if (dontShowAgain) {
			try {
				window.localStorage.setItem(key, "1");
			} catch {
				// Private mode etc: the guide simply shows again next visit.
			}
		}
		setOpen(false);
	};

	return { open, setOpen, close };
}
