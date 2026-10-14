import { createFileRoute, Link, notFound } from "@tanstack/react-router";
import { ArrowLeft } from "lucide-react";
import { useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { EmptyList } from "#/components/common/empty-list";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { VideoCard } from "#/features/videos/components/video-card";
import {
	DEFAULT_VIDEO_FILTERS,
	VideoFilterBar,
	type VideoFilters,
} from "#/features/videos/components/video-filter-bar";
import { getVideoSectionBySlug } from "#/features/videos/fixtures";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/videos/categories/$categorySlug")({
	loader: ({ params: { categorySlug } }) => {
		const section = getVideoSectionBySlug(categorySlug);
		if (!section) throw notFound();
		return section;
	},
	component: VideoCategoryPage,
});

function VideoCategoryPage() {
	const section = Route.useLoaderData();
	const [filters, setFilters] = useState<VideoFilters>(DEFAULT_VIDEO_FILTERS);

	useBreadcrumbs([
		{ label: m["nav.item.videos"](), to: APP_ROUTES.VIDEOS.LIST },
		{ label: section.category.name },
	]);

	// Fixture-only filtering; the API's `search` / `cefrLevel` params replace it.
	const query = filters.search.trim().toLowerCase();
	const videos = section.videos.filter(
		(video) =>
			(filters.level === "all" || video.cefrLevel === filters.level) &&
			(query === "" || video.title.toLowerCase().includes(query)),
	);

	return (
		<PageLayout
			action={
				<Button asChild variant="ghost" size="sm" className="w-fit">
					<Link to={APP_ROUTES.VIDEOS.LIST}>
						<ArrowLeft data-icon="inline-start" />
						{m["videos.category.back"]()}
					</Link>
				</Button>
			}
		>
			<div className="flex flex-wrap items-end justify-between gap-3">
				<h1 className="text-[28px] leading-[36px] font-bold text-foreground md:text-[38px] md:leading-[46px]">
					{section.category.name}
				</h1>
				<span className="chip px-3 py-1 text-sm">
					{m["videos.category.total"]({ count: section.total })}
				</span>
			</div>

			<VideoFilterBar filters={filters} onChange={setFilters} />

			{videos.length === 0 ? (
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
