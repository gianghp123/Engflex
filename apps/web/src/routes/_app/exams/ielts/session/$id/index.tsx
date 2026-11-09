import { createFileRoute, notFound, redirect } from "@tanstack/react-router";
import { z } from "zod";
import { APP_ROUTES } from "#/app/app-route";
import { IeltsSession } from "#/features/exams/components/ielts/ielts-session";
import { SessionPending } from "#/features/exams/components/taking/toeic-session";
import { examAttemptQueryOptions } from "#/features/exams/queries";
import { ApiError } from "#/lib/api/api";

export const Route = createFileRoute("/_app/exams/ielts/session/$id/")({
	loader: async ({ context: { queryClient }, params: { id } }) => {
		if (!z.uuid().safeParse(id).success) throw notFound();
		try {
			const attempt = await queryClient.query({
				...examAttemptQueryOptions(id),
				staleTime: 0,
				retry: (count: number, error: unknown) =>
					!(
						error instanceof ApiError && [401, 404, 409].includes(error.status)
					) && count < 3,
			});
			if (attempt.status === "completed")
				throw redirect({
					to: APP_ROUTES.EXAMS.IELTS.RESULT,
					params: { id },
					replace: true,
				});
			if (attempt.status !== "in_progress")
				throw redirect({ to: APP_ROUTES.EXAMS.IELTS.HISTORY, replace: true });
			if (
				!attempt.examTest ||
				attempt.examTest.category?.type !== "ielts" ||
				!attempt.startedAt ||
				(attempt.mode === "exam" && !attempt.deadlineAt)
			)
				throw new Error("Invalid IELTS exam session");
			return attempt;
		} catch (error) {
			if (error instanceof ApiError) {
				if (error.status === 404) throw notFound();
				if (error.status === 401)
					throw redirect({ href: APP_ROUTES.AUTH.SIGN_IN });
				if (error.status === 409)
					throw redirect({ to: APP_ROUTES.EXAMS.IELTS.HISTORY, replace: true });
			}
			throw error;
		}
	},
	pendingComponent: SessionPending,
	component: IeltsSessionPage,
});

function IeltsSessionPage() {
	const attempt = Route.useLoaderData();
	return <IeltsSession key={attempt.id} initialAttempt={attempt} />;
}
