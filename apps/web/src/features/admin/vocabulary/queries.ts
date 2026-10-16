import type {
	CreateVocabularyCategory,
	CreateVocabularyDeck,
	CreateVocabularyDeckItem,
	CreateVocabularyItem,
	UpdateVocabularyCategory,
	UpdateVocabularyDeck,
	UpdateVocabularyDeckItem,
	UpdateVocabularyItem,
} from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	createVocabularyCategory,
	createVocabularyDeck,
	createVocabularyDeckItem,
	createVocabularyItem,
	deleteVocabularyCategory,
	deleteVocabularyDeck,
	deleteVocabularyDeckItem,
	deleteVocabularyItem,
	listVocabularyCategories,
	listVocabularyDeckItems,
	listVocabularyDecks,
	listVocabularyItems,
	updateVocabularyCategory,
	updateVocabularyDeck,
	updateVocabularyDeckItem,
	updateVocabularyItem,
} from "#/features/admin/vocabulary/service";

const keys = {
	items: (page: number, pageSize: number) =>
		["admin", "vocabulary-items", page, pageSize] as const,
	decks: (page: number, pageSize: number) =>
		["admin", "vocabulary-decks", page, pageSize] as const,
	deckItems: (deckId: string) =>
		["admin", "vocabulary-deck-items", deckId] as const,
	categories: (page: number, pageSize: number) =>
		["admin", "vocabulary-categories", page, pageSize] as const,
};

export function useAdminVocabularyItems(page: number, pageSize: number) {
	return useQuery({
		queryKey: keys.items(page, pageSize),
		queryFn: () => listVocabularyItems(page, pageSize),
	});
}

function useInvalidate(key: readonly unknown[]) {
	const queryClient = useQueryClient();
	return () => queryClient.invalidateQueries({ queryKey: key });
}

export function useCreateVocabularyItem() {
	const invalidate = useInvalidate(["admin", "vocabulary-items"]);
	return useMutation({
		mutationFn: (payload: CreateVocabularyItem) =>
			createVocabularyItem(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateVocabularyItem() {
	const invalidate = useInvalidate(["admin", "vocabulary-items"]);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateVocabularyItem;
		}) => updateVocabularyItem(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteVocabularyItem() {
	const invalidate = useInvalidate(["admin", "vocabulary-items"]);
	return useMutation({
		mutationFn: (id: string) => deleteVocabularyItem(id),
		onSuccess: invalidate,
	});
}

export function useAdminVocabularyDecks(page: number, pageSize: number) {
	return useQuery({
		queryKey: keys.decks(page, pageSize),
		queryFn: () => listVocabularyDecks(page, pageSize),
	});
}

export function useCreateVocabularyDeck() {
	const invalidate = useInvalidate(["admin", "vocabulary-decks"]);
	return useMutation({
		mutationFn: (payload: CreateVocabularyDeck) =>
			createVocabularyDeck(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateVocabularyDeck() {
	const invalidate = useInvalidate(["admin", "vocabulary-decks"]);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateVocabularyDeck;
		}) => updateVocabularyDeck(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteVocabularyDeck() {
	const invalidate = useInvalidate(["admin", "vocabulary-decks"]);
	return useMutation({
		mutationFn: (id: string) => deleteVocabularyDeck(id),
		onSuccess: invalidate,
	});
}

export function useAdminVocabularyDeckItems(deckId: string) {
	return useQuery({
		queryKey: keys.deckItems(deckId),
		queryFn: () => listVocabularyDeckItems(deckId),
		enabled: deckId.length > 0,
	});
}

export function useCreateVocabularyDeckItem(deckId: string) {
	const invalidate = useInvalidate(keys.deckItems(deckId));
	return useMutation({
		mutationFn: (payload: CreateVocabularyDeckItem) =>
			createVocabularyDeckItem(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateVocabularyDeckItem(deckId: string) {
	const invalidate = useInvalidate(keys.deckItems(deckId));
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateVocabularyDeckItem;
		}) => updateVocabularyDeckItem(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteVocabularyDeckItem(deckId: string) {
	const invalidate = useInvalidate(keys.deckItems(deckId));
	return useMutation({
		mutationFn: (id: string) => deleteVocabularyDeckItem(id),
		onSuccess: invalidate,
	});
}

export function useAdminVocabularyCategories(page: number, pageSize: number) {
	return useQuery({
		queryKey: keys.categories(page, pageSize),
		queryFn: () => listVocabularyCategories(page, pageSize),
	});
}

export function useCreateVocabularyCategory() {
	const invalidate = useInvalidate(["admin", "vocabulary-categories"]);
	return useMutation({
		mutationFn: (payload: CreateVocabularyCategory) =>
			createVocabularyCategory(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateVocabularyCategory() {
	const invalidate = useInvalidate(["admin", "vocabulary-categories"]);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateVocabularyCategory;
		}) => updateVocabularyCategory(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteVocabularyCategory() {
	const invalidate = useInvalidate(["admin", "vocabulary-categories"]);
	return useMutation({
		mutationFn: (id: string) => deleteVocabularyCategory(id),
		onSuccess: invalidate,
	});
}
