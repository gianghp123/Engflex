import type { Activity, LessonStatus } from "@engflex/contracts";
import { Store } from "@tanstack/store";

/** Local-only progress derivation (never a wire shape). */
export interface LessonProgress {
	status: LessonStatus;
	percent: number /* int */;
	partsCompleted: number /* int */;
	partsTotal: number /* int */;
}

export type LessonsState = {
	completedParts: Record<string, number[]>;
	/** Activity totals learned when a part completes (the practice page knows
	 *  the activity list); lets hub-level reads size `done/total` without a
	 *  server count field. */
	lessonTotals: Record<string, number>;
};

/** Local-only progress until attempts land. Keys are lesson ids. */
export const lessonsStore = new Store<LessonsState>({
	completedParts: {},
	lessonTotals: {},
});

/** Maps a server summary's completed activity ids onto the lesson's part
 *  numbers (unknown ids — edited lesson configs — are dropped). */
export function partNumbersFor(
	activities: Activity[],
	ids: string[],
): number[] {
	const byId = new Map(
		activities.map((activity) => [activity.id, activity.partNumber]),
	);
	return ids
		.map((id) => byId.get(id))
		.filter((part): part is number => part !== undefined);
}

/** Pure derivation — usable from a `useStore` selector for reactive reads.
 *  `fallbackTotal` is the caller's freshly fetched activity count; 0/undefined
 *  means "not known yet", so the learned total (if any) is used instead. */
export function progressFrom(
	state: LessonsState,
	lessonId: string,
	fallbackTotal?: number,
	server?: { count: number; completed: boolean },
): LessonProgress {
	const completed = state.completedParts[lessonId] ?? [];
	// A part never un-completes and the server set is monotonic, so the
	// larger count is the merged truth (local overlay ∪ server rollup).
	const partsCompleted = Math.max(completed.length, server?.count ?? 0);
	const partsTotal =
		fallbackTotal && fallbackTotal > 0
			? fallbackTotal
			: (state.lessonTotals[lessonId] ?? 0);
	const percent =
		partsTotal === 0
			? 0
			: Math.min(100, Math.round((partsCompleted / partsTotal) * 100));
	const status: LessonStatus =
		server?.completed || percent >= 100
			? "completed"
			: percent > 0
				? "in_progress"
				: "unstarted";
	return {
		status,
		percent,
		partsCompleted,
		partsTotal,
	};
}

/** First activity without a completion entry; a finished lesson restarts at the first. */
export function getFirstIncompleteActivity(
	lessonId: string,
	activities: Activity[],
	serverParts: number[] = [],
): Activity | undefined {
	const completed = new Set([
		...(lessonsStore.state.completedParts[lessonId] ?? []),
		...serverParts,
	]);
	return (
		activities.find((activity) => !completed.has(activity.partNumber)) ??
		activities[0]
	);
}

export function completePart(
	lessonId: string,
	partNumber: number,
	partsTotal: number,
): void {
	lessonsStore.setState((state) => {
		const current = state.completedParts[lessonId] ?? [];
		const lessonTotals = { ...state.lessonTotals, [lessonId]: partsTotal };
		if (current.includes(partNumber)) return { ...state, lessonTotals };
		return {
			...state,
			lessonTotals,
			completedParts: {
				...state.completedParts,
				[lessonId]: [...current, partNumber].sort((a, b) => a - b),
			},
		};
	});
}
