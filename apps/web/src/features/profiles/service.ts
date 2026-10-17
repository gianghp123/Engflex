import type { Preferences, Profile } from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api } from "#/lib/api/api";

/** The caller's profile, or an ApiError(404) before onboarding. */
export function getProfile(): Promise<Profile> {
	return api<Profile>(API_ROUTES.PROFILES.ME, undefined, {
		withCredentials: true,
	});
}

/** Complete onboarding / update preferences (PUT /profiles/me). */
export function updateProfile(preferences: Preferences): Promise<Profile> {
	return api<Profile>(
		API_ROUTES.PROFILES.ME,
		{
			method: "PUT",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(preferences),
		},
		{ withCredentials: true },
	);
}
