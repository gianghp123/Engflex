import type {
	Conversation,
	ListScenarios,
	StartConversation,
} from "@engflex/contracts";
import {
	useInfiniteQuery,
	useMutation,
	useQuery,
	useQueryClient,
} from "@tanstack/react-query";
import {
	analyzeTurn,
	createConversation,
	endConversation,
	getConversation,
	getScenarioById,
	listConversations,
	listScenarios,
	listTopicsWithPreview,
} from "#/features/voice/service";

export const conversationKeys = {
	byId: (id: string) => ["conversation", id] as const,
};

export const conversationListKeys = {
	infinite: (scenarioId?: string) =>
		["conversations", "list", scenarioId ?? "all"] as const,
	recent: (limit: number) => ["conversations", "recent", limit] as const,
	recentForScenario: (scenarioId: string, limit: number) =>
		["conversations", "recent", scenarioId, limit] as const,
};

const HISTORY_PAGE_SIZE = 20;

/** Shared options so the detail route loader and the component read one cache entry. */
export function conversationQueryOptions(id: string) {
	return {
		queryKey: conversationKeys.byId(id),
		queryFn: () => getConversation(id),
	};
}

export function useConversation(
	id: string,
	opts?: { refetchInterval?: number; enabled?: boolean },
) {
	return useQuery({
		...conversationQueryOptions(id),
		enabled: id.length > 0 && (opts?.enabled ?? true),
		refetchInterval: opts?.refetchInterval,
	});
}

export function useCreateConversation() {
	return useMutation({
		mutationFn: (input: StartConversation) => createConversation(input),
	});
}

export function useEndConversation(id: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: () => endConversation(id),
		onSuccess: () =>
			queryClient.invalidateQueries({ queryKey: conversationKeys.byId(id) }),
	});
}

export function useAnalyzeTurn(conversationId: string) {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: (position: number) => analyzeTurn(conversationId, position),
		onSuccess: (feedback, position) =>
			queryClient.setQueryData<Conversation>(
				conversationKeys.byId(conversationId),
				(prev) =>
					prev
						? {
								...prev,
								turns: prev.turns.map((turn) =>
									turn.position === position ? { ...turn, feedback } : turn,
								),
							}
						: prev,
			),
	});
}

export const scenarioKeys = {
	all: ["scenarios"] as const,
	topic: (topicId: string) => ["scenarios", topicId] as const,
};

export function useScenarios(params?: Partial<ListScenarios>) {
	return useQuery({
		queryKey: [
			...scenarioKeys.all,
			params?.topicId ?? "all",
			params?.random ? "random" : "recent",
			params?.pageSize ?? 0,
		],
		queryFn: () => listScenarios(params),
	});
}

export function topicsWithPreviewQueryOptions(params?: {
	difficulty?: string;
	search?: string;
}) {
	const difficulty = params?.difficulty;
	const search = params?.search;
	return {
		queryKey: ["scenario-topics", difficulty ?? "all", search ?? ""] as const,
		queryFn: () => listTopicsWithPreview({ previewK: 3, difficulty, search }),
	};
}

export function useTopicsWithPreview(params?: {
	difficulty?: string;
	search?: string;
}) {
	return useQuery(topicsWithPreviewQueryOptions(params));
}

export function scenarioQueryOptions(id: string) {
	return {
		queryKey: ["scenario", id] as const,
		queryFn: () => getScenarioById(id),
	};
}

export function useScenario(id: string | undefined) {
	return useQuery({
		...scenarioQueryOptions(id ?? "none"),
		enabled: !!id,
	});
}

/** Pages the caller's finished conversations, newest first. */
export function useConversationHistory(scenarioId?: string) {
	return useInfiniteQuery({
		queryKey: conversationListKeys.infinite(scenarioId),
		queryFn: ({ pageParam }) =>
			listConversations({
				scenarioId,
				page: pageParam,
				pageSize: HISTORY_PAGE_SIZE,
			}),
		initialPageParam: 1,
		getNextPageParam: (last) =>
			last.pagination.page < last.pagination.totalPages
				? last.pagination.page + 1
				: undefined,
	});
}

/** The newest sessions for the voice hub preview. */
export function useRecentConversations(limit: number) {
	return useQuery({
		queryKey: conversationListKeys.recent(limit),
		queryFn: () => listConversations({ pageSize: limit }),
	});
}

/** The newest finished sessions on one scenario (the detail page's runs). */
export function useScenarioRuns(scenarioId: string, limit: number) {
	return useQuery({
		queryKey: conversationListKeys.recentForScenario(scenarioId, limit),
		queryFn: () => listConversations({ scenarioId, pageSize: limit }),
	});
}
