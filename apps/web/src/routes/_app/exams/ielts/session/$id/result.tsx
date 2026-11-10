import type {
	ExamReviewAttempt,
	IeltsReviewQuestion,
	IeltsReviewQuestionGroup,
	IeltsReviewSection,
} from "@engflex/contracts";
import {
	createFileRoute,
	Link,
	notFound,
	redirect,
} from "@tanstack/react-router";
import {
	Award,
	CheckCircle2,
	Clock,
	RotateCcw,
	Target,
	XCircle,
} from "lucide-react";
import { useState } from "react";
import { z } from "zod";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
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
						error instanceof ApiError && [401, 404, 409].includes(error.status)
					) && count < 3,
			});
			if (attempt.status !== "completed")
				throw redirect({
					to: APP_ROUTES.EXAMS.IELTS.SESSION,
					params: { id },
					replace: true,
				});
			return attempt as unknown as ExamReviewAttempt;
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
			<div className="grid gap-4 md:grid-cols-4">
				{[1, 2, 3, 4].map((i) => (
					<CardShell
						key={i}
						pending
						className="p-5"
						skeleton={<Skeleton className="h-28 w-full" />}
					/>
				))}
			</div>
			<CardShell
				pending
				className="p-6"
				skeleton={<Skeleton className="h-96 w-full" />}
			/>
		</PageLayout>
	);
}

function formatDuration(seconds: number): string {
	const mm = Math.floor(seconds / 60);
	const ss = seconds % 60;
	return `${mm}m ${ss}s`;
}

