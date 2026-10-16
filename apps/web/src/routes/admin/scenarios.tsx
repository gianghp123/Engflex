import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/admin/scenarios")({
	component: AdminScenarios,
});

function AdminScenarios() {
	useBreadcrumbs([
		{ label: m["admin.nav.overview"](), to: APP_ROUTES.ADMIN.OVERVIEW },
		{ label: m["admin.nav.scenarios"]() },
	]);
	return (
		<div className="container-content flex flex-col gap-2 py-8">
			<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
				{m["admin.nav.scenarios"]()}
			</h1>
		</div>
	);
}
