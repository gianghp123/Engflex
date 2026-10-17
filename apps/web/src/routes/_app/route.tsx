import { RedirectToSignIn, Show } from "@clerk/tanstack-react-start";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { RouteErrorFallback } from "#/components/common/error-pages";
import { AppShell } from "#/components/layout/app-shell";
import { profileQueryOptions } from "#/features/profiles/queries";
import { ApiError } from "#/lib/api";

export const Route = createFileRoute("/_app")({
	beforeLoad: async ({ context: { queryClient }, preload }) => {
		// Speculative lane: a hover must never cost a request.
		if (preload) return;

		// One read answers both questions. The profile entry is the same one the
		// sidebar's useProfile() renders, so the guard and the badge share it
		// instead of each fetching the profile.
		let onboardingCompleted = false;
		try {
			const profile = await queryClient.query(profileQueryOptions());
			onboardingCompleted = profile.onboardingCompletedAt != null;
		} catch (err) {
			// No Bearer: a signed-out request reaches the API unauthenticated,
			// so it 401s. Same convention as lib/route-loader.ts — behind the
			// auth guard a 401 is token plumbing, not an error page.
			if (err instanceof ApiError && err.status === 401) {
				throw redirect({ href: APP_ROUTES.AUTH.SIGN_IN });
			}
			// No profile row yet: that is "not onboarded", not a failure.
			if (!(err instanceof ApiError && err.status === 404)) {
				throw err;
			}
		}
		if (!onboardingCompleted) {
			throw redirect({ to: APP_ROUTES.ONBOARDING });
		}
	},
	component: AppLayout,
	errorComponent: RouteErrorFallback,
});

function AppLayout() {
	return (
		<>
			<Show when="signed-in">
				<AppShell>
					<Outlet />
				</AppShell>
			</Show>
			<Show when="signed-out">
				<RedirectToSignIn />
			</Show>
		</>
	);
}
