import { useAuth } from "@clerk/tanstack-react-start";
import type { ListStories } from "@engflex/contracts";
import { useQuery } from "@tanstack/react-query";
import { ApiError } from "#/lib/api";
import {
	getStoryById,
	getStoryBySlug,
	listStories,
	listStoryTopics,
} from "./service";

export const storyKeys = {
	all: ["stories"] as const,
	lists: () => [...storyKeys.all, "list"] as const,
	list: (params?: Partial<ListStories>) =>
		[...storyKeys.lists(), params] as const,
	details: () => [...storyKeys.all, "detail"] as const,
	detailBySlug: (slug: string) =>
		[...storyKeys.details(), "slug", slug] as const,
	detailById: (id: string) => [...storyKeys.details(), "id", id] as const,
	topics: () => [...storyKeys.all, "topics"] as const,
};

export function storiesListQueryOptions(params?: Partial<ListStories>) {
	return {
		queryKey: storyKeys.list(params),
		queryFn: () => listStories(params),
	};
}

export function storyDetailBySlugQueryOptions(slug: string) {
	return {
		queryKey: storyKeys.detailBySlug(slug),
		queryFn: () => getStoryBySlug(slug),
	};
}

export function storyDetailByIdQueryOptions(id: string) {
	return {
		queryKey: storyKeys.detailById(id),
		queryFn: () => getStoryById(id),
	};
}

export function storyTopicsQueryOptions() {
	return {
		queryKey: storyKeys.topics(),
		queryFn: () => listStoryTopics(),
	};
}

function retryStoryRead(count: number, error: unknown) {
	return (
		!(error instanceof ApiError && [401, 403, 404].includes(error.status)) &&
		count < 3
	);
}
export function useStories(params?: Partial<ListStories>, enabled = true) {
	const { isLoaded } = useAuth();
	return useQuery({
		...storiesListQueryOptions(params),
		enabled: isLoaded && enabled,
		retry: retryStoryRead,
	});
}

export function useStoryDetail(slug: string | undefined) {
	const { isLoaded } = useAuth();
	return useQuery({
		...storyDetailBySlugQueryOptions(slug ?? "none"),
		enabled: isLoaded && !!slug,
		retry: retryStoryRead,
	});
}

export function useStoryTopics() {
	const { isLoaded } = useAuth();
	return useQuery({
		...storyTopicsQueryOptions(),
		enabled: isLoaded,
		retry: retryStoryRead,
	});
}
