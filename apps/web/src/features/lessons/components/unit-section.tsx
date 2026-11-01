import type { Lesson } from "@engflex/contracts";
import { m } from "#/paraglide/messages";
import { LessonCard } from "./lesson-card";

export function UnitSection({
	section,
	units,
	unlocksAfterById,
	recommendedId,
}: {
	section: { id: string; title: string; cefrBand: string };
	units: Lesson[];
	unlocksAfterById: Record<string, string | undefined>;
	recommendedId?: string;
}) {
	return (
		<section className="flex flex-col gap-4">
			<div className="flex items-start justify-between gap-3">
				<div className="flex min-w-0 items-start gap-3">
					<span
						className="inline-flex h-8 shrink-0 items-center justify-center rounded-[12px] bg-primary px-2 text-sm font-bold text-primary-foreground"
						style={{ boxShadow: "0 2px 0 #433095" }}
					>
						{section.cefrBand}
					</span>
					<div className="min-w-0">
						<h2 className="text-lg font-bold text-foreground">
							{section.title}
						</h2>
					</div>
				</div>
				<span className="shrink-0 text-xs font-bold text-muted-foreground">
					{m["lessons.unit.lessons"]({ count: units.length })}
				</span>
			</div>
			<div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
				{units.map((lesson) => (
					<LessonCard
						key={lesson.id}
						lesson={lesson}
						locked={unlocksAfterById[lesson.id] !== undefined}
						unlocksAfter={unlocksAfterById[lesson.id]}
						recommended={recommendedId === lesson.id}
					/>
				))}
			</div>
		</section>
	);
}
