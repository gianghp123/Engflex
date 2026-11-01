import type {
	CreateVocabularyCategory,
	CreateVocabularyDeck,
	CreateVocabularyDeckItem,
	CreateVocabularyItem,
	Pagination,
	UpdateVocabularyCategory,
	UpdateVocabularyDeck,
	UpdateVocabularyDeckItem,
	UpdateVocabularyItem,
	VocabularyCategoryResponse,
	VocabularyDeckDetail,
	VocabularyDeckItemResponse,
	VocabularyItem,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api";

export function listVocabularyItems(
	page: number,
	pageSize: number,
): Promise<{ items: VocabularyItem[]; pagination: Pagination }> {
	return apiPage<VocabularyItem>(API_ROUTES.VOCABULARY_ITEMS.LIST, undefined, {
		query: { page, pageSize },
	});
}

export function createVocabularyItem(
	payload: CreateVocabularyItem,
): Promise<VocabularyItem> {
	return api<VocabularyItem>(API_ROUTES.ADMIN.VOCABULARY_ITEMS, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updateVocabularyItem(
	id: string,
	payload: UpdateVocabularyItem,
): Promise<VocabularyItem> {
	return api<VocabularyItem>(API_ROUTES.VOCABULARY_ITEMS.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deleteVocabularyItem(id: string): Promise<void> {
	return api<void>(API_ROUTES.VOCABULARY_ITEMS.BY_ID(id), {
		method: "DELETE",
	});
}

export function listVocabularyDecks(
	page: number,
	pageSize: number,
): Promise<{ items: VocabularyDeckDetail[]; pagination: Pagination }> {
	return apiPage<VocabularyDeckDetail>(
		API_ROUTES.VOCABULARY_DECKS.LIST,
		undefined,
		{ query: { page, pageSize } },
	);
}

export function createVocabularyDeck(
	payload: CreateVocabularyDeck,
): Promise<VocabularyDeckDetail> {
	return api<VocabularyDeckDetail>(API_ROUTES.ADMIN.VOCABULARY_DECKS, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updateVocabularyDeck(
	id: string,
	payload: UpdateVocabularyDeck,
): Promise<VocabularyDeckDetail> {
	return api<VocabularyDeckDetail>(API_ROUTES.VOCABULARY_DECKS.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deleteVocabularyDeck(id: string): Promise<void> {
	return api<void>(API_ROUTES.VOCABULARY_DECKS.BY_ID(id), {
		method: "DELETE",
	});
}

export function listVocabularyDeckItems(
	deckId: string,
): Promise<VocabularyDeckItemResponse[]> {
	return api<VocabularyDeckItemResponse[]>(
		API_ROUTES.VOCABULARY_DECK_ITEMS.LIST,
		undefined,
		{ query: { deckId } },
	);
}

export function createVocabularyDeckItem(
	payload: CreateVocabularyDeckItem,
): Promise<VocabularyDeckItemResponse> {
	return api<VocabularyDeckItemResponse>(
		API_ROUTES.VOCABULARY_DECK_ITEMS.LIST,
		{
			method: "POST",
			body: JSON.stringify(payload),
		},
	);
}

export function updateVocabularyDeckItem(
	id: string,
	payload: UpdateVocabularyDeckItem,
): Promise<VocabularyDeckItemResponse> {
	return api<VocabularyDeckItemResponse>(
		API_ROUTES.VOCABULARY_DECK_ITEMS.BY_ID(id),
		{ method: "PUT", body: JSON.stringify(payload) },
	);
}

export function deleteVocabularyDeckItem(id: string): Promise<void> {
	return api<void>(API_ROUTES.VOCABULARY_DECK_ITEMS.BY_ID(id), {
		method: "DELETE",
	});
}

export function listVocabularyCategories(
	page: number,
	pageSize: number,
): Promise<{ items: VocabularyCategoryResponse[]; pagination: Pagination }> {
	return apiPage<VocabularyCategoryResponse>(
		API_ROUTES.VOCABULARY_CATEGORIES.LIST,
		undefined,
		{ query: { page, pageSize } },
	);
}

export function createVocabularyCategory(
	payload: CreateVocabularyCategory,
): Promise<VocabularyCategoryResponse> {
	return api<VocabularyCategoryResponse>(
		API_ROUTES.VOCABULARY_CATEGORIES.LIST,
		{
			method: "POST",
			body: JSON.stringify(payload),
		},
	);
}

export function updateVocabularyCategory(
	id: string,
	payload: UpdateVocabularyCategory,
): Promise<VocabularyCategoryResponse> {
	return api<VocabularyCategoryResponse>(
		API_ROUTES.VOCABULARY_CATEGORIES.BY_ID(id),
		{ method: "PUT", body: JSON.stringify(payload) },
	);
}

export function deleteVocabularyCategory(id: string): Promise<void> {
	return api<void>(API_ROUTES.VOCABULARY_CATEGORIES.BY_ID(id), {
		method: "DELETE",
	});
}
