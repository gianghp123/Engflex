import { createFileRoute, notFound, redirect } from "@tanstack/react-router";
import { z } from "zod";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { CardShell } from "#/components/common/card-shell";
import { PageLayout } from "#/components/common/page-layout";
import { Pager } from "#/components/common/pager";
import {
	Empty,
	EmptyDescription,
	EmptyHeader,
	EmptyTitle,
} from "#/components/ui/empty";
import { Skeleton } from "#/components/ui/skeleton";
import { ToeicCategoryHero } from "#/features/exams/components/toeic-category-hero";
import { ToeicTestCard } from "#/features/exams/components/toeic-test-card";
import {
	examCategoriesQueryOptions,
	examTestsQueryOptions,
} from "#/features/exams/queries";
import { loadOr404 } from "#/lib/router/route-loader";
import { parsePage } from "#/lib/router/search-params";
import { m } from "#/paraglide/messages";

const PAGE_SIZE = 10;

export const Route = createFileRoute("/_app/exams/toeic/$categoryId/")({
	validateSearch: (search: Record<string, unknown>): { page?: number } => ({
		page: parsePage(search.page),
	}),
	loaderDeps: ({ search }) => ({ page: search.page ?? 1 }),
	loader: async ({
		context: { queryClient },
		params: { categoryId },
		deps: { page },
	}) => {
		if (!z.uuid().safeParse(categoryId).success) throw notFound();
		const [categories, tests] = await Promise.all([
			loadOr404(queryClient, examCategoriesQueryOptions({ type: "toeic" })),
			loadOr404(
				queryClient,
				examTestsQueryOptions(categoryId, { page, pageSize: PAGE_SIZE }),
			),
		]);
		const category = categories.find((item) => item.id === categoryId);
		if (!category) throw notFound();
		const lastPage = Math.max(1, tests.pagination.totalPages);
		if (page > lastPage) {
			throw redirect({
				to: APP_ROUTES.EXAMS.TOEIC.CATEGORY,
				params: { categoryId },
				search: { page: lastPage },
				replace: true,
			});
		}
		return { category, tests };
	},
	pendingComponent: CategoryPending,
	component: ToeicCategoryDetailPage,
});

function CategoryPending() {
	return (
		<PageLayout>
			<CardShell
				pending
				skeleton={<Skeleton className="h-36 w-full rounded-xl" />}
				className="p-6"
			/>
			<div className="grid grid-cols-1 gap-3.5 sm:grid-cols-2">
				{[0, 1, 2, 3].map((card) => (
					<CardShell
						key={card}
						pending
						skeleton={
							<>
								<Skeleton className="h-5 w-3/4" />
								<Skeleton className="mt-4 h-4 w-1/2" />
								<Skeleton className="mt-4 ml-auto h-8 w-28" />
							</>
						}
						className="p-5"
					/>
				))}
			</div>
		</PageLayout>
	);
}

function ToeicCategoryDetailPage() {
	const { category, tests } = Route.useLoaderData();
	const navigate = Route.useNavigate();
	const { pagination, items } = tests;
	useBreadcrumbs([
		{ label: m["exams.toeic.hub.title"](), to: APP_ROUTES.EXAMS.TOEIC.LIST },
		{ label: category.title },
	]);
	return (
		<PageLayout
			action={
				<BackButton
					to={APP_ROUTES.EXAMS.TOEIC.LIST}
					label={m["exams.toeic.category.back"]()}
				/>
			}
		>
			<ToeicCategoryHero
				title={category.title}
				tags={[category.type.toUpperCase()]}
				testsCount={category.testsCount}
				attemptsCount={category.attemptsCount}
			/>
			<div className="flex flex-col gap-4">
				<h2 className="text-xs font-bold tracking-wider text-muted-foreground uppercase md:text-sm">
					{m["exams.toeic.category.testsListTitle"]()}
				</h2>
				{items.length === 0 ? (
					<Empty className="surface-card">
						<EmptyHeader>
							<EmptyTitle>{m["exams.toeic.category.empty.title"]()}</EmptyTitle>
							<EmptyDescription>
								{m["exams.toeic.category.empty.body"]()}
							</EmptyDescription>
						</EmptyHeader>
					</Empty>
				) : (
					<>
						<div className="grid grid-cols-1 gap-3.5 sm:grid-cols-2">
							{items.map((test) => (
								<ToeicTestCard
									key={test.id}
									item={{
										id: test.id,
										categoryId: test.categoryId,
										title: test.title,
										questionsCount: (test.content?.parts ?? []).reduce(
											(total, part) => total + part.questionCount,
											0,
										),
										attemptsCount: test.attemptsCount,
										durationMinutes: test.durationMinutes,
									}}
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
									? m["exams.toeic.category.pagination.range"]({
											from: (pagination.page - 1) * PAGE_SIZE + 1,
											to: Math.min(
												pagination.page * PAGE_SIZE,
												pagination.total,
											),
											total: pagination.total,
										})
									: m["exams.toeic.category.pagination.all"]({
											count: pagination.total,
										})
							}
						/>
					</>
				)}
			</div>
		</PageLayout>
	);
}
