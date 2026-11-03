import type { ErrorType } from "@engflex/contracts";
import { useQuery } from "@tanstack/react-query";
import { listMistakes, listTopErrors } from "#/features/progress/service";

const progressKeys = {
	topErrors: (limit: number) => ["progress", "errors", limit] as const,
	mistakes: (limit: number, errorType: ErrorType | "all" | undefined) =>
		["progress", "mistakes", limit, errorType ?? "all"] as const,
};

/** The dashboard's "Top mistakes" StatCard feed (`GET /progress/errors`). */
export function useTopErrors(limit = 5) {
	return useQuery({
		queryKey: progressKeys.topErrors(limit),
		queryFn: () => listTopErrors(limit),
	});
}

/** The mistakes history feed (`GET /progress/mistakes`), filterable by type. */
export function useMistakes(params?: {
	limit?: number;
	errorType?: ErrorType | "all";
}) {
	const limit = params?.limit ?? 20;
	const errorType = params?.errorType ?? "all";
	return useQuery({
		queryKey: progressKeys.mistakes(limit, errorType),
		queryFn: () => listMistakes({ limit, errorType }),
	});
}
