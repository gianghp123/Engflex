import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { StoriesPanel } from "#/features/admin/stories/components/stories-panel";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/admin/stories")({
	component: AdminStories,
});

function AdminStories() {
	useBreadcrumbs([
		{ label: m["admin.nav.overview"](), to: APP_ROUTES.ADMIN.OVERVIEW },
		{ label: m["admin.nav.stories"]() },
	]);
	return (
		<div className="container-content flex flex-col gap-4 py-8">
			<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
				{m["admin.nav.stories"]()}
			</h1>
			<StoriesPanel />
		</div>
	);
}
