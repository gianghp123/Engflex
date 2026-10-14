import type { Progress } from "@engflex/contracts";
import { useQuery } from "@tanstack/react-query";
import { API_ROUTES } from "#/app/api-routes";
import { api } from "#/lib/api";

export const progressKeys = {
	summary: ["progress"] as const,
};

/** The measurable-only dashboard rollup (GET /progress). */
export function useProgress() {
	return useQuery({
		queryKey: progressKeys.summary,
		queryFn: () =>
			api<Progress>(API_ROUTES.PROGRESS.SUMMARY, undefined, {
				withCredentials: true,
			}),
	});
}
