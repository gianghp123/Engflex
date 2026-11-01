import type { ActivityType, Lesson } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { useStore } from "@tanstack/react-store";
import { cn } from "cn";
import { Clock3, Layers, Lock, RotateCcw } from "lucide-react";
import { APP_ROUTES } from "#/app/app-route";
import { CefrBadge } from "#/components/common/cefr-badge";
import { PremiumBadge } from "#/components/common/premium-badge";
import { ProgressBar } from "#/components/common/progress-bar";
import { Button } from "#/components/ui/button";
import { useLessonDetail } from "#/features/lessons/queries";
import {
	getFirstIncompleteActivity,
	lessonsStore,
	progressFrom,
} from "#/features/lessons/store";
import { useCanAccessPremium } from "#/features/payments/queries";
import { m } from "#/paraglide/messages";
import { LessonStatusPill } from "./status-pill";

const ACTIVITY_LABEL: Record<ActivityType, () => string> = {
	reading: () => m["lessons.skill.reading"](),
	listening: () => m["lessons.skill.listening"](),
	writing: () => m["lessons.skill.writing"](),
	speaking: () => m["lessons.skill.speaking"](),
};

export function LessonCard({
	lesson,
	locked = false,
	unlocksAfter,
	recommended = false,
}: {
	lesson: Lesson;
	locked?: boolean;
	unlocksAfter?: string;
	recommended?: boolean;
}) {
	const detailQuery = useLessonDetail(lesson.id);
	const detail = detailQuery.data;
	const activities = detail?.activities ?? [];
	// The card already fetches the detail for its activity chips, so the real
	// part count feeds the progress math here — no server-side count field.
	const storeState = useStore(lessonsStore);
	const progress = progressFrom(storeState, lesson.id, activities.length);
	const current = getFirstIncompleteActivity(lesson.id, activities)?.type;
	const canAccessPremium = useCanAccessPremium();
	const premiumBlocked = lesson.isPremium && !canAccessPremium;

	return (
		<article
			className={cn(
				"surface-card flex h-full flex-col gap-3 p-5",
				locked && "opacity-90",
			)}
		>
			<div className="flex items-center justify-between gap-2">
				<div className="flex items-center gap-2">
					<CefrBadge value={lesson.cefrLevel} />
					{lesson.isPremium ? <PremiumBadge /> : null}
				</div>
				<LessonStatusPill
					status={progress.status}
					percent={progress.percent}
					locked={locked}
				/>
			</div>
			{recommended && !locked ? (
				<span className="chip chip-accent w-fit px-2 py-0.5 text-xs">
					{m["lessons.status.recommended"]()}
				</span>
			) : null}
			<h3 className="text-base font-bold text-foreground">{lesson.title}</h3>
			<p className="text-[15px] leading-[22px] font-medium text-muted-foreground">
				{lesson.description}
			</p>
			<div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground">
				<span className="inline-flex items-center gap-1">
					<Clock3 className="size-3.5" />
					{m["lessons.card.meta"]({
						minutes: lesson.details.estimatedDurationMin ?? 0,
					})}
				</span>
				<span className="inline-flex items-center gap-1">
					<Layers className="size-3.5" />
					{lesson.section?.title}
				</span>
			</div>
			{activities.length > 0 ? (
				<div className="flex flex-wrap gap-1.5">
					{activities.map((activity) => (
						<span
							key={activity.id}
							className={cn(
								"chip px-2 py-0.5 text-[11px]",
								!locked &&
									activity.type === current &&
									progress.status === "in_progress" &&
									"chip-selected",
							)}
						>
							{ACTIVITY_LABEL[activity.type]()}
							{!locked &&
							activity.type === current &&
							progress.status === "in_progress"
								? ` (${m["lessons.status.nextUp"]()})`
								: null}
						</span>
					))}
				</div>
			) : null}
			{progress.percent > 0 && !locked ? (
				<ProgressBar value={progress.percent} />
			) : null}
			<div className="mt-auto flex items-center justify-between gap-2 pt-1">
				{locked ? (
					<span className="inline-flex items-center gap-1 text-xs font-bold text-muted-foreground">
						<Lock className="size-3.5" />
						{m["lessons.status.unlocks"]({ id: unlocksAfter ?? "" })}
					</span>
				) : progress.status === "completed" ? (
					<span className="text-xs font-bold text-foreground">
						{m["lessons.status.completed"]()}
					</span>
				) : progress.status === "in_progress" ? (
					<span className="text-xs font-bold text-primary">
						{m["lessons.card.stepOf"]({
							done: progress.partsCompleted,
							total: progress.partsTotal,
						})}
					</span>
				) : (
					<span className="text-xs font-bold text-muted-foreground">
						{m["lessons.status.unstarted"]()}
					</span>
				)}
				{locked ? (
					<Button type="button" variant="ghost" size="sm" disabled>
						{m["lessons.status.locked"]()}
					</Button>
				) : premiumBlocked ? (
					<Button asChild size="sm" className="btn btn-primary">
						<Link to={APP_ROUTES.PAYMENTS.PRICING}>
							{m["common.premium.upgradeAction"]()}
						</Link>
					</Button>
				) : progress.status === "completed" ? (
					<Button
						asChild
						variant="outline"
						size="sm"
						className="btn btn-outline"
					>
						<Link to={APP_ROUTES.LESSONS.DETAIL} params={{ slug: lesson.slug }}>
							<RotateCcw data-icon="inline-start" />
							{m["lessons.card.actionReview"]()}
						</Link>
					</Button>
				) : progress.status === "in_progress" ? (
					<Button asChild size="sm" className="btn btn-primary">
						<Link to={APP_ROUTES.LESSONS.DETAIL} params={{ slug: lesson.slug }}>
							{m["lessons.card.actionContinue"]()}
						</Link>
					</Button>
				) : (
					<Button
						asChild
						variant="outline"
						size="sm"
						className="btn btn-outline"
					>
						<Link to={APP_ROUTES.LESSONS.DETAIL} params={{ slug: lesson.slug }}>
							{m["lessons.hub.start"]()}
						</Link>
					</Button>
				)}
			</div>
		</article>
	);
}
