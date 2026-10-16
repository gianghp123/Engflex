import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/admin/videos")({
	component: AdminVideos,
});

function AdminVideos() {
	useBreadcrumbs([
		{ label: m["admin.nav.overview"](), to: APP_ROUTES.ADMIN.OVERVIEW },
		{ label: m["admin.nav.videos"]() },
	]);
	return (
		<div className="container-content flex flex-col gap-2 py-8">
			<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
				{m["admin.nav.videos"]()}
			</h1>
		</div>
	);
}
