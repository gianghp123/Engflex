/** A saved word is due when its schedule date has passed. */
export function isDue(srsDueAt: string | undefined): boolean {
	if (!srsDueAt) return false;
	return new Date(srsDueAt).getTime() <= Date.now();
}

/** The reader's local midnight — the boundary "today" means to them. */
export function startOfLocalDay(): Date {
	const start = new Date();
	start.setHours(0, 0, 0, 0);
	return start;
}

/**
 * True when an RFC3339 instant falls on the reader's local day. The next
 * midnight is derived by date arithmetic (not +24h), so a DST day is right.
 */
export function isToday(iso: string | undefined): boolean {
	if (!iso) return false;
	const start = startOfLocalDay();
	const next = new Date(start);
	next.setDate(next.getDate() + 1);
	const at = new Date(iso).getTime();
	return at >= start.getTime() && at < next.getTime();
}
