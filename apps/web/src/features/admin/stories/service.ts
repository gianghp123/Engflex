import type {
	CreateStory,
	Pagination,
	StoryItem,
	UpdateStory,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api";

export interface AdminStoryFilters {
	search?: string;
	cefr?: string;
	topic?: string;
}

export function listAdminStories(
	page: number,
	pageSize: number,
	filters: AdminStoryFilters = {},
): Promise<{ items: StoryItem[]; pagination: Pagination }> {
	const query: Record<string, string | number> = { page, pageSize };
	if (filters.search?.trim()) query.search = filters.search.trim();
	if (filters.cefr) query.cefr = filters.cefr;
	if (filters.topic) query.topic = filters.topic;
	return apiPage<StoryItem>(API_ROUTES.STORIES.LIST, undefined, { query });
}

export function createStory(payload: CreateStory): Promise<StoryItem> {
	return api<StoryItem>(API_ROUTES.STORIES.LIST, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updateStory(
	id: string,
	payload: UpdateStory,
): Promise<StoryItem> {
	return api<StoryItem>(API_ROUTES.STORIES.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deleteStory(id: string): Promise<void> {
	return api<void>(API_ROUTES.STORIES.BY_ID(id), { method: "DELETE" });
}

export function uploadStoryAsset(file: File): Promise<{ url: string }> {
	const form = new FormData();
	form.append("file", file);
	return api<{ url: string }>(API_ROUTES.STORIES.UPLOAD, {
		method: "POST",
		body: form,
	});
}
