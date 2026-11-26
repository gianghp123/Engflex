import type { Checkout, CheckoutSession, Plan } from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api } from "#/lib/api";

export function listPlans(): Promise<Plan[]> {
	return api<Plan[]>(API_ROUTES.PAYMENTS.PLANS, undefined, {
		withCredentials: false,
	});
}

export function startCheckout(body: Checkout): Promise<CheckoutSession> {
	return api<CheckoutSession>(API_ROUTES.PAYMENTS.CHECKOUT, {
		method: "POST",
		body: JSON.stringify(body),
	});
}
