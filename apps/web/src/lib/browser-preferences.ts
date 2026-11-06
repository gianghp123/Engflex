import { useStore } from "@tanstack/react-store";
import { Store } from "@tanstack/store";
import { useEffect } from "react";

/** Browser-only preferences. SSR and hydration both start with defaults; the
 * effect synchronizes the external browser storage with the client store. */
export function browserPreferences<T extends object>(
	key: string,
	defaults: T,
	parse: (raw: unknown) => T,
) {
	const store = new Store(defaults);
	function set(update: Partial<T>) {
		const value = { ...store.state, ...update };
		store.setState(() => value);
		try {
			localStorage.setItem(key, JSON.stringify(value));
		} catch {
			/* Storage is optional. */
		}
	}
	function usePreferences() {
		const value = useStore(store);
		useEffect(() => {
			function read() {
				try {
					store.setState(() =>
						parse(JSON.parse(localStorage.getItem(key) ?? "null")),
					);
				} catch {
					store.setState(() => defaults);
				}
			}
			read();
			const sync = (event: StorageEvent) => {
				if (event.key === key || event.key === null) read();
			};
			window.addEventListener("storage", sync);
			return () => window.removeEventListener("storage", sync);
		}, []);
		return [value, set] as const;
	}
	return { usePreferences, reset: () => set(defaults) };
}
