import type {
	CreatePremiumFeature,
	Pagination,
	Plan,
	PremiumFeature,
	Subscription,
	SubscriptionPlan,
	SubscriptionStatus,
	UpdatePlan,
	UpdatePremiumFeature,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api";

export function listAdminPlans(): Promise<Plan[]> {
	return api<Plan[]>(API_ROUTES.ADMIN.PREMIUM_PLANS);
}

export function updatePlan(
	plan: SubscriptionPlan,
	payload: UpdatePlan,
): Promise<Plan> {
	return api<Plan>(API_ROUTES.ADMIN.PREMIUM_PLAN(plan), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function listPremiumFeatures(
	page: number,
	pageSize: number,
): Promise<{ items: PremiumFeature[]; pagination: Pagination }> {
	return apiPage<PremiumFeature>(API_ROUTES.ADMIN.PREMIUM_FEATURES, undefined, {
		query: { page, pageSize },
	});
}

export function createPremiumFeature(
	payload: CreatePremiumFeature,
): Promise<PremiumFeature> {
	return api<PremiumFeature>(API_ROUTES.ADMIN.PREMIUM_FEATURES, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updatePremiumFeature(
	id: string,
	payload: UpdatePremiumFeature,
): Promise<PremiumFeature> {
	return api<PremiumFeature>(API_ROUTES.ADMIN.PREMIUM_FEATURE(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deletePremiumFeature(id: string): Promise<void> {
	return api<void>(API_ROUTES.ADMIN.PREMIUM_FEATURE(id), {
		method: "DELETE",
	});
}

export function listAdminSubscriptions(
	page: number,
	pageSize: number,
	status: SubscriptionStatus | "",
	plan: SubscriptionPlan | "",
): Promise<{ items: Subscription[]; pagination: Pagination }> {
	return apiPage<Subscription>(
		API_ROUTES.ADMIN.PREMIUM_SUBSCRIPTIONS,
		undefined,
		{
			query: { page, pageSize, status, plan },
		},
	);
}
