import type { ExamReviewAttempt } from "@engflex/contracts";
import { createFileRoute, notFound, redirect } from "@tanstack/react-router";
import { z } from "zod";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { StatCard } from "#/components/common/stat-card";
import { Skeleton } from "#/components/ui/skeleton";
import { IeltsReview } from "#/features/exams/components/ielts/ielts-review";
import { examAttemptQueryOptions } from "#/features/exams/queries";
import { ApiError } from "#/lib/api/api";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/exams/ielts/session/$id/result")({
	loader: async ({ context: { queryClient }, params: { id } }) => {
		if (!z.uuid().safeParse(id).success) throw notFound();
		try {
			const attempt = await queryClient.query({
				...examAttemptQueryOptions(id),
				retry: (count: number, error: unknown) =>
					!(
						error instanceof ApiError &&
						[401, 403, 404, 409, 429].includes(error.status)
					) && count < 3,
			});
			if (attempt.status !== "completed")
				throw redirect({
					to: APP_ROUTES.EXAMS.IELTS.SESSION,
					params: { id },
					replace: true,
				});
			if (attempt.examTest?.category?.type !== "ielts") throw notFound();
			return attempt as ExamReviewAttempt;
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
	pendingComponent: ResultPending,
	component: IeltsExamResultPage,
});

function ResultPending() {
	return (
		<PageLayout>
			<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
				<StatCard
					pending
					title={m["exams.ielts.result.bandScore"]()}
					value=""
				/>
				<StatCard
					pending
					title={m["exams.ielts.result.metrics.correct"]()}
					value=""
				/>
				<StatCard
					pending
					title={m["exams.ielts.result.metrics.accuracy"]()}
					value=""
				/>
				<StatCard
					pending
					title={m["exams.ielts.result.metrics.time"]()}
					value=""
				/>
			</div>
			<CardShell
				pending
				className="p-6"
				skeleton={<Skeleton className="h-96 w-full" />}
			/>
		</PageLayout>
	);
}

function IeltsExamResultPage() {
	const attempt = Route.useLoaderData();
	useBreadcrumbs([
		{ label: m["exams.ielts.hub.title"](), to: APP_ROUTES.EXAMS.IELTS.LIST },
		{ label: m["exams.ielts.result.title"]() },
	]);
	return <IeltsReview attempt={attempt} />;
}
