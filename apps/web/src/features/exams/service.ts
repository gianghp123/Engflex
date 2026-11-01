import type {
	ExamAnswer,
	ExamAttempt,
	ExamAttemptStats,
	ExamAttemptSummary,
	ExamCatalogCategory,
	ExamCategory,
	ExamReviewAttempt,
	ExamReviewContent,
	ExamTakingContent,
	ExamTestDetail,
	ExamTestMetadata,
	ExamType,
	Pagination,
	PatchExamAttempt,
	StartExamAttempt,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api, apiPage } from "#/lib/api/api";

export function listExamCategories(params?: {
	type?: ExamType;
}): Promise<ExamCatalogCategory[]> {
	return api<ExamCatalogCategory[]>(
		API_ROUTES.EXAMS.CATEGORIES.LIST,
		undefined,
		{
			query: params?.type ? { type: params.type } : undefined,
		},
	);
}

export function getExamStats(): Promise<ExamAttemptStats> {
	return api<ExamAttemptStats>(API_ROUTES.EXAMS.ATTEMPTS.STATS);
}

export function listExamTests(
	categoryId: string,
	params?: { page?: number; pageSize?: number },
): Promise<{ items: ExamTestMetadata[]; pagination: Pagination }> {
	const query: Record<string, string | number> = {};
	if (params?.page) query.page = params.page;
	if (params?.pageSize) query.pageSize = params.pageSize;
	return apiPage<ExamTestMetadata>(
		API_ROUTES.EXAMS.CATEGORIES.TESTS(categoryId),
		undefined,
		{ query: Object.keys(query).length > 0 ? query : undefined },
	);
}

export function getExamTest(testId: string): Promise<ExamTestDetail> {
	return api<ExamTestDetail>(API_ROUTES.EXAMS.TESTS.BY_ID(testId));
}

export function startExamAttempt(
	testId: string,
	req: StartExamAttempt,
): Promise<ExamAttempt> {
	return api<ExamAttempt>(API_ROUTES.EXAMS.TESTS.START_ATTEMPT(testId), {
		method: "POST",
		body: JSON.stringify(req),
	});
}

export function listExamAttempts(params?: {
	page?: number;
	pageSize?: number;
}): Promise<{ items: ExamAttemptSummary[]; pagination: Pagination }> {
	const query: Record<string, string | number> = {};
	if (params?.page) query.page = params.page;
	if (params?.pageSize) query.pageSize = params.pageSize;
	return apiPage<ExamAttemptSummary>(
		API_ROUTES.EXAMS.ATTEMPTS.LIST,
		undefined,
		{ query: Object.keys(query).length > 0 ? query : undefined },
	);
}

export function cancelExamAttempt(
	attemptId: string,
): Promise<ExamAttemptSummary> {
	return api<ExamAttemptSummary>(API_ROUTES.EXAMS.ATTEMPTS.CANCEL(attemptId), {
		method: "POST",
	});
}

export function getExamAttempt(
	attemptId: string,
): Promise<ExamAttempt | ExamReviewAttempt> {
	return api<ExamAttempt | ExamReviewAttempt>(
		API_ROUTES.EXAMS.ATTEMPTS.BY_ID(attemptId),
	);
}

export function patchExamAttempt(
	attemptId: string,
	req: PatchExamAttempt,
): Promise<ExamAttempt | ExamReviewAttempt> {
	return api<ExamAttempt | ExamReviewAttempt>(
		API_ROUTES.EXAMS.ATTEMPTS.BY_ID(attemptId),
		{
			method: "PATCH",
			body: JSON.stringify(req),
		},
	);
}

export function isExamReviewAttempt(
	attempt: ExamAttempt | ExamReviewAttempt,
): attempt is ExamReviewAttempt {
	return attempt.status === "completed";
}

export type {
	ExamAnswer,
	ExamAttempt,
	ExamAttemptSummary,
	ExamCategory,
	ExamReviewAttempt,
	ExamReviewContent,
	ExamTakingContent,
	ExamTestMetadata,
	ExamType,
	PatchExamAttempt,
	StartExamAttempt,
};
