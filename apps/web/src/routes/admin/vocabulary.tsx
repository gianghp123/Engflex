import { createFileRoute } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "#/components/ui/tabs";
import { CategoriesPanel } from "#/features/admin/vocabulary/components/categories-panel";
import { DecksPanel } from "#/features/admin/vocabulary/components/decks-panel";
import { DictionaryItemsPanel } from "#/features/admin/vocabulary/components/dictionary-items-panel";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/admin/vocabulary")({
	component: AdminVocabulary,
});

function AdminVocabulary() {
	useBreadcrumbs([
		{ label: m["admin.nav.overview"](), to: APP_ROUTES.ADMIN.OVERVIEW },
		{ label: m["admin.nav.vocabulary"]() },
	]);
	return (
		<div className="container-content flex flex-col gap-4 py-8">
			<h1 className="text-[28px] leading-[36px] font-bold text-foreground">
				{m["admin.nav.vocabulary"]()}
			</h1>
			<Tabs defaultValue="items">
				<TabsList>
					<TabsTrigger value="items">Dictionary</TabsTrigger>
					<TabsTrigger value="decks">Decks</TabsTrigger>
					<TabsTrigger value="categories">Categories</TabsTrigger>
				</TabsList>
				<TabsContent value="items">
					<DictionaryItemsPanel />
				</TabsContent>
				<TabsContent value="decks">
					<DecksPanel />
				</TabsContent>
				<TabsContent value="categories">
					<CategoriesPanel />
				</TabsContent>
			</Tabs>
		</div>
	);
}
