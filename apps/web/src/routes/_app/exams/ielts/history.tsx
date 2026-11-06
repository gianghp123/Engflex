import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import { Award, Calendar, CheckCircle2 } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { CardShell } from "#/components/common/card-shell";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { Pager } from "#/components/common/pager";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import {
	examAttemptsQueryOptions,
	examKeys,
	useCancelExamAttempt,
} from "#/features/exams/queries";
import { loadOr404 } from "#/lib/router/route-loader";
import { parsePage } from "#/lib/router/search-params";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";

const PAGE_SIZE = 10;

export const Route = createFileRoute("/_app/exams/ielts/history")({
	validateSearch: (search: Record<string, unknown>): { page?: number } => ({
		page: parsePage(search.page),
	}),
	loaderDeps: ({ search }) => ({ page: search.page ?? 1 }),
	loader: async ({ context: { queryClient }, deps: { page } }) => {
		const data = await loadOr404(
			queryClient,
			examAttemptsQueryOptions({ page, pageSize: PAGE_SIZE }),
		);
		const last = Math.max(1, data.pagination.totalPages);
		if (page > last)
			throw redirect({
				to: APP_ROUTES.EXAMS.IELTS.HISTORY,
				search: { page: last },
				replace: true,
			});
		return data;
	},
	pendingComponent: HistoryPending,
	component: IeltsHistoryPage,
});

function HistoryPending() {
	return (
		<PageLayout>
			{[1, 2, 3].map((i) => (
				<CardShell
					key={i}
					pending
					className="p-5"
					skeleton={
						<>
							<Skeleton className="h-5 w-3/4" />
							<Skeleton className="mt-3 h-4 w-1/2" />
							<Skeleton className="mt-4 ml-auto h-9 w-36" />
						</>
					}
				/>
			))}
		</PageLayout>
	);
}

function formatDate(iso: string | null | undefined): string {
	if (!iso) return "";
	return new Intl.DateTimeFormat(getLocale(), {
		day: "numeric",
		month: "short",
		year: "numeric",
	}).format(new Date(iso));
}

