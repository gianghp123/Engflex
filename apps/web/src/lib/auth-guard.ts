import { auth } from "@clerk/tanstack-react-start/server";
import { createServerFn } from "@tanstack/react-start";
import { getProfile } from "#/features/profiles/service";
import { ApiError } from "#/lib/api";

/**
 * Server-side auth + onboarding probe for route `beforeLoad` guards.
 *
 * Onboarding completion lives in the DB (user_profiles.onboarding_completed_at);
 * there is no JWT claim. `GET /profiles/me` resolves it: a profile row with a
 * completion timestamp means onboarded, a 404 means not. Any other error
 * (API down, 500) is rethrown to the route's error boundary rather than being
 * silently read as "not onboarded" — a completed user must never be bounced
 * back into onboarding because of an outage.
 *
 * Cost: TanStack Router runs `beforeLoad` for every matched route on every
 * URL/history update (parent-first), and `staleTime`/`shouldReload` do not
 * suppress it — they are loader-only. So each in-app navigation (and each
 * link-intent preload) performs one `auth()` token verify plus one indexed
 * `GET /profiles/me`. This is the accepted price of a single-source-of-truth
 * DB gate; a cached variant would reintroduce a second source of truth.
 */
export const getAuthState = createServerFn().handler(async () => {
	const { isAuthenticated } = await auth();
	if (!isAuthenticated) {
		return { isAuthenticated: false, onboardingCompleted: false };
	}
	try {
		const profile = await getProfile();
		return {
			isAuthenticated: true,
			onboardingCompleted: profile.onboardingCompletedAt != null,
		};
	} catch (err) {
		if (err instanceof ApiError && err.status === 404) {
			return { isAuthenticated: true, onboardingCompleted: false };
		}
		throw err;
	}
});
