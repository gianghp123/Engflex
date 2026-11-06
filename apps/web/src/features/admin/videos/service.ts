import type {
	CreateVideoCategory,
	CreateVideoExercise,
	CreateVideoTranscript,
	Pagination,
	UpdateVideoCategory,
	UpdateVideoExercise,
	UpdateVideoTranscript,
	VideoCategoryResponse,
	VideoExerciseResponse,
	VideoTranscriptResponse,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api";

export function listVideoTranscripts(videoId: string) {
	return api<VideoTranscriptResponse[]>(
		API_ROUTES.VIDEO_TRANSCRIPTS.BY_VIDEO(videoId),
	);
}
export function createVideoTranscript(payload: CreateVideoTranscript) {
	return api<VideoTranscriptResponse>(API_ROUTES.VIDEO_TRANSCRIPTS.LIST, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}
export function updateVideoTranscript(
	id: string,
	payload: UpdateVideoTranscript,
) {
	return api<VideoTranscriptResponse>(API_ROUTES.VIDEO_TRANSCRIPTS.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}
export function deleteVideoTranscript(id: string) {
	return api<void>(API_ROUTES.VIDEO_TRANSCRIPTS.BY_ID(id), {
		method: "DELETE",
	});
}

export function listVideoExercises(
	page: number,
	pageSize: number,
): Promise<{ items: VideoExerciseResponse[]; pagination: Pagination }> {
	return apiPage<VideoExerciseResponse>(
		API_ROUTES.VIDEO_EXERCISES.LIST,
		undefined,
		{ query: { page, pageSize } },
	);
}

export function createVideoExercise(
	payload: CreateVideoExercise,
): Promise<VideoExerciseResponse> {
	return api<VideoExerciseResponse>(API_ROUTES.VIDEO_EXERCISES.LIST, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updateVideoExercise(
	id: string,
	payload: UpdateVideoExercise,
): Promise<VideoExerciseResponse> {
	return api<VideoExerciseResponse>(API_ROUTES.VIDEO_EXERCISES.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deleteVideoExercise(id: string): Promise<void> {
	return api<void>(API_ROUTES.VIDEO_EXERCISES.BY_ID(id), { method: "DELETE" });
}

export function listVideoCategories(
	page: number,
	pageSize: number,
): Promise<{ items: VideoCategoryResponse[]; pagination: Pagination }> {
	return apiPage<VideoCategoryResponse>(
		API_ROUTES.VIDEO_CATEGORIES.LIST,
		undefined,
		{ query: { page, pageSize } },
	);
}

export function createVideoCategory(
	payload: CreateVideoCategory,
): Promise<VideoCategoryResponse> {
	return api<VideoCategoryResponse>(API_ROUTES.VIDEO_CATEGORIES.LIST, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updateVideoCategory(
	id: string,
	payload: UpdateVideoCategory,
): Promise<VideoCategoryResponse> {
	return api<VideoCategoryResponse>(API_ROUTES.VIDEO_CATEGORIES.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deleteVideoCategory(id: string): Promise<void> {
	return api<void>(API_ROUTES.VIDEO_CATEGORIES.BY_ID(id), { method: "DELETE" });
}