function IeltsExamResultPage() {
	const attempt = Route.useLoaderData();
	const [filter, setFilter] = useState<"all" | "correct" | "incorrect">("all");

	useBreadcrumbs([
		{ label: m["exams.ielts.hub.title"](), to: APP_ROUTES.EXAMS.IELTS.LIST },
		{
			label: m["exams.ielts.history.title"](),
			to: APP_ROUTES.EXAMS.IELTS.HISTORY,
		},
		{ label: m["exams.ielts.result.title"]() },
	]);

	const ieltsResult = attempt.ieltsResult;
	const ieltsContent = attempt.examTest?.ieltsContent;

	const bandScore = ieltsResult?.bandScore ?? attempt.score ?? 0.0;
	const correctCount = ieltsResult?.correctCount ?? 0;
	const questionCount =
		(ieltsContent?.sections ?? []).reduce(
			(acc: number, s: IeltsReviewSection) =>
				acc +
				s.questionGroups.reduce(
					(gAcc: number, g: IeltsReviewQuestionGroup) =>
						gAcc + g.questions.length,
					0,
				),
			0,
		) ||
		ieltsResult?.answers?.length ||
		40;
	const accuracy = Math.round(
		ieltsResult?.accuracyPct ?? (correctCount / (questionCount || 1)) * 100,
	);
	const durationSec = attempt.durationSec ?? 0;

	const answersMap = new Map<string, string>();
	for (const a of ieltsResult?.answers ?? []) {
		if (a.selectedKey) answersMap.set(a.questionId, a.selectedKey);
		else if (a.textValue) answersMap.set(a.questionId, a.textValue);
		else if (a.selectedKeys && a.selectedKeys.length > 0)
			answersMap.set(a.questionId, a.selectedKeys.join(", "));
	}

	const questions = (ieltsContent?.sections ?? []).flatMap(
		(s: IeltsReviewSection) =>
			s.questionGroups.flatMap((g: IeltsReviewQuestionGroup) =>
				g.questions.map((q: IeltsReviewQuestion) => {
					const userAns = answersMap.get(q.id) ?? "";
					let correctAns = "";
					let isCorrect = false;

					if (q.correctKey) {
						correctAns = q.correctKey;
						isCorrect =
							userAns.trim().toUpperCase() === q.correctKey.toUpperCase();
					} else if (q.correctAnswers && q.correctAnswers.length > 0) {
						correctAns = q.correctAnswers.join(" / ");
						const cleanUser = userAns.trim().toLowerCase();
						isCorrect = q.correctAnswers.some(
							(ans: string) => ans.trim().toLowerCase() === cleanUser,
						);
					} else if (q.correctKeys && q.correctKeys.length > 0) {
						correctAns = q.correctKeys.join(", ");
						isCorrect =
							userAns
								.split(",")
								.map((k: string) => k.trim().toUpperCase())
								.sort()
								.join(",") ===
							q.correctKeys
								.map((k: string) => k.trim().toUpperCase())
								.sort()
								.join(",");
					}

					return {
						id: q.id,
						number: q.questionNumber,
						stem: q.stem,
						userAnswer: userAns,
						correctAnswer: correctAns,
						isCorrect,
						explanation: q.vietnameseExplanation ?? q.explanation,
					};
				}),
			),
	);

	const filteredQuestions = questions.filter((q) => {
		if (filter === "correct") return q.isCorrect;
		if (filter === "incorrect") return !q.isCorrect;
		return true;
	});

	return (
		<PageLayout>
			<div className="flex flex-col gap-6">
				<div>
					<BackButton
						to={APP_ROUTES.EXAMS.IELTS.LIST}
						label={m["exams.ielts.history.backToHub"]()}
					/>
					<h1 className="mt-3 text-2xl font-bold tracking-tight text-foreground md:text-3xl">
						{m["exams.ielts.result.title"]()}
					</h1>
					<p className="text-sm text-muted-foreground">
						{attempt.examTest?.title}
					</p>
				</div>

				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
					<div className="flex items-center gap-4 rounded-2xl border border-amber-500/30 bg-gradient-to-br from-amber-500/10 to-orange-500/10 p-5 shadow-xs">
						<div className="flex size-12 shrink-0 items-center justify-center rounded-xl bg-amber-500 text-white shadow-sm">
							<Award className="size-6" />
						</div>
						<div>
							<span className="text-xs font-semibold text-muted-foreground uppercase">
								{m["exams.ielts.result.bandScore"]()}
							</span>
							<h3 className="text-2xl font-black text-amber-600 dark:text-amber-400">
								{bandScore.toFixed(1)}
							</h3>
						</div>
					</div>

					<div className="flex items-center gap-4 rounded-2xl border border-border/80 bg-card p-5 shadow-xs">
						<div className="flex size-12 shrink-0 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
							<CheckCircle2 className="size-6" />
						</div>
						<div>
							<span className="text-xs font-semibold text-muted-foreground uppercase">
								Câu đúng
							</span>
							<h3 className="text-2xl font-bold text-foreground">
								{correctCount} / {questionCount}
							</h3>
						</div>
					</div>

					<div className="flex items-center gap-4 rounded-2xl border border-border/80 bg-card p-5 shadow-xs">
						<div className="flex size-12 shrink-0 items-center justify-center rounded-xl bg-sky-500/10 text-sky-600 dark:text-sky-400">
							<Target className="size-6" />
						</div>
						<div>
							<span className="text-xs font-semibold text-muted-foreground uppercase">
								Độ chính xác
							</span>
							<h3 className="text-2xl font-bold text-foreground">
								{accuracy}%
							</h3>
						</div>
					</div>

					<div className="flex items-center gap-4 rounded-2xl border border-border/80 bg-card p-5 shadow-xs">
						<div className="flex size-12 shrink-0 items-center justify-center rounded-xl bg-purple-500/10 text-purple-600 dark:text-purple-400">
							<Clock className="size-6" />
						</div>
						<div>
							<span className="text-xs font-semibold text-muted-foreground uppercase">
								Thời gian làm bài
							</span>
							<h3 className="text-2xl font-bold text-foreground">
								{formatDuration(durationSec)}
							</h3>
						</div>
					</div>
				</div>

				<div className="flex flex-col gap-4 rounded-2xl border border-border/80 bg-card p-6 shadow-sm">
					<div className="flex flex-wrap items-center justify-between gap-4 border-b border-border/60 pb-4">
						<h2 className="text-base font-bold text-foreground">
							{m["exams.ielts.result.reviewTitle"]()}
						</h2>

						<div className="flex items-center gap-1.5">
							<Button
								variant={filter === "all" ? "default" : "outline"}
								size="sm"
								onClick={() => setFilter("all")}
								className="rounded-full text-xs"
							>
								{m["exams.ielts.result.filterAll"]()} ({questions.length})
							</Button>
							<Button
								variant={filter === "correct" ? "default" : "outline"}
								size="sm"
								onClick={() => setFilter("correct")}
								className="rounded-full text-xs"
							>
								{m["exams.ielts.result.filterCorrect"]()} ({correctCount})
							</Button>
							<Button
								variant={filter === "incorrect" ? "default" : "outline"}
								size="sm"
								onClick={() => setFilter("incorrect")}
								className="rounded-full text-xs"
							>
								{m["exams.ielts.result.filterIncorrect"]()} (
								{questions.length - correctCount})
							</Button>
						</div>
					</div>

					<div className="flex flex-col divide-y divide-border/60">
						{filteredQuestions.map((q) => (
							<div key={q.id} className="py-4 space-y-2">
								<div className="flex items-start gap-2.5">
									{q.isCorrect ? (
										<CheckCircle2 className="size-5 shrink-0 text-emerald-500 mt-0.5" />
									) : (
										<XCircle className="size-5 shrink-0 text-red-500 mt-0.5" />
									)}
									<div className="flex-1 space-y-1">
										<p className="text-sm font-bold text-foreground">
											Câu {q.number}: {q.stem}
										</p>
										<div className="flex flex-wrap items-center gap-4 text-xs">
											<span
												className={
													q.isCorrect
														? "font-semibold text-emerald-600 dark:text-emerald-400"
														: "font-semibold text-red-600 dark:text-red-400"
												}
											>
												{m["exams.ielts.result.yourAnswer"]({
													answer: q.userAnswer || "(Chưa trả lời)",
												})}
											</span>
											{!q.isCorrect && (
												<span className="font-semibold text-emerald-600 dark:text-emerald-400">
													{m["exams.ielts.result.correctAnswer"]({
														answer: q.correctAnswer,
													})}
												</span>
											)}
										</div>
										{q.explanation && (
											<div className="mt-2 rounded-lg bg-muted/50 p-2.5 text-xs text-muted-foreground whitespace-pre-line">
												<span className="font-semibold text-foreground">
													{m["exams.ielts.result.explanation"]()}{" "}
												</span>
												{q.explanation}
											</div>
										)}
									</div>
								</div>
							</div>
						))}
					</div>
				</div>

				<div className="flex flex-wrap items-center justify-end gap-3 pt-4">
					<Link
						to={APP_ROUTES.EXAMS.IELTS.LIST}
						className="rounded-xl border border-border/80 px-4 py-2 text-sm font-semibold text-foreground hover:bg-muted"
					>
						{m["exams.ielts.result.closeButton"]()}
					</Link>
					{attempt.examTest && (
						<Link
							to={APP_ROUTES.EXAMS.IELTS.TEST}
							params={{
								categoryId: attempt.examTest.categoryId,
								testId: attempt.examTest.id,
							}}
							className="inline-flex items-center gap-2 rounded-xl bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground shadow-sm hover:bg-primary/90"
						>
							<RotateCcw className="size-4" />
							<span>{m["exams.ielts.result.retakeButton"]()}</span>
						</Link>
					)}
				</div>
			</div>
		</PageLayout>
	);
}
