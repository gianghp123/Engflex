import type {
	Pagination,
	UserVocabularyState,
	VocabularyDeckDetail,
	VocabularyDeckItemResponse,
	VocabularyItem,
	VocabularyStats,
} from "@engflex/contracts";
import { MaxPageSize } from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api";
import type { AddWordInput } from "./types";

const JSON_HEADERS = { "content-type": "application/json" };

/** My saved words, with per-user state. One page — the hub filters client-side. */
export function listMyVocabulary(): Promise<{
	items: VocabularyItem[];
	pagination: Pagination;
}> {
	return apiPage<VocabularyItem>(
		API_ROUTES.VOCABULARY.LIBRARY.LIST,
		{ method: "GET" },
		{ query: { pageSize: MaxPageSize } },
	);
}

/**
 * Library stats. `dayStart` is the reader's local midnight (RFC3339), so
 * "saved today" is their day; omitted, the API counts the UTC day.
 */
export function getVocabularyStats(
	dayStart?: string,
): Promise<VocabularyStats> {
	return api<VocabularyStats>(
		API_ROUTES.VOCABULARY.LIBRARY.STATS,
		{ method: "GET" },
		dayStart ? { query: { dayStart } } : undefined,
	);
}

export function getVocabularyItem(id: string): Promise<VocabularyItem> {
	return api<VocabularyItem>(API_ROUTES.VOCABULARY.ITEMS.BY_ID(id), {
		method: "GET",
	});
}

/** Create-or-enrol a word the user typed. Idempotent by term on the server. */
export function addVocabularyWord(
	input: AddWordInput,
): Promise<VocabularyItem> {
	return api<VocabularyItem>(API_ROUTES.VOCABULARY.LIBRARY.ADD, {
		method: "POST",
		headers: JSON_HEADERS,
		body: JSON.stringify({
			term: input.term,
			definition: input.definition,
			cefr: input.cefr,
			partOfSpeech: input.partOfSpeech,
			details: {
				contexts: input.example
					? [{ label: "Example", quote: input.example }]
					: [],
			},
		}),
	});
}

export function patchUserVocabulary(
	itemId: string,
	patch: { mastered?: boolean; note?: string },
): Promise<UserVocabularyState> {
	return api<UserVocabularyState>(API_ROUTES.VOCABULARY.LIBRARY.BY_ID(itemId), {
		method: "PATCH",
		headers: JSON_HEADERS,
		body: JSON.stringify(patch),
	});
}

/** My decks (private), each with up to previewK preview words attached. */
export function listDecks(previewK: number): Promise<{
	items: VocabularyDeckDetail[];
	pagination: Pagination;
}> {
	return apiPage<VocabularyDeckDetail>(
		API_ROUTES.VOCABULARY.DECKS.LIST,
		{ method: "GET" },
		{ query: { previewK, pageSize: MaxPageSize } },
	);
}

export function getDeck(id: string): Promise<VocabularyDeckDetail> {
	return api<VocabularyDeckDetail>(API_ROUTES.VOCABULARY.DECKS.BY_ID(id), {
		method: "GET",
	});
}

export function createDeck(input: {
	name: string;
	description: string;
}): Promise<VocabularyDeckDetail> {
	return api<VocabularyDeckDetail>(API_ROUTES.VOCABULARY.DECKS.LIST, {
		method: "POST",
		headers: JSON_HEADERS,
		body: JSON.stringify(input),
	});
}

export function updateDeck(
	id: string,
	patch: { name?: string; description?: string },
): Promise<VocabularyDeckDetail> {
	return api<VocabularyDeckDetail>(API_ROUTES.VOCABULARY.DECKS.BY_ID(id), {
		method: "PUT",
		headers: JSON_HEADERS,
		body: JSON.stringify(patch),
	});
}

export function deleteDeck(id: string): Promise<null> {
	return api<null>(API_ROUTES.VOCABULARY.DECKS.BY_ID(id), { method: "DELETE" });
}

/**
 * Add a word straight into a deck: the server find-or-creates the caller's own
 * private entry and then creates the membership, so "in a deck" implies "in my
 * words". The curated pool is never consulted (the API's design decision 1).
 */
export function addWordToDeck(
	deckId: string,
	input: AddWordInput,
): Promise<VocabularyItem> {
	return api<VocabularyItem>(API_ROUTES.VOCABULARY.LIBRARY.ADD, {
		method: "POST",
		headers: JSON_HEADERS,
		body: JSON.stringify({
			term: input.term,
			definition: input.definition,
			cefr: input.cefr,
			partOfSpeech: input.partOfSpeech,
			deckId,
			details: {
				contexts: input.example
					? [{ label: "Example", quote: input.example }]
					: [],
			},
		}),
	});
}

/** Add an already-saved word to a deck (enrolment untouched). */
export function addExistingWordToDeck(
	deckId: string,
	itemId: string,
): Promise<VocabularyDeckItemResponse> {
	return api<VocabularyDeckItemResponse>(API_ROUTES.VOCABULARY.DECKS.ITEMS, {
		method: "POST",
		headers: JSON_HEADERS,
		body: JSON.stringify({ deckId, itemId }),
	});
}

/** Remove a word from a deck; the word stays in the library. */
export function removeDeckItem(deckItemId: string): Promise<null> {
	return api<null>(API_ROUTES.VOCABULARY.DECKS.ITEM_BY_ID(deckItemId), {
		method: "DELETE",
	});
}
