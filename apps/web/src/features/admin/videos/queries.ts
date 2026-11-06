import { useAuth } from "@clerk/tanstack-react-start";
import type {
	CreateVideoCategory,
	CreateVideoExercise,
	CreateVideoTranscript,
	UpdateVideoCategory,
	UpdateVideoExercise,
	UpdateVideoTranscript,
} from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	createVideoCategory,
	createVideoExercise,
	createVideoTranscript,
	deleteVideoCategory,
	deleteVideoExercise,
	deleteVideoTranscript,
	listVideoCategories,
	listVideoExercises,
	listVideoTranscripts,
	updateVideoCategory,
	updateVideoExercise,
	updateVideoTranscript,
} from "#/features/admin/videos/service";
import { videoKeys } from "#/features/videos/queries";
import { ApiError } from "#/lib/api";

export function useAdminVideoTranscripts(videoId: string) {
	const { isLoaded } = useAuth();
	return useQuery({
		queryKey: ["admin", "video-transcripts", videoId],
		queryFn: () => listVideoTranscripts(videoId),
		enabled: isLoaded && !!videoId,
		retry: (count, error) =>
			!(error instanceof ApiError && [401, 403, 404].includes(error.status)) &&
			count < 3,
	});
}
function useTranscriptInvalidation(videoId: string) {
	const client = useQueryClient();
	return () =>
		Promise.all([
			client.invalidateQueries({
				queryKey: ["admin", "video-transcripts", videoId],
			}),
			client.invalidateQueries({ queryKey: ["admin", "video-exercises"] }),
			client.invalidateQueries({ queryKey: videoKeys.all }),
		]);
}
export function useCreateVideoTranscript(videoId: string) {
	const invalidate = useTranscriptInvalidation(videoId);
	return useMutation({
		mutationFn: (payload: CreateVideoTranscript) =>
			createVideoTranscript(payload),
		onSuccess: invalidate,
	});
}
export function useUpdateVideoTranscript(videoId: string) {
	const invalidate = useTranscriptInvalidation(videoId);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateVideoTranscript;
		}) => updateVideoTranscript(id, payload),
		onSuccess: invalidate,
	});
}
export function useDeleteVideoTranscript(videoId: string) {
	const invalidate = useTranscriptInvalidation(videoId);
	return useMutation({
		mutationFn: deleteVideoTranscript,
		onSuccess: invalidate,
	});
}

const keys = {
	exercises: (page: number, pageSize: number) =>
		["admin", "video-exercises", page, pageSize] as const,
	categories: (page: number, pageSize: number) =>
		["admin", "video-categories", page, pageSize] as const,
};

function useInvalidate(key: readonly unknown[]) {
	const queryClient = useQueryClient();
	return () => queryClient.invalidateQueries({ queryKey: key });
}

export function useAdminVideoExercises(page: number, pageSize: number) {
	return useQuery({
		queryKey: keys.exercises(page, pageSize),
		queryFn: () => listVideoExercises(page, pageSize),
	});
}

export function useCreateVideoExercise() {
	const invalidate = useInvalidate(["admin", "video-exercises"]);
	return useMutation({
		mutationFn: (payload: CreateVideoExercise) => createVideoExercise(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateVideoExercise() {
	const invalidate = useInvalidate(["admin", "video-exercises"]);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateVideoExercise;
		}) => updateVideoExercise(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteVideoExercise() {
	const invalidate = useInvalidate(["admin", "video-exercises"]);
	return useMutation({
		mutationFn: (id: string) => deleteVideoExercise(id),
		onSuccess: invalidate,
	});
}

export function useAdminVideoCategories(page: number, pageSize: number) {
	return useQuery({
		queryKey: keys.categories(page, pageSize),
		queryFn: () => listVideoCategories(page, pageSize),
	});
}

export function useCreateVideoCategory() {
	const invalidate = useInvalidate(["admin", "video-categories"]);
	return useMutation({
		mutationFn: (payload: CreateVideoCategory) => createVideoCategory(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateVideoCategory() {
	const invalidate = useInvalidate(["admin", "video-categories"]);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateVideoCategory;
		}) => updateVideoCategory(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteVideoCategory() {
	const invalidate = useInvalidate(["admin", "video-categories"]);
	return useMutation({
		mutationFn: (id: string) => deleteVideoCategory(id),
		onSuccess: invalidate,
	});
}
