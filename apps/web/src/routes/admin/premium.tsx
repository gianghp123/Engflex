import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "#/components/ui/tabs";
import { FeaturesPanel } from "#/features/admin/premium/components/features-panel";
import { PlansPanel } from "#/features/admin/premium/components/plans-panel";
import { SubscriptionsPanel } from "#/features/admin/premium/components/subscriptions-panel";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/admin/premium")({
	component: AdminPremium,
});

function AdminPremium() {
	useBreadcrumbs([
		{ label: m["admin.nav.overview"](), to: APP_ROUTES.ADMIN.OVERVIEW },
		{ label: m["admin.nav.premium"]() },
	]);
	return (
		<div className="container-content flex flex-col gap-4 py-8">
			<div>
				<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
					{m["admin.premium.title"]()}
				</h1>
				<p className="text-sm text-muted-foreground">
					{m["admin.premium.description"]()}
				</p>
			</div>
			<Tabs defaultValue="plans">
				<TabsList>
					<TabsTrigger value="plans">
						{m["admin.premium.tabs.plans"]()}
					</TabsTrigger>
					<TabsTrigger value="features">
						{m["admin.premium.tabs.features"]()}
					</TabsTrigger>
					<TabsTrigger value="subscriptions">
						{m["admin.premium.tabs.subscriptions"]()}
					</TabsTrigger>
				</TabsList>
				<TabsContent value="plans">
					<PlansPanel />
				</TabsContent>
				<TabsContent value="features">
					<FeaturesPanel />
				</TabsContent>
				<TabsContent value="subscriptions">
					<SubscriptionsPanel />
				</TabsContent>
			</Tabs>
		</div>
	);
}
