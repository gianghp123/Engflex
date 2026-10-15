import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	addExistingWordToDeck,
	addVocabularyWord,
	addWordToDeck,
	createDeck,
	deleteDeck,
	getDeck,
	getVocabularyItem,
	getVocabularyStats,
	listDecks,
	listMyVocabulary,
	patchUserVocabulary,
	removeDeckItem,
	updateDeck,
} from "./service";
import type { AddWordInput } from "./types";

export const vocabularyKeys = {
	library: ["vocabulary", "library"] as const,
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

export function libraryQueryOptions() {
	return {
		queryKey: vocabularyKeys.library,
		queryFn: () => listMyVocabulary(),
	};
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

export function useMyVocabulary() {
	return useQuery(libraryQueryOptions());
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
