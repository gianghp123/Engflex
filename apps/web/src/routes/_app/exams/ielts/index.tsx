import { useQuery } from "@tanstack/react-query";
import { createFileRoute, Link } from "@tanstack/react-router";
import { History } from "lucide-react";
import { useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageHero } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import {
	IeltsCategoryCard,
	type IeltsCategoryCardItem,
} from "#/features/exams/components/ielts/ielts-category-card";
import { examCategoriesQueryOptions } from "#/features/exams/queries";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/exams/ielts/")({
	component: IeltsHubPage,
});

const CATEGORY_COLORS = [
	"bg-[#0284c7]",
	"bg-[#ea580c]",
	"bg-[#9333ea]",
	"bg-[#2563eb]",
];

function CategoryCardSkeleton() {
	return (
		<div className="flex items-center justify-between gap-4 rounded-2xl border border-border/80 bg-card p-5 shadow-sm">
			<div className="flex min-w-0 flex-1 flex-col gap-3">
				<Skeleton className="h-6 w-3/4 rounded-md" />
				<div className="flex gap-1.5">
					<Skeleton className="h-4 w-14 rounded-md" />
					<Skeleton className="h-4 w-14 rounded-md" />
				</div>
				<Skeleton className="h-4 w-28 rounded-md" />
			</div>
			<Skeleton className="size-14 shrink-0 rounded-2xl md:size-16" />
		</div>
	);
}

function IeltsHubPage() {
	useBreadcrumbs([{ label: m["exams.ielts.hub.title"]() }]);
	const [subTypeFilter, setSubTypeFilter] = useState<
		"all" | "academic" | "general_training"
	>("all");
	const [skillFilter, setSkillFilter] = useState<
		"all" | "reading" | "listening"
	>("all");

	const {
		data: categories,
		isLoading,
		isError,
		refetch,
	} = useQuery(examCategoriesQueryOptions({ type: "ielts" }));

	const filteredCategories = (categories ?? []).filter((c) => {
		if (subTypeFilter !== "all" && c.subType !== subTypeFilter) return false;
		return true;
	});

	return (
		<div className="flex flex-col gap-8 pb-12">
			<div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
				<PageHero
					icon={<MoMascot variant="nice" size={84} />}
					title={m["exams.ielts.hub.title"]()}
					description={m["exams.ielts.hub.description"]()}
				/>
				<div className="flex shrink-0 items-center gap-3">
					<Link
						to={APP_ROUTES.EXAMS.IELTS.HISTORY}
						className="inline-flex items-center gap-2 rounded-xl border border-border/80 bg-background/80 px-4 py-2.5 text-sm font-semibold text-foreground shadow-sm backdrop-blur transition-colors hover:bg-muted"
					>
						<History className="size-4" aria-hidden="true" />
						<span>{m["exams.ielts.hub.historyButton"]()}</span>
					</Link>
				</div>
			</div>

			<div className="flex flex-wrap items-center gap-2 border-b border-border/60 pb-4">
				<Button
					variant={subTypeFilter === "all" ? "default" : "outline"}
					size="sm"
					onClick={() => setSubTypeFilter("all")}
					className="rounded-full"
				>
					{m["exams.ielts.hub.filterAll"]()}
				</Button>
				<Button
					variant={subTypeFilter === "academic" ? "default" : "outline"}
					size="sm"
					onClick={() => setSubTypeFilter("academic")}
					className="rounded-full"
				>
					{m["exams.ielts.hub.filterAcademic"]()}
				</Button>
				<Button
					variant={subTypeFilter === "general_training" ? "default" : "outline"}
					size="sm"
					onClick={() => setSubTypeFilter("general_training")}
					className="rounded-full"
				>
					{m["exams.ielts.hub.filterGeneral"]()}
				</Button>
				<div className="mx-2 h-4 w-px bg-border/80" />
				<Button
					variant={skillFilter === "reading" ? "secondary" : "ghost"}
					size="sm"
					onClick={() =>
						setSkillFilter(skillFilter === "reading" ? "all" : "reading")
					}
					className="rounded-full"
				>
					{m["exams.ielts.hub.filterReading"]()}
				</Button>
				<Button
					variant={skillFilter === "listening" ? "secondary" : "ghost"}
					size="sm"
					onClick={() =>
						setSkillFilter(skillFilter === "listening" ? "all" : "listening")
					}
					className="rounded-full"
				>
					{m["exams.ielts.hub.filterListening"]()}
				</Button>
			</div>

			{isLoading ? (
				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{[1, 2, 3, 4, 5, 6].map((i) => (
						<CategoryCardSkeleton key={i} />
					))}
				</div>
			) : isError ? (
				<div className="flex flex-col items-center justify-center gap-3 rounded-2xl border border-border/80 bg-card p-10 text-center">
					<p className="text-sm font-medium text-destructive">
						{m["exams.ielts.hub.loadFailed"]()}
					</p>
					<Button
						variant="outline"
						size="sm"
						onClick={() => void refetch()}
						className="rounded-lg"
					>
						{m["exams.ielts.hub.retry"]()}
					</Button>
				</div>
			) : filteredCategories.length === 0 ? (
				<div className="flex flex-col items-center justify-center gap-4 rounded-3xl border border-dashed border-border/80 bg-card/50 p-12 text-center">
					<MoMascot variant="thinking" className="size-24" />
					<div className="flex flex-col gap-1">
						<h3 className="text-lg font-bold text-foreground">
							{m["exams.ielts.hub.emptyTitle"]()}
						</h3>
						<p className="max-w-md text-sm text-muted-foreground">
							{m["exams.ielts.hub.emptyDesc"]()}
						</p>
					</div>
				</div>
			) : (
				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{filteredCategories.map((category, index) => {
						const item: IeltsCategoryCardItem = {
							id: category.id,
							title: category.title,
							subType: category.subType,
							testsCount: category.testsCount ?? 0,
							iconBgColor:
								CATEGORY_COLORS[index % CATEGORY_COLORS.length] ?? "bg-primary",
						};
						return <IeltsCategoryCard key={category.id} item={item} />;
					})}
				</div>
			)}
		</div>
	);
}
