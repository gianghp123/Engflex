import { useAuth } from "@clerk/tanstack-react-start";
import type { Checkout } from "@engflex/contracts";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useEntitlements } from "#/features/payments/entitlements";
import {
	getSepayStatus,
	listPlans,
	startCheckout,
	startSepayCheckout,
} from "#/features/payments/service";
import { isAdminClaims } from "#/lib/roles";

/** Exported so a completed checkout can invalidate `me` — paying is the one
 *  thing that changes premium status from outside this module. */
export const PAYMENT_QUERY_KEYS = {
	plans: ["payments", "plans"] as const,
	sepayStatus: (reference: string) =>
		["payments", "sepay", "status", reference] as const,
};

const keys = PAYMENT_QUERY_KEYS;

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

export function useStartSepayCheckout() {
	return useMutation({
		mutationFn: (body: Checkout) => startSepayCheckout(body),
	});
}

/** Polls a SePay checkout's status every 3s while pending — there's no
 *  redirect to land on once paid (see SepayCheckoutSession), so the
 *  frontend is the one watching for the webhook's result to land.
 *  `enabled` lets the caller stop polling once the dialog closes. */
export function useSepayStatus(reference: string | null, enabled: boolean) {
	return useQuery({
		queryKey: keys.sepayStatus(reference ?? "none"),
		queryFn: () => getSepayStatus(reference ?? ""),
		enabled: enabled && Boolean(reference),
		refetchInterval: (query) =>
			query.state.data?.status === "pending" ? 3000 : false,
	});
}

/** The caller's own premium status, used to decide whether a premium-
 *  gated card's action should proceed or redirect to pricing. isPending
 *  defaults the card to "treat as not premium" (safer default than
 *  letting a slow fetch allow a flash of premium content). A selector over
 *  the unified /premium/me query (entitlementKeys.me): one query key, one
 *  fetch, no duplicate ["payments", "me"] observer.
 *  Returns the useEntitlements result with data narrowed to MyPremiumStatus. */
export function useMyPremiumStatus() {
	const query = useEntitlements();
	return {
		...query,
		data: query.data ? { isPremium: query.data.isPremium } : undefined,
	};
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
