import { useMutation, useQuery } from "@tanstack/react-query";
import { getProfile, updateProfile } from "#/features/profiles/service";

export const profileKeys = {
	me: ["profile", "me"] as const,
};

export function profileQueryOptions() {
	return {
		queryKey: profileKeys.me,
		queryFn: () => getProfile(),
	};
}

/** The caller's profile (GET /profiles/me). Only present once onboarded. */
export function useProfile() {
	return useQuery(profileQueryOptions());
}

/** Completes onboarding; the caller handles navigation. */
export function useUpdateProfile() {
	return useMutation({ mutationFn: updateProfile });
}
