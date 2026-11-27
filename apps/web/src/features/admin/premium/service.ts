import type {
	AccessAction,
	AdminActionCost,
	AdminTrialTerms,
	CounterKind,
	CreatePremiumFeature,
	Pagination,
	Plan,
	PlanKey,
	PremiumFeature,
	Subscription,
	SubscriptionPlan,
	SubscriptionStatus,
	UpdatePlan,
	UpdatePremiumFeature,
	UsagePeriod,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api";

/** Local wire copy of models.PlanFeatureLimit (the catalog rows): the
 *  contracts package does not export this model, so the admin console keeps
 *  the same field names here for one utils-free mapping. */
export interface PlanFeatureLimit {
	planKey: PlanKey;
	counterKey: CounterKind;
	isEnabled: boolean;
	limitValue: number;
	period: UsagePeriod;
	createdAt: string;
	updatedAt: string;
}

/** Local wire copy of requests.PlanLimitInput. */
export interface PlanLimitInput {
	counterKey: CounterKind;
	isEnabled: boolean;
	limitValue: number;
	period: UsagePeriod;
}

/** Local wire copy of requests.UpdatePlanLimits. */
export interface UpdatePlanLimits {
	counters: PlanLimitInput[];
}

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

export function listAdminLimits(): Promise<PlanFeatureLimit[]> {
	return api<PlanFeatureLimit[]>(API_ROUTES.ADMIN.PREMIUM_LIMITS);
}

export function updatePlanLimits(
	plan: PlanKey,
	payload: UpdatePlanLimits,
): Promise<void> {
	return api<void>(API_ROUTES.ADMIN.PREMIUM_LIMIT(plan), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

/** Local request shape: the access requests are not tygo-exported. */
export interface ActionCostInput {
	action: AccessAction;
	credits: number;
}

/** Local wire copy of requests.UpdateActionCosts. */
export interface UpdateActionCosts {
	costs: ActionCostInput[];
}

/** Local wire copy of requests.UpdateTrialTerms. */
export interface UpdateTrialTerms {
	days: number;
	credits: number;
	voiceSessions: number;
}

export function listActionCosts(): Promise<AdminActionCost[]> {
	return api<AdminActionCost[]>(API_ROUTES.ADMIN.PREMIUM_ACTION_COSTS);
}

export function updateActionCosts(payload: UpdateActionCosts): Promise<void> {
	return api<void>(API_ROUTES.ADMIN.PREMIUM_ACTION_COSTS, {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function getTrialTerms(): Promise<AdminTrialTerms> {
	return api<AdminTrialTerms>(API_ROUTES.ADMIN.PREMIUM_TRIAL_TERMS);
}

export function updateTrialTerms(payload: UpdateTrialTerms): Promise<void> {
	return api<void>(API_ROUTES.ADMIN.PREMIUM_TRIAL_TERMS, {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}
