import type { MyEntitlements, MyUsage, TrialStatus } from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { API_ROUTES } from "#/app/api-routes";
import { ApiError, api } from "#/lib/api";

export const entitlementKeys = {
	me: ["access", "me"] as const,
	usage: (page: number, pageSize: number) =>
		["access", "usage", page, pageSize] as const,
};

export function getEntitlements(): Promise<MyEntitlements> {
	return api<MyEntitlements>(API_ROUTES.PAYMENTS.ME);
}

/** Fire-and-observe trial start: an already-used trial resolves to
 *  `{started:false}` and must never fail the caller. */
export function startTrial(): Promise<TrialStatus> {
	return api<TrialStatus>(API_ROUTES.PAYMENTS.TRIAL, { method: "POST" });
}

export function useEntitlements() {
	return useQuery({
		queryKey: entitlementKeys.me,
		queryFn: getEntitlements,
	});
}

/** The caller's usage view: plan, both counters, the per-action costs table,
 *  and the paginated charge history. Quota/premium/auth denials settle
 *  without a retry (the page renders its error state); real failures retry
 *  with the default backoff. */
export function getMyUsage(page: number, pageSize: number): Promise<MyUsage> {
	return api<MyUsage>(API_ROUTES.PAYMENTS.USAGE, undefined, {
		query: { page, pageSize },
	});
}

export function useMyUsage(page: number, pageSize: number) {
	return useQuery({
		queryKey: entitlementKeys.usage(page, pageSize),
		queryFn: () => getMyUsage(page, pageSize),
		retry: (failureCount: number, error: unknown) =>
			!(
				error instanceof ApiError && [401, 403, 404, 429].includes(error.status)
			) && failureCount < 3,
	});
}

export function useStartTrial() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: startTrial,
		onSuccess: () => {
			void queryClient.invalidateQueries({
				queryKey: entitlementKeys.me,
			});
		},
	});
}
