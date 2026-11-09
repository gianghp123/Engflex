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
import {
	IeltsTestCard,
	type IeltsTestCardItem,
} from "#/features/exams/components/ielts/ielts-test-card";
import {
	examCategoriesQueryOptions,
	examTestsQueryOptions,
} from "#/features/exams/queries";
import { loadOr404 } from "#/lib/router/route-loader";
import { parsePage } from "#/lib/router/search-params";
import { m } from "#/paraglide/messages";

const PAGE_SIZE = 10;

export const Route = createFileRoute("/_app/exams/ielts/$categoryId/")({
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
			loadOr404(queryClient, examCategoriesQueryOptions({ type: "ielts" })),
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
				to: APP_ROUTES.EXAMS.IELTS.CATEGORY,
				params: { categoryId },
				search: { page: lastPage },
				replace: true,
			});
		}
		return { category, tests };
	},
	pendingComponent: CategoryPending,
	component: IeltsCategoryDetailPage,
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

function IeltsCategoryDetailPage() {
	const { category, tests } = Route.useLoaderData();
	const navigate = Route.useNavigate();
	const { pagination, items } = tests;
	useBreadcrumbs([
		{ label: m["exams.ielts.hub.title"](), to: APP_ROUTES.EXAMS.IELTS.LIST },
		{ label: category.title },
	]);

	return (
		<PageLayout>
			<div className="flex flex-col gap-6">
				<div>
					<BackButton
						to={APP_ROUTES.EXAMS.IELTS.LIST}
						label={m["exams.ielts.category.back"]()}
					/>
					<div className="mt-3 flex flex-wrap items-center gap-3">
						<h1 className="text-2xl font-bold tracking-tight text-foreground md:text-3xl">
							{category.title}
						</h1>
						{category.subType && (
							<span className="rounded-md bg-primary/10 px-2.5 py-0.5 text-xs font-bold text-primary uppercase">
								{category.subType}
							</span>
						)}
					</div>
				</div>

				<div className="flex flex-col gap-4">
					<h2 className="text-xs font-bold uppercase tracking-wider text-muted-foreground">
						{m["exams.ielts.category.testsListTitle"]()}
					</h2>

					{items.length === 0 ? (
						<Empty>
							<EmptyHeader>
								<EmptyTitle>
									{m["exams.ielts.category.emptyTitle"]()}
								</EmptyTitle>
								<EmptyDescription>
									{m["exams.ielts.category.emptyDesc"]()}
								</EmptyDescription>
							</EmptyHeader>
						</Empty>
					) : (
						<div className="flex flex-col gap-3">
							{items.map((test) => {
								const cardItem: IeltsTestCardItem = {
									id: test.id,
									categoryId: category.id,
									title: test.title,
									skill: test.skill,
									questionsCount:
										test.ieltsContent?.sections.reduce(
											(sum, s) => sum + s.questionCount,
											0,
										) ?? 40,
									durationMinutes:
										test.skill === "listening"
											? 40
											: test.skill === "writing"
												? 60
												: 60,
								};
								return <IeltsTestCard key={test.id} item={cardItem} />;
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
			</div>
		</PageLayout>
	);
}
