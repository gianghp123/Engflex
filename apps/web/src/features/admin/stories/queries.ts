import type { CreateStory, UpdateStory } from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	type AdminStoryFilters,
	createStory,
	deleteStory,
	listAdminStories,
	updateStory,
	uploadStoryAsset,
} from "#/features/admin/stories/service";
import { storyKeys } from "#/features/stories/queries";

const keys = {
	all: ["admin", "stories"] as const,
	list: (page: number, pageSize: number, filters: AdminStoryFilters) =>
		["admin", "stories", page, pageSize, filters] as const,
};

function useInvalidate() {
	const queryClient = useQueryClient();
	return () =>
		Promise.all([
			queryClient.invalidateQueries({ queryKey: keys.all }),
			queryClient.invalidateQueries({ queryKey: storyKeys.all }),
		]);
}

export function useAdminStories(
	page: number,
	pageSize: number,
	filters: AdminStoryFilters = {},
) {
	return useQuery({
		queryKey: keys.list(page, pageSize, filters),
		queryFn: () => listAdminStories(page, pageSize, filters),
	});
}

export function useCreateStory() {
	const invalidate = useInvalidate();
	return useMutation({
		mutationFn: (payload: CreateStory) => createStory(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateStory() {
	const invalidate = useInvalidate();
	return useMutation({
		mutationFn: ({ id, payload }: { id: string; payload: UpdateStory }) =>
			updateStory(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteStory() {
	const invalidate = useInvalidate();
	return useMutation({
		mutationFn: (id: string) => deleteStory(id),
		onSuccess: invalidate,
	});
}

export function useUploadStoryAsset() {
	return useMutation({
		mutationFn: (file: File) => uploadStoryAsset(file),
	});
}
