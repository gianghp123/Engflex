import { useAuth } from "@clerk/tanstack-react-start";
import type { ExamAttemptSummary, Pagination } from "@engflex/contracts";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { createFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { History, TrendingUp } from "lucide-react";
import { useRef, useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { CardShell } from "#/components/common/card-shell";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageHero } from "#/components/common/page-layout";
import { SubmitButton } from "#/components/common/submit-button";
import {
	AlertDialog,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "#/components/ui/alert-dialog";
import { Skeleton } from "#/components/ui/skeleton";
import {
	ToeicCategoryCard,
	type ToeicCategoryCardItem,
} from "#/features/exams/components/toeic-category-card";
import {
	type ToeicAttemptItem,
	ToeicHistoryPanel,
} from "#/features/exams/components/toeic-history-panel";
import { ToeicStatsPanel } from "#/features/exams/components/toeic-stats-panel";
import {
	examAttemptsQueryOptions,
	examCategoriesQueryOptions,
	examKeys,
	examStatisticsQueryOptions,
	useCancelExamAttempt,
} from "#/features/exams/queries";
import { ApiError } from "#/lib/api/api";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";

const CATEGORY_COLORS = [
	"bg-[#0284c7]",
	"bg-[#ea580c]",
	"bg-[#9333ea]",
	"bg-[#2563eb]",
];

function formatAttemptDate(iso: string | null | undefined): string {
	if (!iso) return "";
	return new Intl.DateTimeFormat(getLocale(), {
		day: "numeric",
		month: "short",
	}).format(new Date(iso));
}

function CategoryCardSkeleton() {
	return (
		<div className="flex items-center justify-between gap-4 rounded-2xl border border-border/80 bg-card p-5 shadow-sm">
			<div className="flex min-w-0 flex-1 flex-col gap-3">
				<Skeleton className="h-6 w-3/4 rounded-md" />
				<div className="flex gap-1.5">
					<Skeleton className="h-4 w-14 rounded-md" />
					<Skeleton className="h-4 w-14 rounded-md" />
				</div>
				<div className="flex gap-4">
					<Skeleton className="h-4 w-20 rounded-md" />
					<Skeleton className="h-4 w-24 rounded-md" />
				</div>
				<Skeleton className="h-4 w-28 rounded-md" />
			</div>
			<Skeleton className="size-14 shrink-0 rounded-2xl md:size-16" />
		</div>
	);
}

function HistoryPanelSkeleton() {
	return (
		<div className="flex flex-col gap-3">
			<div className="flex items-center justify-between">
				<Skeleton className="h-6 w-32 rounded-md" />
				<Skeleton className="h-4 w-16 rounded-md" />
			</div>
			<div className="flex flex-col divide-y divide-border/60 rounded-2xl border border-border/80 bg-card p-4 shadow-sm">
				{[1, 2].map((i) => (
					<div key={i} className="flex items-center justify-between py-3">
						<div className="flex items-center gap-3">
							<Skeleton className="size-9 rounded-full" />
							<div className="flex flex-col gap-2">
								<Skeleton className="h-4 w-40 rounded-md" />
								<Skeleton className="h-3 w-20 rounded-md" />
							</div>
						</div>
						<Skeleton className="h-8 w-24 rounded-lg" />
					</div>
				))}
			</div>
		</div>
	);
}

export const Route = createFileRoute("/_app/exams/toeic/")({
	component: ToeicHubPage,
});

function retryHubRead(count: number, error: unknown) {
	return (
		!(error instanceof ApiError && [401, 404, 409].includes(error.status)) &&
		count < 3
	);
}

function HubQueryError({
	message,
	pending,
	onRetry,
}: {
	message: string;
	pending: boolean;
	onRetry: () => void;
}) {
	return (
		<CardShell className="col-span-full flex flex-col items-start gap-3 p-5">
			<p role="alert" className="text-sm text-destructive">
				{message}
			</p>
			<SubmitButton pending={pending} onClick={onRetry} variant="outline">
				{m["exams.toeic.hub.load.retry"]()}
			</SubmitButton>
		</CardShell>
	);
}

function ToeicHubPage() {
	useBreadcrumbs([{ label: m["exams.toeic.hub.title"]() }]);
	const { isLoaded } = useAuth();
	const queryClient = useQueryClient();
	const navigate = useNavigate();
	const cancel = useCancelExamAttempt();
	const cancelLock = useRef(false);
	const [attemptToCancel, setAttemptToCancel] = useState<string | null>(null);
	const [cancelError, setCancelError] = useState(false);
	const categoriesQuery = useQuery({
		...examCategoriesQueryOptions({ type: "toeic" }),
		enabled: isLoaded,
		retry: retryHubRead,
	});
	const attemptsQuery = useQuery({
		...examAttemptsQueryOptions({ page: 1, pageSize: 5 }),
		enabled: isLoaded,
		retry: retryHubRead,
	});

	const statsQuery = useQuery({
		...examStatisticsQueryOptions(),
		enabled: isLoaded,
		retry: retryHubRead,
	});

	async function handleCancel() {
		if (!attemptToCancel || cancelLock.current) return;
		cancelLock.current = true;
		setCancelError(false);
		try {
			const response = await cancel.mutateAsync(attemptToCancel);
			if (response.status !== "abandoned" && response.status !== "completed")
				throw new Error("Attempt was not closed");
			queryClient.setQueriesData<{
				items: ExamAttemptSummary[];
				pagination: Pagination;
			}>({ queryKey: [...examKeys.all, "attempts"] }, (cached) =>
				cached
					? {
							...cached,
							items: cached.items.map((item) =>
								item.id === response.id ? response : item,
							),
						}
					: cached,
			);
			setAttemptToCancel(null);
			if (response.status === "abandoned")
				toast.success(m["exams.toeic.hub.history.cancelSuccess"]());
			else
				toast.info(m["exams.toeic.history.alreadyCompleted"](), {
					action: {
						label: m["exams.toeic.hub.history.viewResult"](),
						onClick: () => {
							void navigate({
								to: APP_ROUTES.EXAMS.TOEIC.RESULT,
								params: { id: response.id },
							});
						},
					},
				});
			void queryClient.invalidateQueries({
				queryKey: [...examKeys.all, "attempts"],
			});
			void queryClient.invalidateQueries({
				queryKey: examKeys.attempt(response.id),
			});
		} catch {
			setCancelError(true);
		} finally {
			cancelLock.current = false;
		}
	}

	const categories: ToeicCategoryCardItem[] = (categoriesQuery.data ?? []).map(
		(cat, idx) => ({
			id: cat.id,
			title: cat.title,
			tags: [cat.type.toUpperCase()],
			testsCount: cat.testsCount,
			attemptsCount: cat.attemptsCount,
			iconBgColor: CATEGORY_COLORS[idx % CATEGORY_COLORS.length],
		}),
	);

	const attempts: ToeicAttemptItem[] = (attemptsQuery.data?.items ?? []).map(
		(att) => ({
			id: att.id,
			title: att.examTest?.title ?? "TOEIC Test",
			mode: (att.mode ?? "exam") as "exam" | "practice",
			status: att.status,
			date: formatAttemptDate(att.startedAt ?? att.createdAt),
			scorePct: att.score != null ? Math.round(att.score) : undefined,
		}),
	);

	return (
		<div className="mx-auto flex w-full max-w-6xl flex-col gap-8 pb-12">
			<div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
				<PageHero
					icon={<MoMascot variant="nice" size={84} />}
					title={m["exams.toeic.hub.title"]()}
					description={m["exams.toeic.hub.description"]()}
				/>

				<div className="flex flex-wrap items-center gap-2.5 shrink-0 pt-2 md:pt-4">
					<Link
						to={APP_ROUTES.EXAMS.TOEIC.HISTORY}
						className="inline-flex items-center gap-2 rounded-xl bg-sky-600 px-4 py-2 text-xs font-bold text-white shadow-sm transition-colors hover:bg-sky-700"
					>
						<History className="size-4" aria-hidden="true" />
						<span>{m["exams.toeic.hub.historyButton"]()}</span>
					</Link>
					<button
						type="button"
						className="inline-flex items-center gap-2 rounded-xl bg-emerald-600 px-4 py-2 text-xs font-bold text-white shadow-sm transition-colors hover:bg-emerald-700"
					>
						<TrendingUp className="size-4" aria-hidden="true" />
						<span>{m["exams.toeic.hub.analyticsButton"]()}</span>
					</button>
				</div>
			</div>

			<div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
				{categoriesQuery.isError ? (
					<HubQueryError
						message={m["exams.toeic.hub.load.categoriesFailed"]()}
						pending={categoriesQuery.isFetching}
						onRetry={() => {
							void categoriesQuery.refetch();
						}}
					/>
				) : categoriesQuery.isPending ? (
					<>
						<CategoryCardSkeleton />
						<CategoryCardSkeleton />
						<CategoryCardSkeleton />
						<CategoryCardSkeleton />
					</>
				) : (
					categories.map((category) => (
						<ToeicCategoryCard key={category.id} item={category} />
					))
				)}
			</div>

			<div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
				{statsQuery.isError ? (
					<HubQueryError
						message={m["exams.toeic.hub.load.statisticsFailed"]()}
						pending={statsQuery.isFetching}
						onRetry={() => {
							void statsQuery.refetch();
						}}
					/>
				) : (
					<ToeicStatsPanel
						stats={statsQuery.data}
						pending={statsQuery.isPending}
					/>
				)}
				{attemptsQuery.isError ? (
					<HubQueryError
						message={m["exams.toeic.hub.load.attemptsFailed"]()}
						pending={attemptsQuery.isFetching}
						onRetry={() => {
							void attemptsQuery.refetch();
						}}
					/>
				) : attemptsQuery.isPending ? (
					<HistoryPanelSkeleton />
				) : (
					<ToeicHistoryPanel
						attempts={attempts}
						locked={cancel.isPending}
						onCancel={(id) => {
							if (!cancelLock.current) {
								setCancelError(false);
								setAttemptToCancel(id);
							}
						}}
					/>
				)}
			</div>
			<AlertDialog
				open={attemptToCancel !== null}
				onOpenChange={(open) => {
					if (!open && !cancelLock.current) setAttemptToCancel(null);
				}}
			>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>
							{m["exams.toeic.hub.history.cancelDialogTitle"]()}
						</AlertDialogTitle>
						<AlertDialogDescription>
							{m["exams.toeic.hub.history.cancelDialogDesc"]()}
						</AlertDialogDescription>
					</AlertDialogHeader>
					{cancelError && (
						<p role="alert" className="text-sm text-destructive">
							{m["exams.toeic.history.cancelFailed"]()}
						</p>
					)}
					<AlertDialogFooter>
						<AlertDialogCancel disabled={cancel.isPending}>
							{m["exams.toeic.hub.history.cancelDialogCancel"]()}
						</AlertDialogCancel>
						<SubmitButton
							variant="destructive"
							pending={cancel.isPending}
							onClick={() => {
								void handleCancel();
							}}
						>
							{m["exams.toeic.hub.history.cancelDialogConfirm"]()}
						</SubmitButton>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</div>
	);
}
