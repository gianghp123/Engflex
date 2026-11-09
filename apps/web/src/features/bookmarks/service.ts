import type {
	Bookmark,
	ContentType,
	SaveBookmark,
	SavedLibrary,
	UpdateSavedNote,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { ApiError, api, apiPaginated } from "#/lib/api";

export const savedLibraryKey = ["saved-library"] as const;

export function listBookmarks(type: ContentType) {
	return api<Bookmark[]>(API_ROUTES.BOOKMARKS.LIST, undefined, {
		query: { type },
	});
}

export function listStoryBookmarks() {
	return listBookmarks("story");
}
export function saveBookmark(payload: SaveBookmark) {
	return api<Bookmark>(API_ROUTES.BOOKMARKS.CREATE, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}
export async function unsaveStoryBookmark(id: string) {
	return unsaveBookmark("story", id);
}

export async function unsaveBookmark(type: ContentType, id: string) {
	try {
		await api<void>(API_ROUTES.BOOKMARKS.DELETE(type, id), {
			method: "DELETE",
		});
	} catch (error) {
		if (!(error instanceof ApiError && error.status === 404)) throw error;
	}
}

export function listSavedLibrary(query: {
	type: string;
	q: string;
	page: number;
}) {
	return apiPaginated<SavedLibrary>(API_ROUTES.BOOKMARKS.LIBRARY, undefined, {
		query,
	});
}

export function updateSavedNote(
	type: string,
	id: string,
	payload: UpdateSavedNote,
) {
	const path =
		type === "vocab"
			? API_ROUTES.USER_VOCABULARY.NOTE(id)
			: API_ROUTES.BOOKMARKS.NOTE(type, id);
	return api<void>(path, { method: "PATCH", body: JSON.stringify(payload) });
}
