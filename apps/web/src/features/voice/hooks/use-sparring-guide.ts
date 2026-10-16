import { useInstructionGuide } from "#/hooks/use-instruction-guide";

/** localStorage flag behind "Don't show this tip again". */
const HIDE_GUIDE_KEY = "engflex:hide-sparring-guide";

/**
 * Sparring guide visibility: opens on room mount unless dismissed, reopens on
 * demand. The shared instruction-guide hook owns the read/persist behavior.
 */
export function useSparringGuide() {
	return useInstructionGuide(HIDE_GUIDE_KEY);
}
