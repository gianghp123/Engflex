import type {
	ListVideoExercises,
	SaveVideoProgress,
	StartVideoAttempt,
	VideoAttemptResponse,
	VideoCategoryPreview,
	VideoExerciseResponse,
	VideoPronunciationResult,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { ApiError, api, apiPage } from "#/lib/api/api";

export function listCategoryPreviews(
	previewK: number,
): Promise<VideoCategoryPreview[]> {
	return api<VideoCategoryPreview[]>(
		API_ROUTES.VIDEOS.CATEGORIES_PREVIEW,
		undefined,
		{ query: { previewK } },
	);
}

export function listVideos(params: Partial<ListVideoExercises>) {
	const query: Record<string, string | number> = {};
	if (params.page) query.page = params.page;
	if (params.pageSize) query.pageSize = params.pageSize;
	if (params.search?.trim()) query.search = params.search.trim();
	if (params.categoryId) query.categoryId = params.categoryId;
	if (params.cefrLevel) query.cefrLevel = params.cefrLevel;
	return apiPage<VideoExerciseResponse>(API_ROUTES.VIDEOS.LIST, undefined, {
		query,
	});
}

export function getVideo(id: string): Promise<VideoExerciseResponse> {
	return api<VideoExerciseResponse>(API_ROUTES.VIDEOS.BY_ID(id));
}

export async function getOpenDictationAttempt(videoId: string) {
	try {
		return await api<VideoAttemptResponse>(
			API_ROUTES.VIDEOS.ATTEMPTS.OPEN(videoId),
			undefined,
			{ query: { type: "dictation" } },
		);
	} catch (error) {
		if (error instanceof ApiError && error.status === 404) return null;
		throw error;
	}
}

export function startDictationAttempt(videoId: string, level: string) {
	return api<VideoAttemptResponse>(API_ROUTES.VIDEOS.ATTEMPTS.START(videoId), {
		method: "POST",
		body: JSON.stringify({
			type: "dictation",
			level,
		} satisfies StartVideoAttempt),
	});
}

export async function getOpenShadowingAttempt(videoId: string) {
	try {
		return await api<VideoAttemptResponse>(
			API_ROUTES.VIDEOS.ATTEMPTS.OPEN(videoId),
			undefined,
			{ query: { type: "shadowing" } },
		);
	} catch (error) {
		if (error instanceof ApiError && error.status === 404) return null;
		throw error;
	}
}

export function startShadowingAttempt(videoId: string) {
	return api<VideoAttemptResponse>(API_ROUTES.VIDEOS.ATTEMPTS.START(videoId), {
		method: "POST",
		body: JSON.stringify({
			type: "shadowing",
		} satisfies StartVideoAttempt),
	});
}

// saveDictationProgress/completeDictationAttempt below operate on
// videoId+attemptId only — the backend attempt is type-agnostic (see
// StartVideoAttempt.type), so shadowing reuses them directly rather than
// duplicating identical request plumbing under a new name.
export function saveDictationProgress(
	videoId: string,
	attemptId: string,
	progress: SaveVideoProgress,
) {
	return api<VideoAttemptResponse>(
		API_ROUTES.VIDEOS.ATTEMPTS.SAVE_PROGRESS(videoId, attemptId),
		{
			method: "PUT",
			body: JSON.stringify(progress),
		},
	);
}

export function completeDictationAttempt(videoId: string, attemptId: string) {
	return api<VideoAttemptResponse>(
		API_ROUTES.VIDEOS.ATTEMPTS.COMPLETE(videoId, attemptId),
		{
			method: "POST",
			body: JSON.stringify({}),
		},
	);
}

export function pronounceShadowing(
	videoId: string,
	attemptId: string,
	transcriptId: string,
	audio: Blob,
	mime: string,
): Promise<VideoPronunciationResult> {
	const form = new FormData();
	form.append("transcriptId", transcriptId);
	form.append(
		"audio",
		audio,
		mime.includes("wav") ? "attempt.wav" : "attempt.webm",
	);
	return api<VideoPronunciationResult>(
		API_ROUTES.VIDEOS.ATTEMPTS.PRONOUNCE(videoId, attemptId),
		{ method: "POST", body: form },
	);
}
