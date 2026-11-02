import type {
	CreateVocabularyItem,
	Pagination,
	SavedVocabulary,
	SaveUserVocabulary,
	UserVocabularyState,
	VocabularyItem,
	VocabularyItemSummary,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api";

export function listVocabularyItems(
	page: number,
	pageSize: number,
): Promise<{ items: VocabularyItemSummary[]; pagination: Pagination }> {
	return apiPage<VocabularyItemSummary>(
		API_ROUTES.VOCABULARY_ITEMS.LIST,
		undefined,
		{
			query: { page, pageSize },
		},
	);
}

export function getVocabularyItem(id: string): Promise<VocabularyItem> {
	return api<VocabularyItem>(API_ROUTES.VOCABULARY_ITEMS.BY_ID(id));
}

export function createVocabularyItem(
	payload: Omit<CreateVocabularyItem, "isPremium">,
): Promise<VocabularyItem> {
	return api<VocabularyItem>(API_ROUTES.VOCABULARY_ITEMS.LIST, {
		method: "POST",
		body: JSON.stringify({ ...payload, isPremium: false }),
	});
}

export function listUserVocabulary(
	page: number,
	pageSize: number,
): Promise<{ items: SavedVocabulary[]; pagination: Pagination }> {
	return apiPage<SavedVocabulary>(API_ROUTES.USER_VOCABULARY.LIST, undefined, {
		query: { page, pageSize },
	});
}

export function saveUserVocabulary(
	payload: SaveUserVocabulary,
): Promise<UserVocabularyState> {
	return api<UserVocabularyState>(API_ROUTES.USER_VOCABULARY.LIST, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function unsaveUserVocabulary(itemId: string): Promise<void> {
	return api<void>(API_ROUTES.USER_VOCABULARY.BY_ITEM_ID(itemId), {
		method: "DELETE",
	});
}
