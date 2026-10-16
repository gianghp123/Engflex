import type {
	CreateLesson,
	CreateLessonActivity,
	CreateLessonSection,
	UpdateLesson,
	UpdateLessonActivity,
	UpdateLessonSection,
} from "@engflex/contracts";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
	createLesson,
	createLessonActivity,
	createLessonSection,
	deleteLesson,
	deleteLessonActivity,
	deleteLessonSection,
	getAdminLessonActivity,
	listLessonActivities,
	listLessonHub,
	listLessonSections,
	updateLesson,
	updateLessonActivity,
	updateLessonSection,
} from "#/features/admin/lessons/service";

const keys = {
	hub: ["admin", "lessons"] as const,
	sections: ["admin", "lesson-sections"] as const,
	activities: (lessonId: string) =>
		["admin", "lesson-activities", lessonId] as const,
	activity: (id: string) => ["admin", "lesson-activity", id] as const,
};

function useInvalidate(key: readonly unknown[]) {
	const queryClient = useQueryClient();
	return () => queryClient.invalidateQueries({ queryKey: key });
}

export function useAdminLessonHub() {
	return useQuery({ queryKey: keys.hub, queryFn: () => listLessonHub() });
}

export function useCreateLesson() {
	const invalidate = useInvalidate(keys.hub);
	return useMutation({
		mutationFn: (payload: CreateLesson) => createLesson(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateLesson() {
	const invalidate = useInvalidate(keys.hub);
	return useMutation({
		mutationFn: ({ id, payload }: { id: string; payload: UpdateLesson }) =>
			updateLesson(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteLesson() {
	const invalidate = useInvalidate(keys.hub);
	return useMutation({
		mutationFn: (id: string) => deleteLesson(id),
		onSuccess: invalidate,
	});
}

export function useAdminLessonSections() {
	return useQuery({
		queryKey: keys.sections,
		queryFn: () => listLessonSections(),
	});
}

export function useCreateLessonSection() {
	const invalidate = useInvalidate(keys.sections);
	return useMutation({
		mutationFn: (payload: CreateLessonSection) => createLessonSection(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateLessonSection() {
	const invalidate = useInvalidate(keys.sections);
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateLessonSection;
		}) => updateLessonSection(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteLessonSection() {
	const invalidate = useInvalidate(keys.sections);
	return useMutation({
		mutationFn: (id: string) => deleteLessonSection(id),
		onSuccess: invalidate,
	});
}

export function useAdminLessonActivities(lessonId: string) {
	return useQuery({
		queryKey: keys.activities(lessonId),
		queryFn: () => listLessonActivities(lessonId),
		enabled: lessonId.length > 0,
	});
}

export function useAdminLessonActivity(id: string | null) {
	return useQuery({
		queryKey: keys.activity(id ?? "none"),
		queryFn: () => getAdminLessonActivity(id ?? ""),
		enabled: Boolean(id),
	});
}

export function useCreateLessonActivity(lessonId: string) {
	const invalidate = useInvalidate(keys.activities(lessonId));
	return useMutation({
		mutationFn: (payload: CreateLessonActivity) =>
			createLessonActivity(payload),
		onSuccess: invalidate,
	});
}

export function useUpdateLessonActivity(lessonId: string) {
	const invalidate = useInvalidate(keys.activities(lessonId));
	return useMutation({
		mutationFn: ({
			id,
			payload,
		}: {
			id: string;
			payload: UpdateLessonActivity;
		}) => updateLessonActivity(id, payload),
		onSuccess: invalidate,
	});
}

export function useDeleteLessonActivity(lessonId: string) {
	const invalidate = useInvalidate(keys.activities(lessonId));
	return useMutation({
		mutationFn: (id: string) => deleteLessonActivity(id),
		onSuccess: invalidate,
	});
}
