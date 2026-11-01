/** A word is due once its SRS date has arrived. No fixture clock: the
 *  backend never sets srsDueAt yet (no scheduler), so this always reads
 *  false until that lands — which is the honest state, not a bug here. */
export function isDue(srsDueAt: string | undefined): boolean {
	if (!srsDueAt) return false;
	const today = new Date().toISOString().slice(0, 10);
	return srsDueAt.slice(0, 10) <= today;
}