function IeltsHistoryPage() {
	const { items, pagination } = Route.useLoaderData();
	const navigate = Route.useNavigate();
	const queryClient = useQueryClient();
	const cancel = useCancelExamAttempt();
	const [cancellingId, setCancellingId] = useState<string | null>(null);

	useBreadcrumbs([
		{ label: m["exams.ielts.hub.title"](), to: APP_ROUTES.EXAMS.IELTS.LIST },
		{ label: m["exams.ielts.history.title"]() },
	]);

	const ieltsAttempts = items.filter(
		(a) => a.ieltsResult !== null || a.examTest?.category?.type === "ielts",
	);

	const handleCancel = async (attemptId: string) => {
		setCancellingId(attemptId);
		try {
			await cancel.mutateAsync(attemptId);
			await queryClient.invalidateQueries({ queryKey: examKeys.attempts() });
			toast.success(m["exams.ielts.history.cancel"]());
		} catch {
			toast.error("Không thể hủy bài thi.");
		} finally {
			setCancellingId(null);
		}
	};

	return (
		<PageLayout>
			<div className="flex flex-col gap-6">
				<div>
					<BackButton
						to={APP_ROUTES.EXAMS.IELTS.LIST}
						label={m["exams.ielts.history.backToHub"]()}
					/>
					<h1 className="mt-3 text-2xl font-bold tracking-tight text-foreground md:text-3xl">
						{m["exams.ielts.history.title"]()}
					</h1>
				</div>

				{ieltsAttempts.length === 0 ? (
					<div className="flex flex-col items-center justify-center gap-4 rounded-3xl border border-dashed border-border/80 bg-card/50 p-12 text-center">
						<MoMascot variant="thinking" className="size-24" />
						<div className="flex flex-col gap-1">
							<h3 className="text-lg font-bold text-foreground">
								{m["exams.ielts.history.emptyTitle"]()}
							</h3>
							<p className="max-w-md text-sm text-muted-foreground">
								{m["exams.ielts.history.emptyDesc"]()}
							</p>
						</div>
						<Link
							to={APP_ROUTES.EXAMS.IELTS.LIST}
							className="mt-2 inline-flex items-center justify-center rounded-xl bg-primary px-5 py-2.5 text-sm font-bold text-primary-foreground shadow-sm transition-colors hover:bg-primary/90"
						>
							{m["exams.ielts.history.emptyAction"]()}
						</Link>
					</div>
				) : (
					<div className="flex flex-col gap-3.5">
						{ieltsAttempts.map((attempt) => {
							const band = attempt.score;
							const correct = attempt.ieltsResult?.correctCount ?? 0;
							const total = attempt.ieltsResult?.questionCount ?? 40;

							return (
								<div
									key={attempt.id}
									className="flex flex-col gap-4 rounded-2xl border border-border/80 bg-card p-5 shadow-xs transition-all hover:border-border sm:flex-row sm:items-center sm:justify-between"
								>
									<div className="flex min-w-0 flex-1 flex-col gap-2">
										<div className="flex flex-wrap items-center gap-2">
											<h3 className="truncate text-base font-bold text-card-foreground">
												{attempt.examTest?.title ?? "IELTS Test"}
											</h3>
											<span
												className={`rounded-md px-2 py-0.5 text-xs font-semibold uppercase ${
													attempt.status === "completed"
														? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
														: attempt.status === "in_progress"
															? "bg-amber-500/10 text-amber-600 dark:text-amber-400"
															: "bg-muted text-muted-foreground"
												}`}
											>
												{attempt.status === "completed"
													? m["exams.ielts.history.statusCompleted"]()
													: attempt.status === "in_progress"
														? m["exams.ielts.history.statusInProgress"]()
														: m["exams.ielts.history.statusAbandoned"]()}
											</span>
										</div>

										<div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground">
											<span className="inline-flex items-center gap-1.5">
												<Calendar className="size-3.5" />
												{formatDate(attempt.startedAt)}
											</span>
											{attempt.status === "completed" && band != null && (
												<span className="inline-flex items-center gap-1.5 font-bold text-amber-600 dark:text-amber-400">
													<Award className="size-3.5" />
													{m["exams.ielts.history.bandScore"]({
														score: band.toFixed(1),
													})}
												</span>
											)}
											{attempt.status === "completed" && (
												<span className="inline-flex items-center gap-1.5">
													<CheckCircle2 className="size-3.5" />
													{m["exams.ielts.history.scoreCount"]({
														correct,
														total,
													})}
												</span>
											)}
										</div>
									</div>

									<div className="flex items-center gap-2 self-end sm:self-center">
										{attempt.status === "completed" ? (
											<Link
												to={APP_ROUTES.EXAMS.IELTS.RESULT}
												params={{ id: attempt.id }}
												className="rounded-xl border border-border/80 bg-background px-4 py-2 text-xs font-bold text-foreground transition-colors hover:bg-muted"
											>
												{m["exams.ielts.history.viewResult"]()}
											</Link>
										) : attempt.status === "in_progress" ? (
											<>
												<Link
													to={APP_ROUTES.EXAMS.IELTS.SESSION}
													params={{ id: attempt.id }}
													className="rounded-xl bg-primary px-4 py-2 text-xs font-bold text-primary-foreground shadow-xs hover:bg-primary/90"
												>
													{m["exams.ielts.history.continue"]()}
												</Link>
												<Button
													variant="ghost"
													size="sm"
													disabled={cancellingId === attempt.id}
													onClick={() => void handleCancel(attempt.id)}
													className="rounded-xl text-xs text-muted-foreground hover:text-destructive"
												>
													{m["exams.ielts.history.cancel"]()}
												</Button>
											</>
										) : null}
									</div>
								</div>
							);
						})}
					</div>
				)}

				{pagination.totalPages > 1 && (
					<div className="mt-4 flex justify-center">
						<Pager
							page={pagination.page}
							totalPages={pagination.totalPages}
							onPageChange={(page) => {
								void navigate({
									search: { page },
								});
							}}
						/>
					</div>
				)}
			</div>
		</PageLayout>
	);
}
