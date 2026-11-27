import type { CEFR } from "@engflex/contracts";
import { CEFR_VALUES } from "@engflex/contracts";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useMemo } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { CardShell } from "#/components/common/card-shell";
import { EmptyList } from "#/components/common/empty-list";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { useStoryBookmarks } from "#/features/bookmarks/queries";
import { SavedStories } from "#/features/stories/components/saved-stories";
import {
	StoryCard,
	StoryCardSkeleton,
} from "#/features/stories/components/story-card";
import { StoryFilterBar } from "#/features/stories/components/story-filter-bar";
import { useStories, useStoryTopics } from "#/features/stories/queries";
import {
	STORIES_PAGE_SIZE,
	type StoriesSearch,
} from "#/features/stories/types";
import { parseFilterParam, parsePage } from "#/lib/router/search-params";
import { m } from "#/paraglide/messages";

const SKELETON_KEYS = Array.from(
	{ length: STORIES_PAGE_SIZE },
	(_, i) => `story-skeleton-${i + 1}`,
);

export const Route = createFileRoute("/_app/stories/")({
	validateSearch: (search: Record<string, unknown>): StoriesSearch => ({
		tab: search.tab === "saved" ? "saved" : undefined,
		cefr: parseFilterParam(CEFR_VALUES, search.cefr),
		topic:
			typeof search.topic === "string" && search.topic.trim() !== ""
				? search.topic.trim()
				: undefined,
		search:
			typeof search.search === "string" && search.search.trim() !== ""
				? search.search.trim()
				: undefined,
		page: parsePage(search.page),
	}),
	component: StoriesListPage,
});

function StoriesListPage() {
	useBreadcrumbs([{ label: m["stories.title"]() }]);
	const navigate = useNavigate();
	const searchParams = Route.useSearch();

	const cefr = searchParams.cefr ?? "all";
	const selectedCategory = searchParams.topic ?? "all";
	const search = searchParams.search ?? "";

	const topicsQuery = useStoryTopics();
	const categories = topicsQuery.data ?? [];

	const bookmarksQuery = useStoryBookmarks();
	const savedIds = useMemo(
		() => new Set((bookmarksQuery.data ?? []).map((b) => b.contentId)),
		[bookmarksQuery.data],
	);

	const page = searchParams.page ?? 1;
	const storiesQuery = useStories(
		{
			page,
			pageSize: STORIES_PAGE_SIZE,
			cefr: cefr === "all" ? undefined : cefr,
			topic: selectedCategory === "all" ? undefined : selectedCategory,
			search: search.trim() || undefined,
		},
		searchParams.tab !== "saved",
	);

	const items = storiesQuery.data?.items ?? [];
	const pagination = storiesQuery.data?.pagination;

	/** Any filter change resets to page 1 — a stale page number past the new
	 *  result set's last page would just render empty. */
	function updateSearch(next: Partial<StoriesSearch>) {
		void navigate({
			to: APP_ROUTES.STORIES.LIST,
			search: {
				...searchParams,
				...next,
				tab:
					next.tab !== undefined
						? next.tab === "all"
							? undefined
							: next.tab
						: searchParams.tab,
				page: undefined,
			},
		});
	}

	function goToPage(nextPage: number) {
		void navigate({
			to: APP_ROUTES.STORIES.LIST,
			search: { ...searchParams, page: nextPage > 1 ? nextPage : undefined },
		});
	}

	function handleReset() {
		void navigate({
			to: APP_ROUTES.STORIES.LIST,
			search: searchParams.tab === "saved" ? { tab: "saved" } : {},
		});
	}

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="nice" size={84} />,
				title: m["stories.title"](),
				description: m["stories.description"](),
			}}
		>
			<StoryFilterBar
				tab={searchParams.tab ?? "all"}
				onTabChange={(tab) => updateSearch({ tab })}
				level={cefr}
				onLevelChange={(lvl) =>
					updateSearch({ cefr: lvl === "all" ? undefined : (lvl as CEFR) })
				}
				categories={categories}
				selectedCategory={selectedCategory}
				onCategoryChange={(cat) =>
					updateSearch({ topic: cat === "all" ? undefined : cat })
				}
				search={search}
				onSearchChange={(q) => updateSearch({ search: q ? q : undefined })}
			/>

			{searchParams.tab === "saved" ? (
				<SavedStories
					key={`${cefr}:${selectedCategory}:${search}`}
					filters={searchParams}
					onReset={handleReset}
				/>
			) : storiesQuery.isError ? (
				<CardShell className="p-4">
					<p role="alert">{m["stories.bookmarks.error"]()}</p>
					<SubmitButton
						pending={storiesQuery.isFetching}
						onClick={() => void storiesQuery.refetch()}
					>
						{m["stories.bookmarks.retry"]()}
					</SubmitButton>
				</CardShell>
			) : storiesQuery.isPending ? (
				<div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
					{SKELETON_KEYS.map((key) => (
						<StoryCardSkeleton key={key} />
					))}
				</div>
			) : items.length === 0 ? (
				<EmptyList title={m["stories.empty"]()} onReset={handleReset} />
			) : (
				<>
					<div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
						{items.map((story) => (
							<StoryCard
								key={story.id}
								story={{ ...story, saved: savedIds.has(story.id) }}
							/>
						))}
					</div>
					{pagination && pagination.totalPages > 1 ? (
						<div className="flex items-center justify-center gap-2">
							<Button
								type="button"
								variant="outline"
								size="icon-sm"
								disabled={pagination.page <= 1}
								onClick={() => goToPage(pagination.page - 1)}
							>
								<ChevronLeft />
							</Button>
							<span className="text-sm text-muted-foreground">
								{pagination.page}/{pagination.totalPages}
							</span>
							<Button
								type="button"
								variant="outline"
								size="icon-sm"
								disabled={pagination.page >= pagination.totalPages}
								onClick={() => goToPage(pagination.page + 1)}
							>
								<ChevronRight />
							</Button>
						</div>
					) : null}
				</>
			)}
		</PageLayout>
	);
}
