import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { listUsers, updateUserRole } from "#/features/admin/users/service";

const keys = {
	list: (page: number, pageSize: number, search: string) =>
		["admin", "users", page, pageSize, search] as const,
};

export function useAdminUsers(page: number, pageSize: number, search: string) {
	return useQuery({
		queryKey: keys.list(page, pageSize, search),
		queryFn: () => listUsers(page, pageSize, search),
	});
}

export function useUpdateUserRole() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({ id, role }: { id: string; role: string }) =>
			updateUserRole(id, role),
		onSuccess: () =>
			queryClient.invalidateQueries({ queryKey: ["admin", "users"] }),
	});
}
