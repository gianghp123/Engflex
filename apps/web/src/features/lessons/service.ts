import type {
	Attempt,
	AttemptProgress,
	CheckResult,
	LessonDetail,
	LessonProgressSummary,
	PronunciationResult,
	UnitSection,
	WritingScoreResponse,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api } from "#/lib/api";

/** Whole path, unpaginated: sections arrive with units attached. */
export function listSections(params?: {
	level?: string;
}): Promise<UnitSection[]> {
	return api<UnitSection[]>(API_ROUTES.LESSONS.LIST, undefined, {
		query: params?.level ? { level: params.level } : undefined,
	});
}

export function getLessonDetail(id: string): Promise<LessonDetail> {
	return api<LessonDetail>(API_ROUTES.LESSONS.BY_ID(id));
}

export function getLessonDetailBySlug(slug: string): Promise<LessonDetail> {
	return api<LessonDetail>(API_ROUTES.LESSONS.BY_SLUG(slug));
}

export function listLessonProgress(): Promise<LessonProgressSummary[]> {
	return api<LessonProgressSummary[]>(API_ROUTES.LESSONS.PROGRESS, undefined, {
		withCredentials: true,
	});
}

export function checkAnswer(
	activityId: string,
	attemptId: string,
	questionIndex: number,
	key: string,
): Promise<CheckResult> {
	return api<CheckResult>(API_ROUTES.LESSONS.CHECK(activityId), {
		method: "POST",
		body: JSON.stringify({ attemptId, questionIndex, key }),
	});
}

export function pronounceAttempt(
	activityId: string,
	attemptId: string,
	itemIndex: number,
	audio: Blob,
	mime: string,
): Promise<PronunciationResult> {
	const form = new FormData();
	form.append("attemptId", attemptId);
	form.append("itemIndex", String(itemIndex));
	form.append(
		"audio",
		audio,
		mime.includes("wav") ? "attempt.wav" : "attempt.webm",
	);
	return api<PronunciationResult>(API_ROUTES.LESSONS.PRONOUNCE(activityId), {
		method: "POST",
		body: form,
	});
}

export function startLessonAttempt(lessonId: string): Promise<Attempt> {
	return api<Attempt>(API_ROUTES.LESSONS.START_ATTEMPT(lessonId), {
		method: "POST",
	});
}

/** Open run with graded questions for resume; 404 when no run is open. */
export function getOpenAttempt(lessonId: string): Promise<AttemptProgress> {
	return api<AttemptProgress>(API_ROUTES.LESSONS.OPEN_ATTEMPT(lessonId));
}

export function scoreWriting(
	activityId: string,
	attemptId: string,
	text: string,
): Promise<WritingScoreResponse> {
	return api<WritingScoreResponse>(API_ROUTES.LESSONS.SCORE(activityId), {
		method: "POST",
		body: JSON.stringify({ attemptId, text }),
	});
}

export type {
	Attempt,
	AttemptProgress,
	CheckResult,
	LessonDetail,
	LessonProgressSummary,
	PronunciationResult,
	UnitSection,
	WritingScoreResponse,
};
