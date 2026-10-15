import { RedirectToSignIn, Show } from "@clerk/tanstack-react-start";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { RouteErrorFallback } from "#/components/common/error-pages";
import { AppShell } from "#/components/layout/app-shell";
import { getAuthState } from "#/lib/auth-guard";

export const Route = createFileRoute("/_app")({
	beforeLoad: async () => {
		const { isAuthenticated, onboardingCompleted } = await getAuthState();
		if (!isAuthenticated) {
			throw redirect({ href: APP_ROUTES.AUTH.SIGN_IN });
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
