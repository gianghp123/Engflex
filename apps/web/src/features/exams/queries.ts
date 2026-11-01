import { useAuth } from "@clerk/tanstack-react-start";
import type {
	ExamType,
	PatchExamAttempt,
	StartExamAttempt,
} from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { clearAudioPosition } from "#/features/exams/audio-position";
import {
	cancelExamAttempt,
	getExamAttempt,
	getExamStats,
	getExamTest,
	listExamAttempts,
	listExamCategories,
	listExamTests,
	patchExamAttempt,
	startExamAttempt,
} from "#/features/exams/service";

export const examKeys = {
	all: ["exams"] as const,
	stats: ["exams", "stats"] as const,
	categories: (params?: { type?: ExamType }) =>
		[...examKeys.all, "categories", params?.type ?? "all"] as const,
	tests: (categoryId: string, params?: { page?: number; pageSize?: number }) =>
		[...examKeys.all, "categories", categoryId, "tests", params] as const,
	attempts: (params?: { page?: number; pageSize?: number }) =>
		[...examKeys.all, "attempts", params] as const,
	attempt: (id: string) => [...examKeys.all, "attempt", id] as const,
	test: (id: string) => [...examKeys.all, "test", id] as const,
};

export function examStatsQueryOptions() {
	return {
		queryKey: examKeys.stats,
		queryFn: getExamStats,
		staleTime: 0,
		refetchOnMount: "always" as const,
	};
}

export function useCancelExamAttempt() {
	const queryClient = useQueryClient();
	const { userId } = useAuth();
	return useMutation({
		mutationFn: cancelExamAttempt,
		retry: false,
		onSuccess: (attempt) => {
			if (
				userId &&
				(attempt.status === "completed" || attempt.status === "abandoned")
			)
				clearAudioPosition(userId, attempt.id);
			void queryClient.invalidateQueries({ queryKey: examKeys.stats });
		},
	});
}

export function useContinueExamAttempt() {
	const queryClient = useQueryClient();
	const { userId } = useAuth();
	return useMutation({
		mutationFn: getExamAttempt,
		retry: false,
		onSuccess: (attempt) => {
			if (
				userId &&
				(attempt.status === "completed" || attempt.status === "abandoned")
			)
				clearAudioPosition(userId, attempt.id);
			if (attempt.status === "completed") {
				void queryClient.invalidateQueries({ queryKey: examKeys.stats });
			}
		},
	});
}

export function examTestQueryOptions(id: string) {
	return {
		queryKey: examKeys.test(id),
		queryFn: () => getExamTest(id),
	};
}

export function examCategoriesQueryOptions(params?: { type?: ExamType }) {
	return {
		queryKey: examKeys.categories(params),
		queryFn: () => listExamCategories(params),
	};
}

export function examTestsQueryOptions(
	categoryId: string,
	params?: { page?: number; pageSize?: number },
) {
	return {
		queryKey: examKeys.tests(categoryId, params),
		queryFn: () => listExamTests(categoryId, params),
	};
}

export function examAttemptsQueryOptions(params?: {
	page?: number;
	pageSize?: number;
}) {
	return {
		queryKey: examKeys.attempts(params),
		queryFn: () => listExamAttempts(params),
	};
}

export function examAttemptQueryOptions(id: string) {
	return {
		queryKey: examKeys.attempt(id),
		queryFn: () => getExamAttempt(id),
	};
}

export function useExamCategories(params?: { type?: ExamType }) {
	return useQuery(examCategoriesQueryOptions(params));
}

export function useExamTests(
	categoryId: string | undefined,
	params?: { page?: number; pageSize?: number },
) {
	return useQuery({
		...examTestsQueryOptions(categoryId ?? "none", params),
		enabled: !!categoryId,
	});
}

export function useExamAttempts(params?: { page?: number; pageSize?: number }) {
	return useQuery(examAttemptsQueryOptions(params));
}

export function useExamAttempt(id: string | undefined) {
	return useQuery({
		...examAttemptQueryOptions(id ?? "none"),
		enabled: !!id,
	});
}

export function useStartExamAttempt() {
	return useMutation({
		retry: false,
		mutationFn: (input: { testId: string; req: StartExamAttempt }) =>
			startExamAttempt(input.testId, input.req),
	});
}

export function usePatchExamAttempt() {
	return useMutation({
		retry: false,
		mutationFn: (input: { attemptId: string; req: PatchExamAttempt }) =>
			patchExamAttempt(input.attemptId, input.req),
	});
}
