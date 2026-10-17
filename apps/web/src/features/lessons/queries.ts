import { useMutation, useQuery } from "@tanstack/react-query";
import {
	checkAnswer,
	getLessonDetail,
	getLessonDetailBySlug,
	getOpenAttempt,
	listLessonProgress,
	listSections,
	pronounceAttempt,
	scoreWriting,
	startLessonAttempt,
} from "#/features/lessons/service";
import { ApiError } from "#/lib/api/api";

export const lessonKeys = {
	sections: ["lesson-sections"] as const,
	progress: ["lessons", "progress"] as const,
	detail: (id: string) => ["lesson", id] as const,
	detailBySlug: (slug: string) => ["lesson", "slug", slug] as const,
	openAttempt: (id: string) => ["lesson", id, "open-attempt"] as const,
};

export function useLessonSections() {
	return useQuery(lessonSectionsQueryOptions());
}

/** Shared config so hub-level aggregates can `useQueries` the same objects
 *  the cards fetch (React Query dedupes identical keys). */
export function lessonDetailQueryOptions(id: string) {
	return {
		queryKey: lessonKeys.detail(id),
		queryFn: () => getLessonDetail(id),
	};
}

export function lessonDetailBySlugQueryOptions(slug: string) {
	return {
		queryKey: lessonKeys.detailBySlug(slug),
		queryFn: () => getLessonDetailBySlug(slug),
	};
}

export function lessonSectionsQueryOptions() {
	return {
		queryKey: lessonKeys.sections,
		queryFn: () => listSections(),
	};
}

export function lessonProgressQueryOptions() {
	return {
		queryKey: lessonKeys.progress,
		queryFn: () => listLessonProgress(),
	};
}

export function useLessonProgress() {
	return useQuery(lessonProgressQueryOptions());
}

export function useLessonDetail(id: string | undefined) {
	return useQuery({
		...lessonDetailQueryOptions(id ?? "none"),
		enabled: !!id,
	});
}

export function useLessonDetailBySlug(slug: string | undefined) {
	return useQuery({
		...lessonDetailBySlugQueryOptions(slug ?? "none"),
		enabled: !!slug,
	});
}

export function useCheckAnswer() {
	return useMutation({
		mutationFn: (input: {
			activityId: string;
			attemptId: string;
			questionIndex: number;
			key: string;
		}) =>
			checkAnswer(
				input.activityId,
				input.attemptId,
				input.questionIndex,
				input.key,
			),
	});
}

export function usePronounceAttempt() {
	return useMutation({
		mutationFn: (input: {
			activityId: string;
			attemptId: string;
			itemIndex: number;
			audio: Blob;
			mime: string;
		}) =>
			pronounceAttempt(
				input.activityId,
				input.attemptId,
				input.itemIndex,
				input.audio,
				input.mime,
			),
	});
}

export function useStartLessonAttempt() {
	return useMutation({
		mutationFn: (lessonId: string) => startLessonAttempt(lessonId),
	});
}

/** Open-run resume: 404 (no open run) is control flow to the fallback card,
 *  never retried; everything else retries with the default backoff instead
 *  of crashing fail-fast (doc Pattern 7). */
export function openAttemptQueryOptions(id: string) {
	return {
		queryKey: lessonKeys.openAttempt(id),
		queryFn: () => getOpenAttempt(id),
		retry: (failureCount: number, error: unknown) =>
			!(error instanceof ApiError && error.status === 404) && failureCount < 3,
	};
}

export function useScoreWriting() {
	return useMutation({
		mutationFn: (input: {
			activityId: string;
			attemptId: string;
			text: string;
		}) => scoreWriting(input.activityId, input.attemptId, input.text),
	});
}
