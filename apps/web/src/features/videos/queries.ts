import type { ListVideoExercises } from "@engflex/contracts";
import { useQuery } from "@tanstack/react-query";
import {
	getVideo,
	listCategoryPreviews,
	listVideos,
} from "#/features/videos/service";

export const videoKeys = {
	all: ["videos"] as const,
	preview: (previewK: number) =>
		[...videoKeys.all, "preview", previewK] as const,
	list: (params: Partial<ListVideoExercises>) =>
		[...videoKeys.all, "list", params] as const,
	detail: (id: string) => [...videoKeys.all, "detail", id] as const,
};

export function videoPreviewQueryOptions(previewK: number) {
	return {
		queryKey: videoKeys.preview(previewK),
		queryFn: () => listCategoryPreviews(previewK),
	};
}

export function videoListQueryOptions(params: Partial<ListVideoExercises>) {
	return {
		queryKey: videoKeys.list(params),
		queryFn: () => listVideos(params),
	};
}

export function videoDetailQueryOptions(id: string) {
	return {
		queryKey: videoKeys.detail(id),
		queryFn: () => getVideo(id),
	};
}

export function useVideoPreview(previewK: number) {
	return useQuery(videoPreviewQueryOptions(previewK));
}

export function useVideoList(params: Partial<ListVideoExercises>) {
	return useQuery(videoListQueryOptions(params));
}

export function useVideo(id: string) {
	return useQuery(videoDetailQueryOptions(id));
}
