import { ChevronRight } from "lucide-react";
import { Button } from "#/components/ui/button";
import type { VideoSection as VideoSectionData } from "#/features/videos/types";
import { m } from "#/paraglide/messages";
import { VideoCard } from "./video-card";

export function VideoSection({ section }: { section: VideoSectionData }) {
	const headingId = `video-section-${section.category.slug}`;
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
						{section.category.name}
					</h2>
					<span className="shrink-0 text-sm font-semibold text-muted-foreground">
						{m["videos.section.count"]({ count: section.total })}
					</span>
				</div>
				<Button type="button" variant="outline" className="btn btn-outline">
					{m["videos.section.seeAll"]()}
					<ChevronRight data-icon="inline-end" />
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
