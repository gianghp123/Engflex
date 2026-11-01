import type { ExamMode, ExamTestDetail } from "@engflex/contracts";
import { ExamDurationMinutes } from "@engflex/contracts";
import { useQueryClient } from "@tanstack/react-query";
import {
	createFileRoute,
	Link,
	notFound,
	useNavigate,
} from "@tanstack/react-router";
import { Clock, FileText } from "lucide-react";
import { useRef, useState } from "react";
import { z } from "zod";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import { ToeicConfigPanel } from "#/features/exams/components/toeic-config-panel";
import { ToeicPartsBreakdown } from "#/features/exams/components/toeic-parts-breakdown";
import { toeicParts } from "#/features/exams/parts";
import {
	examKeys,
	examTestQueryOptions,
	useStartExamAttempt,
} from "#/features/exams/queries";
import { ApiError } from "#/lib/api/api";
import { loadOr404 } from "#/lib/router/route-loader";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/exams/toeic/$categoryId/$testId/")({
	loader: async ({
		context: { queryClient },
		params: { categoryId, testId },
	}) => {
		if (
			!z.uuid().safeParse(categoryId).success ||
			!z.uuid().safeParse(testId).success
		)
			throw notFound();
		const test = await loadOr404(queryClient, examTestQueryOptions(testId));
		if (test.categoryId !== categoryId || test.category?.type !== "toeic")
			throw notFound();
		return test;
	},
	pendingComponent: TestPending,
	component: ToeicTestDetailPage,
});

function TestPending() {
	return (
		<PageLayout>
			<Skeleton className="h-9 w-3/4" />
			<Skeleton className="h-4 w-1/2" />
			<div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
				<CardShell
					pending
					className="p-5 lg:col-span-7"
					skeleton={<Skeleton className="h-96 w-full rounded-xl" />}
				/>
				<CardShell
					pending
					className="p-5 lg:col-span-5"
					skeleton={<Skeleton className="h-[36rem] w-full rounded-xl" />}
				/>
			</div>
		</PageLayout>
	);
}

function ToeicTestDetailPage() {
	const test = Route.useLoaderData();
	return <TestConfiguration key={test.id} test={test} />;
}

function TestConfiguration({ test }: { test: ExamTestDetail }) {
	const navigate = useNavigate();
	const queryClient = useQueryClient();
	const start = useStartExamAttempt();
	const startLock = useRef(false);
	const [selectedMode, setSelectedMode] = useState<ExamMode>("exam");
	const parts = toeicParts(test.content);
	const allParts = parts.map((part) => part.partNumber);
	const [practiceParts, setPracticeParts] = useState<number[]>(allParts);
	const selectedParts = selectedMode === "exam" ? allParts : practiceParts;
	const [startError, setStartError] = useState<
		"failed" | "uncertain" | "navigation" | null
	>(null);
	const attemptId = start.data?.id;
	const busy = start.isPending || !!attemptId;
	const canStart =
		selectedParts.length > 0 &&
		(selectedMode !== "exam" || allParts.length === 7);
	useBreadcrumbs([
		{ label: m["exams.toeic.hub.title"](), to: APP_ROUTES.EXAMS.TOEIC.LIST },
		{
			label: test.category?.title ?? "",
			to: APP_ROUTES.EXAMS.TOEIC.CATEGORY,
			params: { categoryId: test.categoryId },
		},
		{ label: test.title },
	]);

	async function openSession(id: string) {
		try {
			await navigate({ to: APP_ROUTES.EXAMS.TOEIC.SESSION, params: { id } });
		} catch {
			setStartError("navigation");
		}
	}

	async function handleStart() {
		if (startLock.current || !canStart || busy || startError === "uncertain")
			return;
		startLock.current = true;
		setStartError(null);
		try {
			const attempt = await start.mutateAsync({
				testId: test.id,
				req: {
					mode: selectedMode,
					selectedParts: [...selectedParts].sort((a, b) => a - b),
				},
			});
			queryClient.setQueryData(examKeys.attempt(attempt.id), attempt);
			void queryClient.invalidateQueries({
				queryKey: [...examKeys.all, "attempts"],
			});
			await openSession(attempt.id);
		} catch (error) {
			setStartError(
				error instanceof ApiError && error.status < 500
					? "failed"
					: "uncertain",
			);
			startLock.current = false;
		}
	}

	return (
		<PageLayout
			action={
				<BackButton
					to={APP_ROUTES.EXAMS.TOEIC.CATEGORY}
					params={{ categoryId: test.categoryId }}
					label={m["exams.toeic.test.back"]()}
				/>
			}
		>
			<div className="flex flex-col gap-2">
				<h1 className="text-2xl font-bold tracking-tight text-foreground md:text-3xl">
					{test.title}
				</h1>
				<div className="flex flex-wrap items-center gap-4 text-xs text-muted-foreground md:text-sm">
					<span className="inline-flex items-center gap-1.5">
						<FileText className="size-4" aria-hidden="true" />
						{m["exams.toeic.category.questionsCount"]({
							count: parts.reduce((sum, part) => sum + part.questionsCount, 0),
						})}
					</span>
					<span className="inline-flex items-center gap-1.5">
						<Clock className="size-4" aria-hidden="true" />
						{selectedMode === "exam"
							? m["exams.toeic.category.durationMinutes"]({
									count: ExamDurationMinutes,
								})
							: m["exams.toeic.category.unlimitedDuration"]()}
					</span>
				</div>
			</div>
			<div className="grid grid-cols-1 items-start gap-6 lg:grid-cols-12">
				<div className="lg:col-span-7">
					<ToeicPartsBreakdown parts={parts} />
				</div>
				<div className="flex flex-col gap-4 lg:col-span-5">
					<ToeicConfigPanel
						parts={parts}
						selectedMode={selectedMode}
						selectedParts={selectedParts}
						pending={start.isPending}
						disabled={!!attemptId || startError === "uncertain" || !canStart}
						locked={busy}
						onSelectMode={setSelectedMode}
						onTogglePart={(partNumber) =>
							setPracticeParts((current) =>
								current.includes(partNumber)
									? current.filter((part) => part !== partNumber)
									: [...current, partNumber].sort((a, b) => a - b),
							)
						}
						onSelectAll={() => setPracticeParts(allParts)}
						onDeselectAll={() => setPracticeParts([])}
						onStartTest={() => {
							void handleStart();
						}}
					/>
					{startError && (
						<CardShell className="flex flex-col gap-3 p-4">
							<p role="alert" className="text-sm text-destructive">
								{startError === "uncertain"
									? m["exams.toeic.test.start.uncertain"]()
									: startError === "navigation"
										? m["exams.toeic.test.start.navigation"]()
										: m["exams.toeic.test.start.failed"]()}
							</p>
							{attemptId ? (
								<Button
									onClick={() => {
										void openSession(attemptId);
									}}
								>
									{m["exams.toeic.test.start.continue"]()}
								</Button>
							) : (
								<Link
									to={APP_ROUTES.EXAMS.TOEIC.HISTORY}
									className="text-sm font-semibold underline"
								>
									{m["exams.toeic.test.start.history"]()}
								</Link>
							)}
						</CardShell>
					)}
				</div>
			</div>
		</PageLayout>
	);
}
