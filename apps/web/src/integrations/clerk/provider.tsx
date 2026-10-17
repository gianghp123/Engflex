import { enUS } from "@clerk/localizations/en-US";
import { viVN } from "@clerk/localizations/vi-VN";
import { ClerkProvider, useAuth } from "@clerk/tanstack-react-start";
import { useRouter } from "@tanstack/react-router";
import { useEffect, useRef } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { profileKeys } from "#/features/profiles/queries";
import { registerTokenGetter } from "#/lib/auth-token";
import { getLocale } from "#/paraglide/runtime";

/** Clerk's component copy follows the app locale (vi base + en). */
const CLERK_LOCALIZATION = { en: enUS, vi: viVN } as const;

/**
 * `signInUrl`/`signUpUrl` keep Clerk's own links — the sign-in card's "Sign
 * up" footer, `<SignInButton />`, `<RedirectToSignIn />` — on the app's splat
 * pages. Without them Clerk falls back to the hosted Account Portal
 * (`*.accounts.dev`), a page outside the app chrome.
 */
export default function AppClerkProvider({
	children,
}: {
	children: React.ReactNode;
}) {
	return (
		<ClerkProvider
			localization={CLERK_LOCALIZATION[getLocale()]}
			signInUrl={APP_ROUTES.AUTH.SIGN_IN}
			signUpUrl={APP_ROUTES.AUTH.SIGN_UP}
		>
			<TokenBridge />
			<ProfileCacheBridge />
			{children}
		</ClerkProvider>
	);
}

/**
 * Bridges Clerk's session token to non-component call sites (`lib/api.ts`).
 * Clerk's documented external-API pattern: `await getToken()` → Bearer token.
 */
function TokenBridge() {
	const { getToken } = useAuth();

	useEffect(() => {
		registerTokenGetter(() => getToken());
		return () => registerTokenGetter(null);
	}, [getToken]);

	return null;
}

/**
 * Drops the cached profile when the Clerk user changes.
 *
 * The profile entry is cached for the life of the app and the route guard
 * answers "has this user finished onboarding?" from it, so without this a
 * sign-out followed by a sign-in as someone else — all inside one app load,
 * since Clerk navigates client-side — would let the guard answer with the
 * previous user's flag: bypassing onboarding one way, or re-running it and
 * overwriting the new user's stored preferences the other.
 *
 * `useEffect` is the right tool here per docs/superpowers/rules/react-effects.md:
 * it syncs *outside-React* state (Clerk's session) into the Query cache — never
 * React state into React state.
 *
 * The first settled value only seeds the ref: `userId` is `undefined` while
 * Clerk loads, and treating that as a change would invalidate on every page
 * load.
 *
 * Invalidation, not `staleTime`, is what makes this correct: an invalidated
 * entry reports stale regardless of the configured staleness
 * (`Query.isStaleByTime` returns true when `state.isInvalidated`), so the
 * guard's next `query()` must fetch the new user's profile instead of
 * trusting the previous user's cached answer. The window that remains is the
 * gap between Clerk changing `userId` and this effect flushing.
 */
function ProfileCacheBridge() {
	const router = useRouter();
	const { isLoaded, userId } = useAuth();
	const seen = useRef<string | null | undefined>(undefined);

	useEffect(() => {
		if (!isLoaded) return;
		if (seen.current === undefined) {
			seen.current = userId;
			return;
		}
		if (seen.current !== userId) {
			seen.current = userId;
			// invalidateQueries marks the entry stale and refetches the active
			// observer (the sidebar badge), so the next guard read fetches the
			// new user's profile before answering.
			void router.options.context.queryClient.invalidateQueries({
				queryKey: profileKeys.me,
			});
		}
	}, [isLoaded, userId, router]);

	return null;
}
