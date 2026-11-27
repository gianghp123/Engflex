import type {
	CreatePremiumFeature,
	PlanKey,
	SubscriptionPlan,
	SubscriptionStatus,
	UpdatePlan,
	UpdatePremiumFeature,
} from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	createPremiumFeature,
	deletePremiumFeature,
	getTrialTerms,
	listActionCosts,
	listAdminLimits,
	listAdminPlans,
	listAdminSubscriptions,
	listPremiumFeatures,
	type UpdateActionCosts,
	type UpdatePlanLimits,
	type UpdateTrialTerms,
	updateActionCosts,
	updatePlan,
	updatePlanLimits,
	updatePremiumFeature,
	updateTrialTerms,
} from "#/features/admin/premium/service";

const keys = {
	plans: ["admin", "premium-plans"] as const,
	limits: ["admin", "premium-limits"] as const,
	actionCosts: ["admin", "action-costs"] as const,
	trialTerms: ["admin", "trial-terms"] as const,
	features: (page: number, pageSize: number) =>
		["admin", "premium-features", page, pageSize] as const,
	subscriptions: (
		page: number,
		pageSize: number,
		status: string,
		plan: string,
	) =>
		["admin", "premium-subscriptions", page, pageSize, status, plan] as const,
};

function useInvalidate(key: readonly unknown[]) {
	const queryClient = useQueryClient();
	return () => queryClient.invalidateQueries({ queryKey: key });
}

export function useAdminPlans() {
	return useQuery({
		queryKey: keys.plans,
		queryFn: listAdminPlans,
	});
}

export function useAdminLimits() {
	return useQuery({
		queryKey: keys.limits,
		queryFn: listAdminLimits,
	});
}

export function useUpdatePlanLimits() {
	const invalidate = useInvalidate(["admin", "premium-limits"]);
	return useMutation({
		mutationFn: ({
			plan,
			payload,
		}: {
			plan: PlanKey;
			payload: UpdatePlanLimits;
		}) => updatePlanLimits(plan, payload),
		onSuccess: invalidate,
	});
}

export function useUpdatePlan() {
	const invalidate = useInvalidate(["admin", "premium-plans"]);
	return useMutation({
		mutationFn: ({
			plan,
			payload,
		}: {
			plan: SubscriptionPlan;
			payload: UpdatePlan;
		}) => updatePlan(plan, payload),
		onSuccess: invalidate,
	});
}

export function useAdminPremiumFeatures(page: number, pageSize: number) {
	return useQuery({
		queryKey: keys.features(page, pageSize),
		queryFn: () => listPremiumFeatures(page, pageSize),
	});
}

export function useCreatePremiumFeature() {
	const invalidate = useInvalidate(["admin", "premium-features"]);
	return useMutation({
		mutationFn: (payload: CreatePremiumFeature) =>
			createPremiumFeature(payload),
		onSuccess: invalidate,
	});
}

export function useUpdatePremiumFeature() {
	const invalidate = useInvalidate(["admin", "premium-features"]);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdatePremiumFeature;
		}) => updatePremiumFeature(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeletePremiumFeature() {
	const invalidate = useInvalidate(["admin", "premium-features"]);
	return useMutation({
		mutationFn: (id: string) => deletePremiumFeature(id),
		onSuccess: invalidate,
	});
}

export function useAdminSubscriptions(
	page: number,
	pageSize: number,
	status: SubscriptionStatus | "",
	plan: SubscriptionPlan | "",
) {
	return useQuery({
		queryKey: keys.subscriptions(page, pageSize, status, plan),
		queryFn: () => listAdminSubscriptions(page, pageSize, status, plan),
	});
}

export function useAdminActionCosts() {
	return useQuery({ queryKey: keys.actionCosts, queryFn: listActionCosts });
}

export function useUpdateActionCosts() {
	const invalidate = useInvalidate(keys.actionCosts);
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (payload: UpdateActionCosts) => updateActionCosts(payload),
		onSuccess: () => {
			invalidate();
			// The pricing pages read the public catalog; refresh it in-session too.
			void queryClient.invalidateQueries({ queryKey: ["pricing", "catalog"] });
		},
	});
}

export function useAdminTrialTerms() {
	return useQuery({ queryKey: keys.trialTerms, queryFn: getTrialTerms });
}

export function useUpdateTrialTerms() {
	const invalidate = useInvalidate(keys.trialTerms);
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (payload: UpdateTrialTerms) => updateTrialTerms(payload),
		onSuccess: () => {
			invalidate();
			void queryClient.invalidateQueries({ queryKey: ["pricing", "catalog"] });
		},
	});
}
