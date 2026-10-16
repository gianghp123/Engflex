import { createFileRoute, notFound } from "@tanstack/react-router";
import { Gauge } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { BackButton } from "#/components/common/back-button";
import { PageLayout } from "#/components/common/page-layout";
import { ActivityRow } from "#/features/lessons/components/activity-row";
import { LessonHeader } from "#/features/lessons/components/lesson-header";
import { TipBar } from "#/features/lessons/components/tip-bar";
import { useStartPractice } from "#/features/lessons/hooks/use-start-practice";
import {
	lessonDetailBySlugQueryOptions,
	useLessonDetailBySlug,
	useLessonProgress,
} from "#/features/lessons/queries";
import {
	getFirstIncompleteActivity,
	partNumbersFor,
} from "#/features/lessons/store";
import { loadOr404 } from "#/lib/route-loader";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/lessons/$slug/")({
	// Preload here (not in the component): loader notFound() renders the
	// route's notFoundComponent; the same throw during render escapes down
	// the error-boundary path. Non-empty activities are enforced here too,
	// so the component never throws for missing data.
	loader: async ({ context: { queryClient }, params: { slug } }) => {
		const detail = await loadOr404(
			queryClient,
			lessonDetailBySlugQueryOptions(slug),
		);
		if (detail.activities.length === 0) throw notFound();
	},
	component: LessonDetailPage,
});

function LessonDetailPage() {
	const { slug } = Route.useParams();
	const detailQuery = useLessonDetailBySlug(slug);
	const detail = detailQuery.data;
	const starting = useStartPractice();
	const progressQuery = useLessonProgress();
	useBreadcrumbs(
		detail
			? [
					{ label: m["nav.item.lessons"](), to: APP_ROUTES.LESSONS.LIST },
					{ label: detail.title },
				]
			: [],
	);

	if (detailQuery.isPending) return null;
	// A failed fetch is an error, not a missing lesson (see loader above:
	// genuine 404s and empty lessons never reach the component).
	if (detailQuery.isError) throw detailQuery.error;
	if (!detail) return null;

	const serverItem = progressQuery.data?.find(
		(item) => item.lessonId === detail.id,
	);

	const startActivity = getFirstIncompleteActivity(
		detail.id,
		detail.activities,
		serverItem
			? partNumbersFor(detail.activities, serverItem.completedActivityIds)
			: [],
	);
	// Unreachable: the loader enforces non-empty activities, and the lookup
	// falls back to the first activity. Kept for type narrowing only.
	if (!startActivity) return null;

	return (
		<PageLayout
			action={
				<BackButton
					to={APP_ROUTES.LESSONS.LIST}
					label={m["lessons.detail.backToLesson"]()}
				/>
			}
		>
			<div className="flex flex-wrap items-center gap-2">
				<span className="chip px-2 py-0.5 text-xs">
					{detail.section?.title}
				</span>
				<span className="chip ml-auto px-2 py-0.5 text-xs">
					<Gauge className="size-3.5" />
					{m["lessons.detail.pace"]()}
				</span>
			</div>

			<LessonHeader
				lesson={detail}
				startPart={startActivity.type}
				startPartNumber={startActivity.partNumber}
				onStart={() => void starting.begin(detail.slug, detail.id)}
				startPending={starting.pending}
			/>

			<section className="flex flex-col gap-3">
				<div className="flex flex-wrap items-center gap-2">
					<h2 className="text-lg font-bold text-foreground">
						{m["lessons.detail.activitiesTitle"]()}
					</h2>
					<span className="text-xs font-bold text-muted-foreground">
						{m["lessons.detail.activitiesSubtitle"]()}
					</span>
				</div>
				<div className="flex flex-col gap-3">
					{detail.activities.map((activity) => (
						<ActivityRow
							key={activity.id}
							activity={activity}
							onStart={() => void starting.begin(detail.slug, detail.id)}
							startPending={starting.pending}
						/>
					))}
				</div>
			</section>

			<TipBar tip={m["lessons.detail.calmQuote"]()} />
		</PageLayout>
	);
}
