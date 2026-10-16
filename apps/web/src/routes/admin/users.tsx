import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { UsersPanel } from "#/features/admin/users/components/users-panel";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/admin/users")({
	component: AdminUsers,
});

function AdminUsers() {
	useBreadcrumbs([
		{ label: m["admin.nav.overview"](), to: APP_ROUTES.ADMIN.OVERVIEW },
		{ label: m["admin.nav.users"]() },
	]);
	return (
		<div className="container-content flex flex-col gap-4 py-8">
			<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
				{m["admin.nav.users"]()}
			</h1>
			<UsersPanel />
		</div>
	);
}
