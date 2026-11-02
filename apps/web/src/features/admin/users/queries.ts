import type { AttemptType, GrantSubscription } from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	grantSubscription,
	listUserAttempts,
	listUsers,
	revokeSubscription,
	updateUserRole,
} from "#/features/admin/users/service";

const keys = {
	list: (page: number, pageSize: number, search: string) =>
		["admin", "users", page, pageSize, search] as const,
	attempts: (
		userId: string,
		page: number,
		pageSize: number,
		type?: AttemptType,
	) => ["admin", "users", userId, "attempts", page, pageSize, type] as const,
};

export function useAdminUsers(page: number, pageSize: number, search: string) {
	return useQuery({
		queryKey: keys.list(page, pageSize, search),
		queryFn: () => listUsers(page, pageSize, search),
	});
}

export function useAdminUserAttempts(
	userId: string,
	page: number,
	pageSize: number,
	type?: AttemptType,
) {
	return useQuery({
		queryKey: keys.attempts(userId, page, pageSize, type),
		queryFn: () => listUserAttempts(userId, page, pageSize, type),
		enabled: Boolean(userId),
	});
}

export function useUpdateUserRole() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ id, role }: { id: string; role: string }) =>
			updateUserRole(id, role),
		onSuccess: () =>
			Promise.all([
				queryClient.invalidateQueries({ queryKey: ["admin", "users"] }),
				queryClient.invalidateQueries({ queryKey: ["admin", "premium"] }),
				queryClient.invalidateQueries({ queryKey: ["payments"] }),
			]),
	});
}

export function useGrantSubscription() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			userId,
			body,
		}: {
			userId: string;
			body: GrantSubscription;
		}) => grantSubscription(userId, body),
		onSuccess: () =>
			Promise.all([
				queryClient.invalidateQueries({ queryKey: ["admin", "users"] }),
				queryClient.invalidateQueries({ queryKey: ["admin", "premium"] }),
				queryClient.invalidateQueries({ queryKey: ["payments"] }),
			]),
	});
}

export function useRevokeSubscription() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (userId: string) => revokeSubscription(userId),
		onSuccess: () =>
			Promise.all([
				queryClient.invalidateQueries({ queryKey: ["admin", "users"] }),
				queryClient.invalidateQueries({ queryKey: ["admin", "premium"] }),
				queryClient.invalidateQueries({ queryKey: ["payments"] }),
			]),
	});
}
