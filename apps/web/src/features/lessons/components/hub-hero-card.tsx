import type { Lesson } from "@engflex/contracts";
import { useQueries } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { MoMascot } from "#/components/common/mo-mascot";
import { ProgressBar } from "#/components/common/progress-bar";
import { Button } from "#/components/ui/button";
import { lessonDetailQueryOptions } from "#/features/lessons/queries";
import type { LessonProgress } from "#/features/lessons/store";
import { m } from "#/paraglide/messages";

export function HubHeroCard({
	section,
	lessons,
	progressById,
	resumeSlug,
	resumeTitle,
}: {
	section: string;
	lessons: Lesson[];
	progressById: Record<string, LessonProgress>;
	resumeSlug: string;
	resumeTitle: string;
}) {
	// Totals come from the same detail queries the cards below already run
	// (identical keys dedupe), so untouched lessons still count their parts.
	const detailQueries = useQueries({
		queries: lessons.map((lesson) => lessonDetailQueryOptions(lesson.id)),
	});
	const totalParts = detailQueries.reduce(
		(sum, query) => sum + (query.data?.activities.length ?? 0),
		0,
	);
	const doneParts = lessons.reduce(
		(sum, lesson) => sum + (progressById[lesson.id]?.partsCompleted ?? 0),
		0,
	);
	const percent =
		totalParts === 0 ? 0 : Math.round((doneParts / totalParts) * 100);

	return (
		<section className="surface-hero flex flex-col gap-5 p-5 md:flex-row">
			<div className="flex min-w-0 flex-1 flex-col gap-3">
				<div className="flex flex-wrap items-center gap-2">
					<span className="chip border-cefr-b1-border bg-cefr-b1-bg px-2 py-0.5 text-xs text-cefr-b1">
						{m["lessons.hub.inFlight"]()}
					</span>
					<span className="text-xs font-bold text-muted-foreground">
						{m["lessons.unit.label"]({ n: section })}
					</span>
				</div>
				<h2 className="text-xl font-bold text-foreground md:text-2xl">
					{resumeTitle}
				</h2>
				<div className="flex flex-col gap-1">
					<div className="flex items-center justify-between text-xs">
						<span className="font-bold text-foreground">
							{m["lessons.hub.overall"]()}
						</span>
						<span className="font-bold text-primary">{percent}%</span>
					</div>
					<ProgressBar value={percent} />
				</div>
				<div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
					{lessons.map((lesson) => {
						const progress = progressById[lesson.id];
						const dot =
							progress?.status === "completed"
								? "bg-primary"
								: progress?.status === "in_progress"
									? "bg-secondary"
									: "bg-border";
						return (
							<span
								key={lesson.id}
								className="inline-flex items-center gap-1.5"
							>
								<span className={`size-1.5 rounded-full ${dot}`} />
								{lesson.title}
							</span>
						);
					})}
				</div>
				<div>
					<Button asChild className="btn btn-primary w-fit">
						<Link to={APP_ROUTES.LESSONS.DETAIL} params={{ slug: resumeSlug }}>
							{m["lessons.hub.resume"]({ unit: resumeTitle })}
						</Link>
					</Button>
				</div>
			</div>
			<div className="flex shrink-0 items-center justify-center">
				<div className="relative hidden size-[112px] items-center justify-center rounded-full bg-accent shadow-edge-sm sm:flex">
					<MoMascot variant="wave" size={80} />
					<span className="absolute -top-1 -right-1 rounded-full bg-tip-bg px-2 py-0.5 text-[11px] font-bold text-tip-ink shadow-edge-xs">
						{m["lessons.hub.paced"]()}
					</span>
				</div>
			</div>
		</section>
	);
}
