import type {
	ExamType,
	PatchExamAttempt,
	StartExamAttempt,
} from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	cancelExamAttempt,
	getExamAttempt,
	getExamStatistics,
	getExamTest,
	listExamAttempts,
	listExamCategories,
	listExamTests,
	patchExamAttempt,
	startExamAttempt,
} from "#/features/exams/service";

export const examKeys = {
	all: ["exams"] as const,
	statistics: () => [...examKeys.all, "statistics"] as const,
	categories: (params?: { type?: ExamType }) =>
		[...examKeys.all, "categories", params?.type ?? "all"] as const,
	tests: (categoryId: string, params?: { page?: number; pageSize?: number }) =>
		[...examKeys.all, "categories", categoryId, "tests", params] as const,
	attempts: (params?: { page?: number; pageSize?: number; type?: ExamType }) =>
		[...examKeys.all, "attempts", params] as const,
	attempt: (id: string) => [...examKeys.all, "attempt", id] as const,
	test: (id: string) => [...examKeys.all, "test", id] as const,
};

export function useCancelExamAttempt() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: cancelExamAttempt,
		retry: false,
		onSuccess: () => {
			void queryClient.invalidateQueries({ queryKey: examKeys.statistics() });
		},
	});
}

export function useContinueExamAttempt() {
	const queryClient = useQueryClient();
	return useMutation({
		mutationFn: getExamAttempt,
		retry: false,
		onSuccess: (attempt) => {
			if (attempt.status === "completed")
				void queryClient.invalidateQueries({ queryKey: examKeys.statistics() });
		},
	});
}

export function examStatisticsQueryOptions() {
	return { queryKey: examKeys.statistics(), queryFn: getExamStatistics };
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
	type?: ExamType;
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

export function useExamAttempts(params?: {
	page?: number;
	pageSize?: number;
	type?: ExamType;
}) {
	return useQuery(examAttemptsQueryOptions(params));
}

export function useExamAttempt(id: string | undefined) {
	return useQuery({
		...examAttemptQueryOptions(id ?? "none"),
		enabled: !!id,
	});
}

export function useStartExamAttempt() {
	const queryClient = useQueryClient();
	return useMutation({
		retry: false,
		mutationFn: (input: { testId: string; req: StartExamAttempt }) =>
			startExamAttempt(input.testId, input.req),
		onSuccess: (_, input) => {
			void queryClient.invalidateQueries({
				queryKey: [...examKeys.all, "categories"],
			});
			void queryClient.invalidateQueries({
				queryKey: examKeys.test(input.testId),
			});
		},
	});
}

export function usePatchExamAttempt() {
	const queryClient = useQueryClient();
	return useMutation({
		retry: false,
		mutationFn: (input: { attemptId: string; req: PatchExamAttempt }) =>
			patchExamAttempt(input.attemptId, input.req),
		onSuccess: (attempt) => {
			if (attempt.status === "completed")
				void queryClient.invalidateQueries({ queryKey: examKeys.statistics() });
		},
	});
}
