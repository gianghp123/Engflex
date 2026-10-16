import { MaxPageSize } from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	addExistingWordToDeck,
	addVocabularyWord,
	addWordToDeck,
	createDeck,
	deleteDeck,
	generateWordDraft,
	getDeck,
	getVocabularyItem,
	getVocabularyStats,
	listDecks,
	listMyVocabulary,
	patchUserVocabulary,
	removeDeckItem,
	updateDeck,
} from "./service";
import type { AddWordInput, VocabFilters } from "./types";

export const vocabularyKeys = {
	library: ["vocabulary", "library"] as const,
	// Searches nest under the library prefix, so library invalidations reach
	// them without a second invalidate call.
	librarySearch: (search: string) =>
		["vocabulary", "library", "search", search] as const,
	// The hub's numbered pages. Filters are part of the identity — a chip
	// change is a different list, not a refetch of the old one — and the key
	// still nests under `library`, so library invalidations reach it.
	libraryPage: (filters: VocabFilters, page: number) =>
		[
			"vocabulary",
			"library",
			"page",
			page,
			filters.search,
			filters.sort,
			filters.source,
			filters.cefr,
		] as const,
	// The stats response depends on the reader's local day, so that day is part
	// of the key — a new day is a different query, not a stale cache hit.
	stats: (dayStart?: string) =>
		dayStart
			? (["vocabulary", "stats", dayStart] as const)
			: (["vocabulary", "stats"] as const),
	item: (id: string) => ["vocabulary", "item", id] as const,
	decks: (previewK: number) => ["vocabulary", "decks", previewK] as const,
	deck: (id: string) => ["vocabulary", "deck", id] as const,
};

/** Every stats read shares this prefix, so one invalidate covers every day. */
const statsPrefix = ["vocabulary", "stats"] as const;

/** Every deck read shares this prefix, so one invalidate covers list + detail. */
const decksPrefix = ["vocabulary", "decks"] as const;
const deckPrefix = ["vocabulary", "deck"] as const;

/** Cards per page on the All-words grid: a full row on the 2- and 3-col grids. */
const WORDS_PAGE_SIZE = 24;

export function decksQueryOptions(previewK: number) {
	return {
		queryKey: vocabularyKeys.decks(previewK),
		queryFn: () => listDecks(previewK),
	};
}

export function deckQueryOptions(id: string) {
	return { queryKey: vocabularyKeys.deck(id), queryFn: () => getDeck(id) };
}

export function useDecks(previewK: number) {
	return useQuery(decksQueryOptions(previewK));
}

export function useDeck(id: string) {
	return useQuery(deckQueryOptions(id));
}

export function useCreateDeck() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: createDeck,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: decksPrefix });
		},
	});
}

export function useUpdateDeck(id: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (patch: { name?: string; description?: string }) =>
			updateDeck(id, patch),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: deckPrefix });
			queryClient.invalidateQueries({ queryKey: decksPrefix });
		},
	});
}

export function useDeleteDeck() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: deleteDeck,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: decksPrefix });
		},
	});
}

export function vocabularyStatsQueryOptions(dayStart?: string) {
	return {
		queryKey: vocabularyKeys.stats(dayStart),
		queryFn: () => getVocabularyStats(dayStart),
	};
}

export function vocabularyItemQueryOptions(id: string) {
	return {
		queryKey: vocabularyKeys.item(id),
		queryFn: () => getVocabularyItem(id),
	};
}

/**
 * One page of my words matching `%search%` over term or definition. The deck
 * picker submits it explicitly; the add-word form debounces the typed term
 * through it. `null` disables the query — nothing is fetched until a needle
 * exists.
 */
export function useLibrarySearch(search: string | null) {
	return useQuery({
		queryKey: vocabularyKeys.librarySearch(search ?? ""),
		queryFn: () =>
			listMyVocabulary({ search: search ?? undefined, pageSize: MaxPageSize }),
		enabled: search !== null,
	});
}

/** One numbered page of my words under the given filters. */
export function libraryPageQueryOptions(filters: VocabFilters, page: number) {
	return {
		queryKey: vocabularyKeys.libraryPage(filters, page),
		queryFn: () =>
			listMyVocabulary({
				page,
				pageSize: WORDS_PAGE_SIZE,
				search: filters.search || undefined,
				sort: filters.sort,
				cefr: filters.cefr === "all" ? undefined : filters.cefr,
				source: filters.source === "all" ? undefined : filters.source,
			}),
	};
}

export function useLibraryPage(filters: VocabFilters, page: number) {
	return useQuery(libraryPageQueryOptions(filters, page));
}

export function useVocabularyStats(dayStart?: string) {
	return useQuery(vocabularyStatsQueryOptions(dayStart));
}

export function useVocabularyItem(id: string) {
	return useQuery(vocabularyItemQueryOptions(id));
}

/** Mastery and note writes share one endpoint; both refresh the hub + detail. */
export function usePatchVocabulary() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: ({
			itemId,
			patch,
		}: {
			itemId: string;
			patch: { mastered?: boolean; note?: string };
		}) => patchUserVocabulary(itemId, patch),
		onSuccess: (_data, { itemId }) => {
			queryClient.invalidateQueries({ queryKey: vocabularyKeys.library });
			queryClient.invalidateQueries({ queryKey: statsPrefix });
			queryClient.invalidateQueries({ queryKey: vocabularyKeys.item(itemId) });
			// Deck pages render these same items, so their reads must refresh too
			// — otherwise the card's mastered state never flips there.
			queryClient.invalidateQueries({ queryKey: deckPrefix });
			queryClient.invalidateQueries({ queryKey: decksPrefix });
		},
	});
}

export function useAddWord() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (input: AddWordInput & { deckId?: string }) =>
			input.deckId
				? addWordToDeck(input.deckId, input)
				: addVocabularyWord(input),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: vocabularyKeys.library });
			queryClient.invalidateQueries({ queryKey: statsPrefix });
			queryClient.invalidateQueries({ queryKey: deckPrefix });
			queryClient.invalidateQueries({ queryKey: decksPrefix });
		},
	});
}

/** Generation writes nothing, so there is nothing to invalidate. */
export function useGenerateWordDraft() {
	return useMutation({ mutationFn: generateWordDraft });
}

export function useAddExistingWordToDeck(deckId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (itemId: string) => addExistingWordToDeck(deckId, itemId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: deckPrefix });
			queryClient.invalidateQueries({ queryKey: decksPrefix });
		},
	});
}

export function useRemoveDeckItem() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: removeDeckItem,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: deckPrefix });
			queryClient.invalidateQueries({ queryKey: decksPrefix });
		},
	});
}
