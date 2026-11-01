import { useAuth } from "@clerk/tanstack-react-start";
import type { ExamReviewAttempt } from "@engflex/contracts";
import { createFileRoute, notFound, redirect } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { z } from "zod";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { Skeleton } from "#/components/ui/skeleton";
import { clearAudioPosition } from "#/features/exams/audio-position";
import { ToeicResultHeaderCard } from "#/features/exams/components/result/toeic-result-header-card";
import { ToeicResultPaletteCard } from "#/features/exams/components/result/toeic-result-palette-card";
import { ToeicResultPartsCard } from "#/features/exams/components/result/toeic-result-parts-card";
import { ToeicResultReviewSection } from "#/features/exams/components/result/toeic-result-review-section";
import { examAttemptQueryOptions } from "#/features/exams/queries";
import { partStatistics, type ReviewFilter } from "#/features/exams/review";
import { isExamReviewAttempt } from "#/features/exams/service";
import { ApiError } from "#/lib/api/api";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/exams/toeic/session/$id/result")({
	loader: async ({ context: { queryClient }, params: { id } }) => {
		if (!z.uuid().safeParse(id).success) throw notFound();
		try {
			const attempt = await queryClient.query({
				...examAttemptQueryOptions(id),
				retry: (count: number, error: unknown) =>
					!(
						error instanceof ApiError && [401, 404, 409].includes(error.status)
					) && count < 3,
			});
			if (!isExamReviewAttempt(attempt))
				throw redirect({
					to: APP_ROUTES.EXAMS.TOEIC.SESSION,
					params: { id },
					replace: true,
				});
			if (
				!attempt.examTest?.content?.parts.length ||
				!attempt.result ||
				attempt.result.correctCount === null ||
				attempt.result.accuracyPct === null
			)
				throw new Error("Invalid completed exam result");
			return attempt;
		} catch (error) {
			if (error instanceof ApiError) {
				if (error.status === 404) throw notFound();
				if (error.status === 401)
					throw redirect({ href: APP_ROUTES.AUTH.SIGN_IN });
				if (error.status === 409)
					throw redirect({ to: APP_ROUTES.EXAMS.TOEIC.HISTORY, replace: true });
			}
			throw error;
		}
	},
	pendingComponent: ResultPending,
	component: ToeicExamResultPage,
});
function ResultPending() {
	return (
		<PageLayout>
			<div className="grid gap-4 md:grid-cols-3">
				{[0, 1, 2].map((card) => (
					<CardShell
						key={card}
						pending
						className="p-5"
						skeleton={<Skeleton className="h-32 w-full" />}
					/>
				))}
			</div>
			<CardShell
				pending
				className="p-6"
				skeleton={<Skeleton className="h-80 w-full" />}
			/>
			<CardShell
				pending
				className="p-6"
				skeleton={<Skeleton className="h-96 w-full" />}
			/>
		</PageLayout>
	);
}
function ToeicExamResultPage() {
	const attempt = Route.useLoaderData();
	return <ResultContent key={attempt.id} attempt={attempt} />;
}
function ResultContent({ attempt }: { attempt: ExamReviewAttempt }) {
	const { userId } = useAuth();
	useEffect(() => {
		if (userId) clearAudioPosition(userId, attempt.id);
	}, [userId, attempt.id]);
	const content = attempt.examTest?.content;
	const parts = content?.parts ?? [];
	const answers = new Map(
		(attempt.result?.answers ?? []).map((answer) => [
			answer.questionId,
			answer,
		]),
	);
	const [activePart, setActivePart] = useState(parts[0]?.partNumber ?? 1);
	const [selectedFilter, setSelectedFilter] = useState<ReviewFilter>("all");
	const [focusedQuestion, setFocusedQuestion] = useState<string | null>(null);
	useBreadcrumbs([
		{ label: m["exams.toeic.hub.title"](), to: APP_ROUTES.EXAMS.TOEIC.LIST },
		{
			label: m["exams.toeic.hub.history.title"](),
			to: APP_ROUTES.EXAMS.TOEIC.HISTORY,
		},
		{ label: attempt.examTest?.title ?? m["exams.toeic.result.title"]() },
	]);
	return (
		<PageLayout
			action={
				<BackButton
					to={APP_ROUTES.EXAMS.TOEIC.HISTORY}
					label={m["exams.toeic.result.back"]()}
				/>
			}
		>
			<h1 className="text-2xl font-bold">{m["exams.toeic.result.title"]()}</h1>
			<ToeicResultHeaderCard attempt={attempt} answers={answers} />
			<ToeicResultPartsCard
				parts={parts.map((part) => partStatistics(part, answers))}
			/>
			<ToeicResultPaletteCard
				parts={parts}
				answers={answers}
				onNavigateQuestion={(id, part) => {
					if (
						id === focusedQuestion &&
						part === activePart &&
						selectedFilter === "all"
					) {
						document
							.getElementById(id)
							?.scrollIntoView({ behavior: "smooth", block: "start" });
					}
					setActivePart(part);
					setSelectedFilter("all");
					setFocusedQuestion(id);
				}}
			/>
			<ToeicResultReviewSection
				parts={parts}
				assets={content?.assets ?? []}
				answers={answers}
				activePart={activePart}
				selectedFilter={selectedFilter}
				focusedQuestion={focusedQuestion}
				onSelectPart={(part) => {
					setActivePart(part);
					setFocusedQuestion(null);
				}}
				onSelectFilter={(filter) => {
					setSelectedFilter(filter);
					setFocusedQuestion(null);
				}}
			/>
		</PageLayout>
	);
}
