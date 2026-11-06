import type { Bookmark, SaveBookmark } from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { ApiError, api } from "#/lib/api";

export function listStoryBookmarks() {
	return api<Bookmark[]>(API_ROUTES.BOOKMARKS.LIST, undefined, {
		query: { type: "story" },
	});
}
export function saveBookmark(payload: SaveBookmark) {
	return api<Bookmark>(API_ROUTES.BOOKMARKS.CREATE, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}
export async function unsaveStoryBookmark(id: string) {
	try {
		await api<void>(API_ROUTES.BOOKMARKS.DELETE("story", id), {
			method: "DELETE",
		});
	} catch (error) {
		if (!(error instanceof ApiError && error.status === 404)) throw error;
	}
}
