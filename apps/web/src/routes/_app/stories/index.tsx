import type { CEFR } from "@engflex/contracts";
import { CEFR_VALUES } from "@engflex/contracts";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { CardShell } from "#/components/common/card-shell";
import { EmptyList } from "#/components/common/empty-list";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { Pager } from "#/components/common/pager";
import { SubmitButton } from "#/components/common/submit-button";
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
		page: parsePage(search.page),
		cefr: parseFilterParam(CEFR_VALUES, search.cefr),
		topic:
			typeof search.topic === "string" && search.topic.trim() !== ""
				? search.topic.trim()
				: undefined,
		search:
			typeof search.search === "string" && search.search.trim() !== ""
				? search.search.trim()
				: undefined,
	}),
	component: StoriesListPage,
});

function StoriesListPage() {
	useBreadcrumbs([{ label: m["stories.title"]() }]);
	const navigate = useNavigate();
	const searchParams = Route.useSearch();

	const page = searchParams.page ?? 1;
	const cefr = searchParams.cefr ?? "all";
	const selectedCategory = searchParams.topic ?? "all";
	const search = searchParams.search ?? "";

	const topicsQuery = useStoryTopics();
	const categories = topicsQuery.data ?? [];

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
	const total = pagination?.total ?? 0;
	const totalPages = pagination?.totalPages ?? 1;

	const from = total === 0 ? 0 : (page - 1) * STORIES_PAGE_SIZE + 1;
	const to = Math.min(page * STORIES_PAGE_SIZE, total);

	function updateSearch(next: Partial<StoriesSearch>) {
		void navigate({
			to: APP_ROUTES.STORIES.LIST,
			search: {
				...searchParams,
				...next,
				page:
					next.page !== undefined
						? next.page <= 1
							? undefined
							: next.page
						: undefined,
			},
		});
	}

	function handleReset() {
		void navigate({
			to: APP_ROUTES.STORIES.LIST,
			search: {},
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
				onTabChange={(tab) => updateSearch({ tab, page: 1 })}
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
					filters={searchParams}
					onPageChange={(page) => updateSearch({ page })}
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
			) : total === 0 ? (
				<EmptyList title={m["stories.empty"]()} onReset={handleReset} />
			) : (
				<>
					<div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6">
						{items.map((story) => (
							<StoryCard key={story.id} story={story} />
						))}
					</div>
					<Pager
						page={page}
						totalPages={totalPages}
						summary={
							totalPages > 1
								? m["stories.showingRange"]({ from, to, total })
								: m["stories.showingAll"]({ total })
						}
						onPageChange={(next) =>
							void navigate({
								to: APP_ROUTES.STORIES.LIST,
								search: {
									...searchParams,
									page: next <= 1 ? undefined : next,
								},
							})
						}
					/>
				</>
			)}
		</PageLayout>
	);
}
