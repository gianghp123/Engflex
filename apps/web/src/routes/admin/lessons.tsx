import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "#/components/ui/tabs";
import { LessonsPanel } from "#/features/admin/lessons/components/lessons-panel";
import { SectionsPanel } from "#/features/admin/lessons/components/sections-panel";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/admin/lessons")({
	component: AdminLessons,
});

function AdminLessons() {
	useBreadcrumbs([
		{ label: m["admin.nav.overview"](), to: APP_ROUTES.ADMIN.OVERVIEW },
		{ label: m["admin.nav.lessons"]() },
	]);
	return (
		<div className="container-content flex flex-col gap-4 py-8">
			<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
				{m["admin.nav.lessons"]()}
			</h1>
			<Tabs defaultValue="lessons">
				<TabsList>
					<TabsTrigger value="lessons">Lessons</TabsTrigger>
					<TabsTrigger value="sections">Sections</TabsTrigger>
				</TabsList>
				<TabsContent value="lessons">
					<LessonsPanel />
				</TabsContent>
				<TabsContent value="sections">
					<SectionsPanel />
				</TabsContent>
			</Tabs>
		</div>
	);
}
