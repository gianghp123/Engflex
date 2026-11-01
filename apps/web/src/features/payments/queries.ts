import { useAuth } from "@clerk/tanstack-react-start";
import type { Checkout } from "@engflex/contracts";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
	getMyPremiumStatus,
	listPlans,
	startCheckout,
} from "#/features/payments/service";
import { isAdminClaims } from "#/lib/roles";

const keys = {
	plans: ["payments", "plans"] as const,
	me: ["payments", "me"] as const,
};

export function usePlans() {
	return useQuery({
		queryKey: keys.plans,
		queryFn: listPlans,
	});
}

export function useStartCheckout() {
	return useMutation({
		mutationFn: (body: Checkout) => startCheckout(body),
	});
}

/** The caller's own premium status, used to decide whether a premium-
 *  gated card's action should proceed or redirect to pricing. isPending
 *  defaults the card to "treat as not premium" (safer default than
 *  letting a slow fetch allow a flash of premium content). */
export function useMyPremiumStatus() {
	return useQuery({
		queryKey: keys.me,
		queryFn: getMyPremiumStatus,
	});
}

/** Whether the caller may open a premium-gated card's normal action
 *  (vs. being redirected to pricing instead). Admins always pass — same
 *  preview bypass the backend applies — and the status fetch defaults to
 *  blocked while loading. */
export function useCanAccessPremium(): boolean {
	const { sessionClaims } = useAuth();
	const { data } = useMyPremiumStatus();
	return isAdminClaims(sessionClaims) || Boolean(data?.isPremium);
}
