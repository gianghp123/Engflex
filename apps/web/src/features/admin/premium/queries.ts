import type {
	CreatePremiumFeature,
	SubscriptionPlan,
	SubscriptionStatus,
	UpdatePlan,
	UpdatePremiumFeature,
} from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	createPremiumFeature,
	deletePremiumFeature,
	listAdminPlans,
	listAdminSubscriptions,
	listPremiumFeatures,
	updatePlan,
	updatePremiumFeature,
} from "#/features/admin/premium/service";

const keys = {
	plans: ["admin", "premium-plans"] as const,
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
