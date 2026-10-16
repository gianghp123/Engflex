import { Link } from "@tanstack/react-router";
import { useStore } from "@tanstack/react-store";
import { APP_ROUTES } from "#/app/app-route";
import { CardShell } from "#/components/common/card-shell";
import { ProgressBar } from "#/components/common/progress-bar";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { Skeleton } from "#/components/ui/skeleton";
import { useLessonDetail, useLessonSections } from "#/features/lessons/queries";
import { lessonsStore, progressFrom } from "#/features/lessons/store";
import { m } from "#/paraglide/messages";

export function ActiveLessonCard() {
	const sectionsQuery = useLessonSections();
	const units = sectionsQuery.data?.flatMap((section) => section.units) ?? [];
	const storeState = useStore(lessonsStore);
	const lesson =
		units.find(
			(unit) => progressFrom(storeState, unit.id).status === "in_progress",
		) ?? units[0];
	// Only the picked lesson needs its activity total; the store already knows
	// the totals of any lesson practiced this session.
	const detailQuery = useLessonDetail(lesson?.id);
	const progress = lesson
		? progressFrom(storeState, lesson.id, detailQuery.data?.activities.length)
		: null;

	if (sectionsQuery.isPending) {
		return (
			<CardShell
				pending
				surface="surface-hero"
				className="flex h-full flex-col gap-4 p-5"
				skeleton={
					<>
						<Skeleton className="h-6 w-28 rounded-full" />
						<div className="flex flex-col gap-2">
							<Skeleton className="h-5 w-3/4" />
							<Skeleton className="h-4 w-full" />
						</div>
						<div className="mt-auto flex flex-col gap-2">
							<Skeleton className="h-3 w-40" />
							<Skeleton className="h-2 w-full rounded-full" />
						</div>
						<Skeleton className="h-9 w-36 rounded-lg" />
					</>
				}
			/>
		);
	}

	if (!lesson || !progress) return null;

	return (
		<CardShell
			surface="surface-hero"
			className="flex h-full flex-col gap-4 p-5"
		>
			<Badge variant="outline" className="chip w-fit">
				{m["dashboard.activeLesson.unitLabel"]()}
			</Badge>
			<div>
				<h3 className="text-lg font-bold text-foreground">{lesson.title}</h3>
				<p className="mt-1 text-[15px] font-medium text-muted-foreground">
					{lesson.description}
				</p>
			</div>
			<div className="mt-auto flex flex-col gap-2">
				<div className="flex items-center justify-between text-xs text-muted-foreground">
					<span>
						{m["dashboard.activeLesson.progress"]({
							done: progress.partsCompleted,
							total: progress.partsTotal,
						})}
					</span>
					<span className="font-semibold text-foreground">
						{progress.percent}%
					</span>
				</div>
				<ProgressBar value={progress.percent} />
			</div>
			<Button asChild className="btn btn-primary w-fit">
				<Link to={APP_ROUTES.LESSONS.DETAIL} params={{ slug: lesson.slug }}>
					{m["dashboard.activeLesson.continue"]()}
				</Link>
			</Button>
		</CardShell>
	);
}
