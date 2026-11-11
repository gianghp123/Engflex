import type { ExamAttemptSummary } from "@engflex/contracts";
import { useQueryClient } from "@tanstack/react-query";
import {
	createFileRoute,
	Link,
	redirect,
	useRouter,
} from "@tanstack/react-router";
import { useRef, useState } from "react";
import { toast } from "sonner";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { Pager } from "#/components/common/pager";
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
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import { ToeicHistoryCard } from "#/features/exams/components/history/toeic-history-card";
import {
	examAttemptsQueryOptions,
	examKeys,
	useCancelExamAttempt,
	useContinueExamAttempt,
} from "#/features/exams/queries";
import { ApiError } from "#/lib/api/api";
import { loadOr404 } from "#/lib/router/route-loader";
import { parsePage } from "#/lib/router/search-params";
import { m } from "#/paraglide/messages";

const PAGE_SIZE = 10;
export const Route = createFileRoute("/_app/exams/toeic/history")({
	validateSearch: (search: Record<string, unknown>): { page?: number } => ({
		page: parsePage(search.page),
	}),
	loaderDeps: ({ search }) => ({ page: search.page ?? 1 }),
	loader: async ({ context: { queryClient }, deps: { page } }) => {
		const data = await loadOr404(
			queryClient,
			examAttemptsQueryOptions({ page, pageSize: PAGE_SIZE, type: "toeic" }),
		);
		const last = Math.max(1, data.pagination.totalPages);
		if (page > last)
			throw redirect({
				to: APP_ROUTES.EXAMS.TOEIC.HISTORY,
				search: { page: last },
				replace: true,
			});
		return data;
	},
	pendingComponent: HistoryPending,
	component: ToeicHistoryPage,
});
function HistoryPending() {
	return (
		<PageLayout>
			{[0, 1, 2].map((item) => (
				<CardShell
					key={item}
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
function ToeicHistoryPage() {
	const { items, pagination } = Route.useLoaderData();
	const navigate = Route.useNavigate();
	const queryClient = useQueryClient();
	const router = useRouter();
	const cancel = useCancelExamAttempt();
	const resume = useContinueExamAttempt();
	const actionLock = useRef(false);
	const [attemptToCancel, setAttemptToCancel] =
		useState<ExamAttemptSummary | null>(null);
	useBreadcrumbs([
		{ label: m["exams.toeic.hub.title"](), to: APP_ROUTES.EXAMS.TOEIC.LIST },
		{ label: m["exams.toeic.hub.history.title"]() },
	]);
	async function refresh(id: string) {
		await Promise.all([
			queryClient.invalidateQueries({
				queryKey: [...examKeys.all, "attempts"],
			}),
			queryClient.invalidateQueries({ queryKey: examKeys.attempt(id) }),
		]);
		// Reload the same URL so the route consumes the refreshed server snapshot.
		await router.invalidate();
	}
	async function handleCancel() {
		if (!attemptToCancel || actionLock.current) return;
		actionLock.current = true;
		try {
			const response = await cancel.mutateAsync(attemptToCancel.id);
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
			await refresh(response.id);
		} catch {
			toast.error(m["exams.toeic.history.cancelFailed"]());
		} finally {
			actionLock.current = false;
		}
	}
	async function handleContinue(id: string) {
		if (actionLock.current) return;
		actionLock.current = true;
		try {
			const attempt = await resume.mutateAsync(id);
			queryClient.setQueryData(examKeys.attempt(id), attempt);
			await navigate({
				to:
					attempt.status === "completed"
						? APP_ROUTES.EXAMS.TOEIC.RESULT
						: APP_ROUTES.EXAMS.TOEIC.SESSION,
				params: { id },
			});
		} catch (error) {
			if (error instanceof ApiError && error.status === 409) await refresh(id);
			else toast.error(m["exams.toeic.history.continueFailed"]());
		} finally {
			actionLock.current = false;
		}
	}
	return (
		<PageLayout
			action={
				<BackButton
					to={APP_ROUTES.EXAMS.TOEIC.LIST}
					label={m["exams.toeic.category.back"]()}
				/>
			}
		>
			<h1 className="text-2xl font-bold">
				{m["exams.toeic.hub.history.title"]()}
			</h1>
			{items.length === 0 ? (
				<CardShell className="flex flex-col items-center gap-3 p-12 text-center">
					<h2 className="font-bold">
						{m["exams.toeic.hub.history.emptyTitle"]()}
					</h2>
					<p>{m["exams.toeic.hub.history.emptyDesc"]()}</p>
					<Button asChild>
						<Link to={APP_ROUTES.EXAMS.TOEIC.LIST}>
							{m["exams.toeic.hub.history.emptyAction"]()}
						</Link>
					</Button>
				</CardShell>
			) : (
				<>
					<div className="flex flex-col gap-3.5">
						{items.map((item) => (
							<ToeicHistoryCard
								key={item.id}
								item={item}
								onCancel={setAttemptToCancel}
								onContinue={(id) => {
									void handleContinue(id);
								}}
								pending={resume.isPending && resume.variables === item.id}
								locked={cancel.isPending || resume.isPending}
							/>
						))}
					</div>
					<Pager
						page={pagination.page}
						totalPages={pagination.totalPages}
						onPageChange={(page) => {
							void navigate({ search: { page } });
						}}
						summary={
							pagination.totalPages > 1
								? m["exams.toeic.history.pagination.range"]({
										from: (pagination.page - 1) * PAGE_SIZE + 1,
										to: Math.min(pagination.page * PAGE_SIZE, pagination.total),
										total: pagination.total,
									})
								: m["exams.toeic.history.pagination.all"]({
										count: pagination.total,
									})
						}
					/>
				</>
			)}
			<AlertDialog
				open={!!attemptToCancel}
				onOpenChange={(open) => {
					if (!open && !actionLock.current) setAttemptToCancel(null);
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
		</PageLayout>
	);
}
