import type { Pagination, User } from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api";

export function listUsers(
	page: number,
	pageSize: number,
	search: string,
): Promise<{ items: User[]; pagination: Pagination }> {
	return apiPage<User>(API_ROUTES.ADMIN.USERS, undefined, {
		query: { page, pageSize, search },
	});
}

export function updateUserRole(id: string, role: string): Promise<User> {
	return api<User>(API_ROUTES.ADMIN.USER_ROLE(id), {
		method: "PATCH",
		body: JSON.stringify({ role }),
	});
}
