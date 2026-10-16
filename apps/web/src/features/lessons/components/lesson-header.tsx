import type { ActivityType, LessonDetail } from "@engflex/contracts";
import { Clock3, Layers } from "lucide-react";
import { CefrBadge } from "#/components/common/cefr-badge";
import { MoMascot } from "#/components/common/mo-mascot";
import { SubmitButton } from "#/components/common/submit-button";
import { PART_META } from "#/features/lessons/parts";
import { m } from "#/paraglide/messages";

export function LessonHeader({
	lesson,
	startPart,
	startPartNumber,
	onStart,
	startPending,
}: {
	lesson: LessonDetail;
	startPart: ActivityType;
	startPartNumber: number;
	onStart: () => void;
	startPending: boolean;
}) {
	const partCount = lesson.activities.length;
	return (
		<section className="surface-hero flex flex-col gap-5 p-5 md:flex-row md:p-6">
			<div className="flex min-w-0 flex-1 flex-col gap-4">
				<div className="flex flex-wrap items-center gap-2">
					<CefrBadge value={lesson.cefrLevel} />
					<span className="chip px-2 py-0.5 text-xs">
						<Clock3 className="size-3.5" />
						{m["lessons.card.estDuration"]({
							minutes: lesson.details.estimatedDurationMin ?? 0,
						})}
					</span>
					<span className="chip px-2 py-0.5 text-xs">
						<Layers className="size-3.5" />
						{m["lessons.card.activitiesCount"]({ count: partCount })}
					</span>
				</div>
				<div>
					<h1 className="text-[28px] leading-[36px] font-bold text-foreground md:text-[38px] md:leading-[46px]">
						{lesson.title}
					</h1>
					<p className="mt-2 max-w-2xl text-[15px] leading-[22px] font-medium text-muted-foreground">
						{lesson.description}
					</p>
				</div>
				<div className="flex flex-wrap gap-3">
					<SubmitButton
						type="button"
						className="btn btn-primary"
						onClick={onStart}
						pending={startPending}
					>
						{m["lessons.detail.startLesson"]({
							number: startPartNumber,
							part: PART_META[startPart].label(),
						})}
					</SubmitButton>
				</div>
			</div>
			<div className="flex shrink-0 items-start justify-center">
				<div className="flex w-44 flex-col items-center gap-2 rounded-[24px] border-2 border-border bg-tip-bg p-4 shadow-edge">
					<MoMascot variant="cheer" size={112} />
					<span className="rounded-full bg-primary px-2.5 py-0.5 text-[11px] font-bold text-primary-foreground">
						{m["lessons.detail.coach"]()}
					</span>
					<span className="text-xs font-bold text-foreground">
						{m["lessons.detail.steady"]()}
					</span>
				</div>
			</div>
		</section>
	);
}
