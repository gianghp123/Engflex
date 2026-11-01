import type {
	Activity,
	AdminActivity,
	CreateLesson,
	CreateLessonActivity,
	CreateLessonSection,
	Lesson,
	Section,
	UnitSection,
	UpdateLesson,
	UpdateLessonActivity,
	UpdateLessonSection,
} from "@engflex/contracts";
import { API_ROUTES } from "#/app/api-routes";
import { api } from "#/lib/api";

/** The hub endpoint always returns sections grouped with their units — the
 *  admin table flattens this itself rather than the server exposing a
 *  second, parallel listing shape. Unpaginated, like the learner hub. */
export function listLessonHub(): Promise<UnitSection[]> {
	return api<UnitSection[]>(API_ROUTES.LESSONS.LIST);
}

export function createLesson(payload: CreateLesson): Promise<Lesson> {
	return api<Lesson>(API_ROUTES.LESSONS.LIST, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updateLesson(
	id: string,
	payload: UpdateLesson,
): Promise<Lesson> {
	return api<Lesson>(API_ROUTES.LESSONS.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deleteLesson(id: string): Promise<void> {
	return api<void>(API_ROUTES.LESSONS.BY_ID(id), { method: "DELETE" });
}

/** Every section, including ones with no lessons yet — admin-only
 *  unfiltered list, unlike the hub's grouped-with-units shape. */
export function listLessonSections(): Promise<Section[]> {
	return api<Section[]>(API_ROUTES.LESSON_SECTIONS.LIST);
}

export function createLessonSection(
	payload: CreateLessonSection,
): Promise<Section> {
	return api<Section>(API_ROUTES.LESSON_SECTIONS.LIST, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updateLessonSection(
	id: string,
	payload: UpdateLessonSection,
): Promise<Section> {
	return api<Section>(API_ROUTES.LESSON_SECTIONS.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deleteLessonSection(id: string): Promise<void> {
	return api<void>(API_ROUTES.LESSON_SECTIONS.BY_ID(id), { method: "DELETE" });
}

/** Learner-shaped list (title/type/part/duration) — enough for the admin
 *  table; opening one for edit fetches the full config separately. */
export function listLessonActivities(lessonId: string): Promise<Activity[]> {
	return api<Activity[]>(API_ROUTES.LESSON_ACTIVITIES.LIST, undefined, {
		query: { lessonId },
	});
}

export function getAdminLessonActivity(id: string): Promise<AdminActivity> {
	return api<AdminActivity>(API_ROUTES.ADMIN.LESSON_ACTIVITY(id));
}

export function createLessonActivity(
	payload: CreateLessonActivity,
): Promise<AdminActivity> {
	return api<AdminActivity>(API_ROUTES.LESSON_ACTIVITIES.LIST, {
		method: "POST",
		body: JSON.stringify(payload),
	});
}

export function updateLessonActivity(
	id: string,
	payload: UpdateLessonActivity,
): Promise<AdminActivity> {
	return api<AdminActivity>(API_ROUTES.LESSON_ACTIVITIES.BY_ID(id), {
		method: "PUT",
		body: JSON.stringify(payload),
	});
}

export function deleteLessonActivity(id: string): Promise<void> {
	return api<void>(API_ROUTES.LESSON_ACTIVITIES.BY_ID(id), {
		method: "DELETE",
	});
}
