import type { VideoCategoryPreview } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { ChevronRight } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";
import { VideoCard } from "./video-card";
export function VideoSection({ section }: { section: VideoCategoryPreview }) {
	const headingId = `video-section-${section.slug}`;
	return (
		<section aria-labelledby={headingId} className="flex flex-col gap-4">
			<div className="surface-card flex flex-wrap items-center justify-between gap-3 px-5 py-3">
				<div className="flex min-w-0 items-center gap-3">
					<span
						aria-hidden="true"
						className="h-7 w-1.5 shrink-0 rounded-full bg-primary shadow-[0_2px_0_#433095]"
					/>
					<h2
						id={headingId}
						className="truncate text-lg font-bold text-foreground md:text-xl"
					>
						{section.name}
					</h2>
					<span className="shrink-0 text-sm font-semibold text-muted-foreground">
						{m["videos.section.count"]({ count: section.videoCount })}
					</span>
				</div>
				<Button asChild variant="outline" className="btn btn-outline">
					<Link
						to={APP_ROUTES.VIDEOS.CATEGORY}
						params={{ categorySlug: section.slug }}
					>
						{m["videos.section.seeAll"]()}
						<ChevronRight data-icon="inline-end" />
					</Link>
				</Button>
			</div>
			<div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
				{section.videos.map((video) => (
					<VideoCard key={video.id} video={video} />
				))}
			</div>
		</section>
	);
}
