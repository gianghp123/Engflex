import type {
	GrantSubscription,
	Pagination,
	Subscription,
	User,
} from "@engflex/contracts";
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

export function grantSubscription(
	userId: string,
	body: GrantSubscription,
): Promise<Subscription> {
	return api<Subscription>(API_ROUTES.ADMIN.USER_SUBSCRIPTIONS(userId), {
		method: "POST",
		body: JSON.stringify(body),
	});
}

export function revokeSubscription(userId: string): Promise<null> {
	return api<null>(API_ROUTES.ADMIN.USER_SUBSCRIPTIONS(userId), {
		method: "DELETE",
	});
}
