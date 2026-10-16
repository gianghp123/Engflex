import type {
	CreateVideoCategory,
	CreateVideoExercise,
	UpdateVideoCategory,
	UpdateVideoExercise,
} from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	createVideoCategory,
	createVideoExercise,
	deleteVideoCategory,
	deleteVideoExercise,
	listVideoCategories,
	listVideoExercises,
	updateVideoCategory,
	updateVideoExercise,
} from "#/features/admin/videos/service";

const keys = {
	exercises: (page: number, pageSize: number) =>
		["admin", "video-exercises", page, pageSize] as const,
	categories: (page: number, pageSize: number) =>
		["admin", "video-categories", page, pageSize] as const,
};

function useInvalidate(key: readonly unknown[]) {
	const queryClient = useQueryClient();
	return () => queryClient.invalidateQueries({ queryKey: key });
}

export function useAdminVideoExercises(page: number, pageSize: number) {
	return useQuery({
		queryKey: keys.exercises(page, pageSize),
		queryFn: () => listVideoExercises(page, pageSize),
	});
}

export function useCreateVideoExercise() {
	const invalidate = useInvalidate(["admin", "video-exercises"]);
	return useMutation({
		mutationFn: (payload: CreateVideoExercise) => createVideoExercise(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateVideoExercise() {
	const invalidate = useInvalidate(["admin", "video-exercises"]);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateVideoExercise;
		}) => updateVideoExercise(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteVideoExercise() {
	const invalidate = useInvalidate(["admin", "video-exercises"]);
	return useMutation({
		mutationFn: (id: string) => deleteVideoExercise(id),
		onSuccess: invalidate,
	});
}

export function useAdminVideoCategories(page: number, pageSize: number) {
	return useQuery({
		queryKey: keys.categories(page, pageSize),
		queryFn: () => listVideoCategories(page, pageSize),
	});
}

export function useCreateVideoCategory() {
	const invalidate = useInvalidate(["admin", "video-categories"]);
	return useMutation({
		mutationFn: (payload: CreateVideoCategory) => createVideoCategory(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateVideoCategory() {
	const invalidate = useInvalidate(["admin", "video-categories"]);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateVideoCategory;
		}) => updateVideoCategory(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteVideoCategory() {
	const invalidate = useInvalidate(["admin", "video-categories"]);
	return useMutation({
		mutationFn: (id: string) => deleteVideoCategory(id),
		onSuccess: invalidate,
	});
}
