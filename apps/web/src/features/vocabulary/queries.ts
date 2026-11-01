import type {
	CreateVocabularyItem,
	SaveUserVocabulary,
	UserVocabularyState,
	VocabularyItem,
} from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	createVocabularyItem,
	getVocabularyItem,
	listUserVocabulary,
	listVocabularyItems,
	saveUserVocabulary,
	unsaveUserVocabulary,
} from "#/features/vocabulary/service";

// Large, single-page fetches: the catalog and a learner's saved set are
// both admin/personal-curated and small (tens to low hundreds of rows), so
// one page covers "my vocabulary" without its own pagination UI.
const CATALOG_PAGE_SIZE = 200;

export const vocabularyKeys = {
	catalog: ["vocabulary-items"] as const,
	item: (id: string) => ["vocabulary-item", id] as const,
	userVocabulary: ["user-vocabulary"] as const,
};

export function vocabularyItemQueryOptions(id: string) {
	return {
		queryKey: vocabularyKeys.item(id),
		queryFn: () => getVocabularyItem(id),
	};
}

export function useVocabularyItem(id: string | undefined) {
	return useQuery({
		...vocabularyItemQueryOptions(id ?? "none"),
		enabled: !!id,
	});
}

function userVocabularyListQueryOptions() {
	return {
		queryKey: vocabularyKeys.userVocabulary,
		queryFn: () => listUserVocabulary(1, CATALOG_PAGE_SIZE),
	};
}

function catalogQueryOptions() {
	return {
		queryKey: vocabularyKeys.catalog,
		queryFn: () => listVocabularyItems(1, CATALOG_PAGE_SIZE),
	};
}

function attachUserState(
	item: VocabularyItem,
	state: UserVocabularyState | undefined,
): VocabularyItem {
	return state ? { ...item, userState: state } : item;
}

/** "My vocabulary": every word the caller saved, joined with its dictionary
 *  entry. A saved row whose item lookup is missing (deleted entry) is
 *  dropped rather than rendered half-empty. */
export function useMyVocabularyItems() {
	const savedQuery = useQuery(userVocabularyListQueryOptions());
	const catalogQuery = useQuery(catalogQueryOptions());

	const saved = savedQuery.data?.items ?? [];
	const catalog = catalogQuery.data?.items ?? [];
	const catalogById = new Map(catalog.map((item) => [item.id, item]));

	const items: VocabularyItem[] = saved.flatMap((state) => {
		const item = catalogById.get(state.itemId);
		return item ? [attachUserState(item, state)] : [];
	});

	return {
		items,
		isPending: savedQuery.isPending || catalogQuery.isPending,
		isError: savedQuery.isError || catalogQuery.isError,
	};
}

/** One word's detail, with the caller's saved state attached if present
 *  (shares the cache key with useMyVocabularyItems' saved-list fetch). */
export function useVocabularyItemDetail(id: string | undefined) {
	const itemQuery = useVocabularyItem(id);
	const savedQuery = useQuery(userVocabularyListQueryOptions());
	const state = savedQuery.data?.items.find((row) => row.itemId === id);
	return {
		...itemQuery,
		data: itemQuery.data ? attachUserState(itemQuery.data, state) : undefined,
	};
}

function useInvalidateVocabulary() {
	const queryClient = useQueryClient();
	return () => {
		void queryClient.invalidateQueries({ queryKey: vocabularyKeys.catalog });
		void queryClient.invalidateQueries({
			queryKey: vocabularyKeys.userVocabulary,
		});
	};
}

export function useSaveUserVocabulary() {
	const invalidate = useInvalidateVocabulary();
	return useMutation({
		mutationFn: (payload: SaveUserVocabulary) => saveUserVocabulary(payload),
		onSuccess: invalidate,
	});
}

export function useUnsaveUserVocabulary() {
	const invalidate = useInvalidateVocabulary();
	return useMutation({
		mutationFn: (itemId: string) => unsaveUserVocabulary(itemId),
		onSuccess: invalidate,
	});
}

/** Create a personal word, then save it so it shows up in "my vocabulary"
 *  immediately — the backend has no implicit self-save on create. */
export function useCreateCustomWord() {
	const invalidate = useInvalidateVocabulary();
	return useMutation({
		mutationFn: async (payload: Omit<CreateVocabularyItem, "isPremium">) => {
			const item = await createVocabularyItem(payload);
			await saveUserVocabulary({ itemId: item.id, sourceType: "manual" });
			return item;
		},
		onSuccess: invalidate,
	});
}
