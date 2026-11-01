import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/admin/")({
	component: AdminOverview,
});

function AdminOverview() {
	useBreadcrumbs([
		{ label: m["admin.nav.overview"](), to: APP_ROUTES.ADMIN.OVERVIEW },
	]);
	return (
		<div className="container-content flex flex-col gap-2 py-8">
			<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
				{m["admin.overview.title"]()}
			</h1>
			<p className="text-muted-foreground">
				{m["admin.overview.description"]()}
			</p>
		</div>
	);
}
