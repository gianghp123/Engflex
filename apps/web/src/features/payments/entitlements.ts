import type { MyEntitlements, TrialStatus } from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { API_ROUTES } from "#/app/api-routes";
import { api } from "#/lib/api";

export const entitlementKeys = {
	me: ["access", "me"] as const,
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
