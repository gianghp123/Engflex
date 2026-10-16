import { RedirectToSignIn, Show } from "@clerk/tanstack-react-start";
import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { RouteErrorFallback } from "#/components/common/error-pages";
import { AdminShell } from "#/components/layout/admin-shell";
import { getAuthState } from "#/lib/auth-guard";

export const Route = createFileRoute("/admin")({
	beforeLoad: async () => {
		const { isAuthenticated, isAdmin } = await getAuthState();
		if (!isAuthenticated) {
			throw redirect({ href: APP_ROUTES.AUTH.SIGN_IN });
		}
		if (!isAdmin) {
			throw redirect({ to: APP_ROUTES.HOME });
		}
	},
	component: AdminLayout,
	errorComponent: RouteErrorFallback,
});

function AdminLayout() {
	return (
		<>
			<Show when="signed-in">
				<AdminShell>
					<Outlet />
				</AdminShell>
			</Show>
			<Show when="signed-out">
				<RedirectToSignIn />
			</Show>
		</>
	);
}
