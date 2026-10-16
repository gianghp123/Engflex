/**
 * URL search-param coercion for `validateSearch`. Params are unvalidated: the
 * router's JSON parser converts numeric-looking values (`?page=2` arrives as
 * the number `2`) and leaves the rest as strings, so each helper accepts both
 * shapes and returns the value or `undefined` — never a guessed default. A
 * garbage URL then renders the unfiltered view instead of crashing or lying.
 */

/** A 1-based page number from `?page=…`; anything unrecognizable is "no page". */
export function parsePage(raw: unknown): number | undefined {
	const page = typeof raw === "string" ? Number(raw) : raw;
	return typeof page === "number" && Number.isInteger(page) && page >= 1
		? page
		: undefined;
}

/** A member of a closed set from the URL; anything else is "no filter". */
export function parseEnumParam<T extends string>(
	values: readonly T[],
	raw: unknown,
): T | undefined {
	return typeof raw === "string" && (values as readonly string[]).includes(raw)
		? (raw as T)
		: undefined;
}

/** `parseEnumParam` plus the filter convention's `all` sentinel. */
export function parseFilterParam<T extends string>(
	values: readonly T[],
	raw: unknown,
): "all" | T | undefined {
	if (raw === "all") return "all";
	return parseEnumParam(values, raw);
}
