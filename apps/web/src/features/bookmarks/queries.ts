import { useAuth } from "@clerk/tanstack-react-start";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ApiError } from "#/lib/api";
import {
	listStoryBookmarks,
	saveBookmark,
	unsaveStoryBookmark,
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
