import { useMutation, useQuery } from "@tanstack/react-query";
import { getProfile, updateProfile } from "#/features/profiles/service";

export const profileKeys = {
	me: ["profile", "me"] as const,
};

export function profileQueryOptions() {
	return {
		queryKey: profileKeys.me,
		// Cached for the life of the app. This entry is read by the route guard
		// (routes/_app/route.tsx) for the onboarding flag, so a per-navigation
		// refetch would put a request on every navigation — and the guard's read
		// is the same entry the sidebar's useProfile() renders, so the profile
		// is fetched once per document instead of twice.
		//
		// There is exactly one writer today — onboarding's `finish()`, which
		// writes through with `setQueryData` before navigating. Refresh comes
		// from the writes that change the profile, not from a timer, so ANY
		// future profile mutation must write through or invalidate
		// `profileKeys.me`.
		staleTime: Number.POSITIVE_INFINITY,
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
