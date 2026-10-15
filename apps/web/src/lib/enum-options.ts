/**
 * Build a picker's options from a contract enum's value array. `labels` is
 * required only for enums whose display text is prose (a Record<Enum, …> makes
 * a new member a compile error until its copy exists); code enums omit it and
 * render the code itself.
 */
export function enumOptions<T extends string>(
	values: readonly T[],
	labels?: Record<T, () => string>,
): Array<{ value: T; label: () => string }> {
	return values.map((value) => ({
		value,
		label: labels ? labels[value] : () => value,
	}));
}
