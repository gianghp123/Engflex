import { createFileRoute } from "@tanstack/react-router";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { PageLayout } from "#/components/common/page-layout";
import { VideoSection } from "#/features/videos/components/video-section";
import { HUB_PREVIEW_K } from "#/features/videos/constants";
import { VIDEO_SECTIONS } from "#/features/videos/fixtures";
import { useVideoPreview } from "#/features/videos/queries";
import { m } from "#/paraglide/messages";
export const Route = createFileRoute("/_app/videos/")({
	component: VideosPage,
});

function VideosPage() {
	const probe = useVideoPreview(HUB_PREVIEW_K);
	console.log("[probe] preview", probe.status, probe.data);
	useBreadcrumbs([{ label: m["nav.item.videos"]() }]);
	return (
		<PageLayout>
			{VIDEO_SECTIONS.map((section) => (
				<VideoSection key={section.category.id} section={section} />
			))}
		</PageLayout>
	);
}
