import { createFileRoute } from "@tanstack/react-router";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { PageLayout } from "#/components/common/page-layout";
import { VideoSection } from "#/features/videos/components/video-section";
import { VIDEO_SECTIONS } from "#/features/videos/fixtures";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/videos/")({
	component: VideosPage,
});

function VideosPage() {
	useBreadcrumbs([{ label: m["nav.item.videos"]() }]);
	return (
		<PageLayout>
			{VIDEO_SECTIONS.map((section) => (
				<VideoSection key={section.category.id} section={section} />
			))}
		</PageLayout>
	);
}
