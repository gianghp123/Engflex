import type { VideoCategoryPreview } from "@engflex/contracts";
import { createFileRoute, notFound } from "@tanstack/react-router";
import { useDeferredValue, useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { EmptyList } from "#/components/common/empty-list";
import { PageLayout } from "#/components/common/page-layout";
import { Skeleton } from "#/components/ui/skeleton";
import { VideoCard } from "#/features/videos/components/video-card";
import {
	DEFAULT_VIDEO_FILTERS,
	VideoFilterBar,
	type VideoFilters,
} from "#/features/videos/components/video-filter-bar";
import { HUB_PREVIEW_K } from "#/features/videos/constants";
import {
	useVideoList,
	videoListQueryOptions,
	videoPreviewQueryOptions,
} from "#/features/videos/queries";
import { loadOr404 } from "#/lib/route-loader";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/videos/categories/$categorySlug")({
	loader: async ({
		context: { queryClient },
		params: { categorySlug },
	}): Promise<VideoCategoryPreview> => {
		const categories = await loadOr404(
			queryClient,
			videoPreviewQueryOptions(HUB_PREVIEW_K),
		);
		const category = categories.find((c) => c.slug === categorySlug);
		if (!category) throw notFound();

		await queryClient.prefetchQuery(
			videoListQueryOptions({ categoryId: category.id, pageSize: 50 }),
		);

		return category;
	},
	pendingComponent: CategoryPending,
	component: VideoCategoryPage,
});

function CategoryPending() {
	return (
		<PageLayout>
			<div className="flex flex-col gap-6">
				<Skeleton className="h-10 w-48 rounded-xl" />
				<Skeleton className="h-24 w-full rounded-2xl" />
				<div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
					{[0, 1, 2, 3].map((card) => (
						<Skeleton key={card} className="aspect-[4/3] rounded-[24px]" />
					))}
				</div>
			</div>
		</PageLayout>
	);
}

function VideoCategoryPage() {
	const category = Route.useLoaderData();
	const [filters, setFilters] = useState<VideoFilters>(DEFAULT_VIDEO_FILTERS);
	const deferredSearch = useDeferredValue(filters.search.trim());

	useBreadcrumbs([
		{ label: m["nav.item.videos"](), to: APP_ROUTES.VIDEOS.LIST },
		{ label: category.name },
	]);

	const videoListQuery = useVideoList({
		categoryId: category.id,
		pageSize: 50,
		search: deferredSearch || undefined,
		cefrLevel: filters.level === "all" ? undefined : filters.level,
	});

	if (videoListQuery.isError) {
		throw videoListQuery.error;
	}

	const videos = videoListQuery.data?.items ?? [];
	const total = videoListQuery.data?.pagination?.total ?? category.videoCount;

	return (
		<PageLayout
			action={
				<BackButton
					to={APP_ROUTES.VIDEOS.LIST}
					label={m["videos.category.back"]()}
				/>
			}
		>
			<div className="flex flex-wrap items-end justify-between gap-3">
				<h1 className="text-[28px] leading-[36px] font-bold text-foreground md:text-[38px] md:leading-[46px]">
					{category.name}
				</h1>
				<span className="chip px-3 py-1 text-sm">
					{m["videos.category.total"]({ count: total })}
				</span>
			</div>

			<VideoFilterBar filters={filters} onChange={setFilters} />

			{videoListQuery.isPending ? (
				<div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
					{[0, 1, 2, 3].map((card) => (
						<Skeleton key={card} className="aspect-[4/3] rounded-[24px]" />
					))}
				</div>
			) : videos.length === 0 ? (
				<EmptyList
					title={m["videos.category.empty.title"]()}
					description={m["videos.category.empty.body"]()}
					onReset={() => setFilters(DEFAULT_VIDEO_FILTERS)}
				/>
			) : (
				<div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
					{videos.map((video) => (
						<VideoCard key={video.id} video={video} />
					))}
				</div>
			)}
		</PageLayout>
	);
}
