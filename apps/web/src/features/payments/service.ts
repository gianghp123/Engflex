import type {
	Checkout,
	CheckoutSession,
	MyPremiumStatus,
	PricingCatalog,
	SepayCheckoutSession,
	SepayStatus,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api } from "#/lib/api";

export function getPricingCatalog(): Promise<PricingCatalog> {
	return api<PricingCatalog>(API_ROUTES.PRICING.CATALOG, undefined, {
		withCredentials: false,
	});
}

export function startCheckout(body: Checkout): Promise<CheckoutSession> {
	return api<CheckoutSession>(API_ROUTES.PAYMENTS.CHECKOUT, {
		method: "POST",
		body: JSON.stringify(body),
	});
}

export function getMyPremiumStatus(): Promise<MyPremiumStatus> {
	return api<MyPremiumStatus>(API_ROUTES.PAYMENTS.ME);
}

export function startSepayCheckout(
	body: Checkout,
): Promise<SepayCheckoutSession> {
	return api<SepayCheckoutSession>(API_ROUTES.PAYMENTS.SEPAY_CHECKOUT, {
		method: "POST",
		body: JSON.stringify(body),
	});
}

export function getSepayStatus(reference: string): Promise<SepayStatus> {
	return api<SepayStatus>(API_ROUTES.PAYMENTS.SEPAY_STATUS(reference));
}
