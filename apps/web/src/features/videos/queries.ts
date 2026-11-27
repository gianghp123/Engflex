import type { ListVideoExercises } from "@engflex/contracts";
import { useMutation, useQuery } from "@tanstack/react-query";
import {
	getOpenDictationAttempt,
	getOpenShadowingAttempt,
	getVideo,
	listCategoryPreviews,
	listVideos,
	pronounceShadowing,
} from "#/features/videos/service";

export const videoKeys = {
	all: ["videos"] as const,
	preview: (previewK: number) =>
		[...videoKeys.all, "preview", previewK] as const,
	list: (params: Partial<ListVideoExercises>) =>
		[...videoKeys.all, "list", params] as const,
	detail: (id: string) => [...videoKeys.all, "detail", id] as const,
	openDictation: (id: string) =>
		[...videoKeys.all, "dictation", "open", id] as const,
	openShadowing: (id: string) =>
		[...videoKeys.all, "shadowing", "open", id] as const,
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

export function useOpenDictationAttempt(id: string, enabled: boolean) {
	return useQuery({
		queryKey: videoKeys.openDictation(id),
		queryFn: () => getOpenDictationAttempt(id),
		enabled,
		refetchOnMount: "always",
		refetchOnWindowFocus: false,
	});
}

export function useOpenShadowingAttempt(id: string, enabled: boolean) {
	return useQuery({
		queryKey: videoKeys.openShadowing(id),
		queryFn: () => getOpenShadowingAttempt(id),
		enabled,
		refetchOnMount: "always",
		refetchOnWindowFocus: false,
	});
}

export function usePronounceShadowing() {
	return useMutation({
		mutationFn: (input: {
			videoId: string;
			attemptId: string;
			transcriptId: string;
			audio: Blob;
			mime: string;
		}) =>
			pronounceShadowing(
				input.videoId,
				input.attemptId,
				input.transcriptId,
				input.audio,
				input.mime,
			),
	});
}
