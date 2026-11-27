import { useInstructionGuide } from "#/hooks/use-instruction-guide";

/** localStorage flag behind the credit guide's "Don't show this tip again". */
const HIDE_GUIDE_KEY = "engflex:hide-credit-guide";

/**
 * Credit guide visibility: opens on the usage page unless dismissed with
 * "don't show again", and reopens on demand from the page header. The shared
 * instruction-guide hook owns the read/persist behavior.
 */
export function useCreditGuide() {
	return useInstructionGuide(HIDE_GUIDE_KEY);
}
