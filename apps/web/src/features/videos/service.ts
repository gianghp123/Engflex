import type {
	ListVideoExercises,
	VideoCategoryPreview,
	VideoExerciseResponse,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api/api";

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
