import type { ErrorType } from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api } from "#/lib/api/api";

/**
 * One taxonomy bucket with its count. Matches the backend `ErrorStat` DTO
 * (inline primitives, no tygo): `{errorType, count}`.
 */
export interface ErrorStat {
	errorType: string;
	count: number;
}

/**
 * One flattened learner error for the mistakes display. Matches the backend
 * `Mistake` DTO (inline primitives, no tygo): the span plus where it came
 * from and when.
 */
export interface Mistake {
	errorType: string;
	text: string;
	correction: string;
	reason: string;
	source: "feedback" | "writing";
	createdAt: string;
}

/** Top error buckets, most common first (`GET /progress/errors`). */
export function listTopErrors(limit = 5): Promise<ErrorStat[]> {
	return api<ErrorStat[]>(API_ROUTES.PROGRESS.ERRORS, undefined, {
		withCredentials: true,
		query: { limit },
	});
}

/** Recent mistake rows, newest first (`GET /progress/mistakes`). */
export function listMistakes(params?: {
	limit?: number;
	errorType?: ErrorType | "all";
}): Promise<Mistake[]> {
	const query: Record<string, string | number> = {};
	query.limit = String(params?.limit ?? 20);
	if (params?.errorType && params.errorType !== "all") {
		query.errorType = params.errorType;
	}
	return api<Mistake[]>(API_ROUTES.PROGRESS.MISTAKES, undefined, {
		withCredentials: true,
		query,
	});
}

export type { ErrorType };
