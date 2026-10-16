import type { LessonProgressSummary } from "@engflex/contracts";
import { createFileRoute } from "@tanstack/react-router";
import { useStore } from "@tanstack/react-store";
import { useMemo, useState } from "react";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { EmptyList } from "#/components/common/empty-list";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { HubHeroCard } from "#/features/lessons/components/hub-hero-card";
import {
	DEFAULT_LESSON_FILTERS,
	LessonFilterBar,
	type LessonFilters,
} from "#/features/lessons/components/lesson-filter-bar";
import { UnitSection } from "#/features/lessons/components/unit-section";
import {
	lessonSectionsQueryOptions,
	useLessonProgress,
	useLessonSections,
} from "#/features/lessons/queries";
import { lessonsStore, progressFrom } from "#/features/lessons/store";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/lessons/")({
	// Preload so a failed fetch reaches errorComponent instead of rendering
	// the empty state below (which currently masquerades outages as "no lessons").
	loader: ({ context: { queryClient } }) =>
		queryClient.query(lessonSectionsQueryOptions()),
	component: LessonsPage,
});

function LessonsPage() {
	useBreadcrumbs([{ label: m["nav.item.lessons"]() }]);
	const [filters, setFilters] = useState<LessonFilters>(DEFAULT_LESSON_FILTERS);
	const storeState = useStore(lessonsStore);
	const sectionsQuery = useLessonSections();
	const sections = sectionsQuery.data ?? [];

	const units = useMemo(
		() => sections.flatMap((section) => section.units),
		[sections],
	);

	const progressQuery = useLessonProgress();
	const serverById = useMemo(() => {
		const byId: Record<string, LessonProgressSummary | undefined> = {};
		for (const item of progressQuery.data ?? []) byId[item.lessonId] = item;
		return byId;
	}, [progressQuery.data]);

	const progressById = useMemo(() => {
		return Object.fromEntries(
			units.map((lesson) => {
				const server = serverById[lesson.id];
				return [
					lesson.id,
					progressFrom(storeState, lesson.id, undefined, {
						count: server?.completedActivityIds.length ?? 0,
						completed: server?.status === "completed",
					}),
				];
			}),
		);
	}, [storeState, units, serverById]);

	const visible = useMemo(() => {
		const search = filters.search.trim().toLowerCase();
		return sections
			.map((section) => ({
				...section,
				units: section.units.filter((lesson) => {
					if (search) {
						const haystack =
							`${lesson.title} ${lesson.description}`.toLowerCase();
						if (!haystack.includes(search)) return false;
					}
					if (filters.level !== "all" && lesson.cefrLevel !== filters.level)
						return false;
					return true;
				}),
			}))
			.filter((section) => section.units.length > 0);
	}, [sections, filters]);

	const hero = useMemo(() => {
		for (const section of sections) {
			const resume = section.units.find(
				(lesson) => progressById[lesson.id]?.status === "in_progress",
			);
			if (resume) {
				const recommended = section.units.find(
					(lesson) => progressById[lesson.id]?.status === "unstarted",
				);
				return { section, resume, recommendedId: recommended?.id };
			}
		}
		const first = sections[0];
		return {
			section: first,
			resume: first?.units[0],
			recommendedId: first?.units[0]?.id,
		};
	}, [sections, progressById]);

	// A failed fetch is an error: without this the empty state below
	// masquerades every outage as "no lessons". Thrown after all hooks so
	// hook order stays stable across renders.
	if (sectionsQuery.isError) throw sectionsQuery.error;

	return (
		<PageLayout
			hero={{
				icon: <MoMascot variant="nice" size={84} />,
				title: m["lessons.hub.title"](),
				description: m["lessons.hub.moBody"](),
			}}
		>
			<LessonFilterBar filters={filters} onChange={setFilters} />
			{hero.section && hero.resume ? (
				<HubHeroCard
					section={hero.section.title}
					lessons={hero.section.units}
					progressById={progressById}
					resumeSlug={hero.resume.slug}
					resumeTitle={hero.resume.title}
				/>
			) : null}
			{sectionsQuery.isPending ? null : visible.length === 0 ? (
				<EmptyList
					title={m["lessons.hub.emptyTitle"]()}
					description={m["lessons.hub.emptyBody"]()}
					onReset={() => setFilters(DEFAULT_LESSON_FILTERS)}
				/>
			) : (
				visible.map((section) => (
					<UnitSection
						key={section.id}
						section={section}
						units={section.units}
						serverById={serverById}
						recommendedId={
							section.id === hero.section?.id ? hero.recommendedId : undefined
						}
					/>
				))
			)}
		</PageLayout>
	);
}
