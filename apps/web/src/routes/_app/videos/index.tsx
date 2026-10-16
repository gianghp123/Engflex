import { createFileRoute } from "@tanstack/react-router";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { VideoHubEmpty } from "#/features/videos/components/video-hub-empty";
import { VideoHubSkeleton } from "#/features/videos/components/video-hub-skeleton";
import { VideoSection } from "#/features/videos/components/video-section";
import { HUB_PREVIEW_K } from "#/features/videos/constants";
import {
	useVideoPreview,
	videoPreviewQueryOptions,
} from "#/features/videos/queries";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/videos/")({
	loader: ({ context: { queryClient } }) =>
		queryClient.query(videoPreviewQueryOptions(HUB_PREVIEW_K)),
	pendingComponent: VideoPending,
	component: VideosPage,
});

function VideoPending() {
	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="open" size={84} />,
				title: m["videos.hub.title"](),
				description: m["videos.hub.body"](),
			}}
		>
			<VideoHubSkeleton />
		</PageLayout>
	);
}

function VideosPage() {
	const previewQuery = useVideoPreview(HUB_PREVIEW_K);
	if (previewQuery.isError) {
		throw previewQuery.error;
	}
	const sections = previewQuery.data ?? [];

	useBreadcrumbs([{ label: m["nav.item.videos"]() }]);
	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="open" size={84} />,
				title: m["videos.hub.title"](),
				description: m["videos.hub.body"](),
			}}
		>
			{sections.length === 0 ? (
				<VideoHubEmpty />
			) : (
				sections.map((section) => (
					<VideoSection key={section.id} section={section} />
				))
			)}
		</PageLayout>
	);
}
