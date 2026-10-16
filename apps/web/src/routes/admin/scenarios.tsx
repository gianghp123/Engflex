import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "#/components/ui/tabs";
import { PersonasPanel } from "#/features/admin/scenarios/components/personas-panel";
import { ScenariosPanel } from "#/features/admin/scenarios/components/scenarios-panel";
import { TopicsPanel } from "#/features/admin/scenarios/components/topics-panel";
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
		<div className="container-content flex flex-col gap-4 py-8">
			<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
				{m["admin.nav.scenarios"]()}
			</h1>
			<Tabs defaultValue="scenarios">
				<TabsList>
					<TabsTrigger value="scenarios">Scenarios</TabsTrigger>
					<TabsTrigger value="topics">Topics</TabsTrigger>
					<TabsTrigger value="personas">Personas</TabsTrigger>
				</TabsList>
				<TabsContent value="scenarios">
					<ScenariosPanel />
				</TabsContent>
				<TabsContent value="topics">
					<TopicsPanel />
				</TabsContent>
				<TabsContent value="personas">
					<PersonasPanel />
				</TabsContent>
			</Tabs>
		</div>
	);
}
