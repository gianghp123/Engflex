import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "#/components/ui/tabs";
import { VideoCategoriesPanel } from "#/features/admin/videos/components/video-categories-panel";
import { VideoExercisesPanel } from "#/features/admin/videos/components/video-exercises-panel";
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
		<div className="container-content flex flex-col gap-4 py-8">
			<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
				{m["admin.nav.videos"]()}
			</h1>
			<Tabs defaultValue="exercises">
				<TabsList>
					<TabsTrigger value="exercises">Exercises</TabsTrigger>
					<TabsTrigger value="categories">Categories</TabsTrigger>
				</TabsList>
				<TabsContent value="exercises">
					<VideoExercisesPanel />
				</TabsContent>
				<TabsContent value="categories">
					<VideoCategoriesPanel />
				</TabsContent>
			</Tabs>
		</div>
	);
}
