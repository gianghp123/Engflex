import { useAuth } from "@clerk/tanstack-react-start";
import type { ContentType } from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { vocabularyKeys } from "#/features/vocabulary/queries";
import { unsaveUserVocabulary } from "#/features/vocabulary/service";
import { ApiError } from "#/lib/api";
import {
	listBookmarks,
	listSavedLibrary,
	listStoryBookmarks,
	saveBookmark,
	savedLibraryKey,
	unsaveBookmark,
	unsaveStoryBookmark,
	updateSavedNote,
} from "./service";

export const bookmarkKeys = {
	stories: (userId: string | null | undefined) =>
		["bookmarks", userId, "story"] as const,
};
export function useStoryBookmarks() {
	const { isLoaded, userId } = useAuth();
	return useQuery({
		queryKey: bookmarkKeys.stories(userId),
		queryFn: listStoryBookmarks,
		enabled: isLoaded && !!userId,
		retry: (count, error) =>
			!(error instanceof ApiError && [401, 403, 404].includes(error.status)) &&
			count < 3,
	});
}
export function useToggleStoryBookmark() {
	const { userId } = useAuth();
	const client = useQueryClient();
	return useMutation({
		mutationFn: async ({ id, saved }: { id: string; saved: boolean }) => {
			if (saved) await unsaveStoryBookmark(id);
			else return await saveBookmark({ contentType: "story", contentId: id });
		},
		onSuccess: (bookmark, { id, saved }) => {
			void client.invalidateQueries({ queryKey: savedLibraryKey });
			client.setQueryData<Awaited<ReturnType<typeof listStoryBookmarks>>>(
				bookmarkKeys.stories(userId),
				(cached) => {
					if (!cached) return cached;
					const remaining = cached.filter((item) => item.contentId !== id);
					return saved || !bookmark ? remaining : [bookmark, ...remaining];
				},
			);
			return client.invalidateQueries({
				queryKey: bookmarkKeys.stories(userId),
			});
		},
	});
}

export function useContentBookmarks(type: ContentType) {
	const { isLoaded, userId } = useAuth();
	return useQuery({
		queryKey: ["bookmarks", userId, type],
		queryFn: () => listBookmarks(type),
		enabled: isLoaded && !!userId,
		retry: (count, error) =>
			!(error instanceof ApiError && [401, 403, 404].includes(error.status)) &&
			count < 3,
	});
}

export function useSavedLibrary(type: string, q: string, page: number) {
	const { isLoaded, userId } = useAuth();
	return useQuery({
		queryKey: [...savedLibraryKey, userId, type, q, page],
		queryFn: () => listSavedLibrary({ type, q, page }),
		enabled: isLoaded && !!userId,
		retry: (count, error) =>
			!(error instanceof ApiError && [401, 403, 404].includes(error.status)) &&
			count < 3,
	});
}

export function useSaveContentBookmark() {
	const client = useQueryClient();
	return useMutation({
		mutationFn: saveBookmark,
		onSuccess: async () => {
			await Promise.all([
				client.invalidateQueries({ queryKey: ["bookmarks"] }),
				client.invalidateQueries({ queryKey: savedLibraryKey }),
			]);
		},
	});
}

export function useRemoveSavedItem() {
	const client = useQueryClient();
	return useMutation({
		mutationFn: ({ type, id }: { type: ContentType | "vocab"; id: string }) =>
			type === "vocab" ? unsaveUserVocabulary(id) : unsaveBookmark(type, id),
		onSuccess: async (_data, { type }) => {
			await Promise.all([
				client.invalidateQueries({
					queryKey:
						type === "vocab" ? vocabularyKeys.userVocabulary : ["bookmarks"],
				}),
				client.invalidateQueries({ queryKey: savedLibraryKey }),
			]);
		},
	});
}

export function useUpdateSavedNote() {
	const client = useQueryClient();
	return useMutation({
		mutationFn: ({
			type,
			id,
			note,
		}: {
			type: string;
			id: string;
			note: string;
		}) => updateSavedNote(type, id, { note }),
		onSuccess: async (_data, { type }) => {
			await Promise.all([
				client.invalidateQueries({ queryKey: savedLibraryKey }),
				type === "vocab"
					? client.invalidateQueries({
							queryKey: vocabularyKeys.userVocabulary,
						})
					: undefined,
			]);
		},
	});
}
