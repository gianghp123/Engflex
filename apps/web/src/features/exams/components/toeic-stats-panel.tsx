import type { ExamStatistics } from "@engflex/contracts";
import {
	Activity,
	ArrowRight,
	Compass,
	TrendingUp,
	Trophy,
	Zap,
} from "lucide-react";
import { StatCard } from "#/components/common/stat-card";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
export function ToeicStatsPanel({
	stats,
	pending = false,
}: {
	stats?: ExamStatistics;
	pending?: boolean;
}) {
	const formatScore = (score: number | null | undefined) =>
		score == null
			? "\u2014"
			: new Intl.NumberFormat(getLocale(), { maximumFractionDigits: 1 }).format(
					score,
				);
	const metrics = [
		{
			id: "completed",
			title: m["exams.toeic.hub.analytics.completedTests"](),
			value: stats?.completedTests ?? "\u2014",
			icon: Compass,
			tone: "border-sky-100 bg-sky-50/70 dark:border-sky-900/40 dark:bg-sky-950/20",
			iconTone: "text-sky-600",
		},
		{
			id: "highest",
			title: m["exams.toeic.hub.analytics.highestScore"](),
			value: formatScore(stats?.highestScore),
			icon: Trophy,
			tone: "border-amber-100 bg-amber-50/70 dark:border-amber-900/40 dark:bg-amber-950/20",
			iconTone: "text-amber-500",
		},
		{
			id: "average",
			title: m["exams.toeic.hub.analytics.averageScore"](),
			value: formatScore(stats?.averageScore),
			icon: Activity,
			tone: "border-purple-100 bg-purple-50/70 dark:border-purple-900/40 dark:bg-purple-950/20",
			iconTone: "text-purple-600",
		},
		{
			id: "latest",
			title: m["exams.toeic.hub.analytics.latestScore"](),
			value: formatScore(stats?.latestScore),
			icon: Zap,
			tone: "border-emerald-100 bg-emerald-50/70 dark:border-emerald-900/40 dark:bg-emerald-950/20",
			iconTone: "text-emerald-600",
		},
	];
	return (
		<div className="flex flex-col gap-3">
			<div className="flex items-center justify-between">
				<div className="flex items-center gap-2 text-base font-bold text-foreground">
					<TrendingUp className="size-5 text-emerald-600" aria-hidden="true" />
					<span>{m["exams.toeic.hub.analytics.title"]()}</span>
				</div>
				<button
					type="button"
					className="inline-flex items-center gap-1 text-xs font-semibold text-muted-foreground transition-colors hover:text-foreground"
				>
					<span>{m["exams.toeic.hub.analytics.viewAll"]()}</span>
					<ArrowRight className="size-3.5" aria-hidden="true" />
				</button>
			</div>
			<div className="grid grid-cols-2 gap-3 md:gap-4">
				{metrics.map((metric) => (
					<StatCard
						key={metric.id}
						title={metric.title}
						value={metric.value}
						pending={pending}
						className={`rounded-xl p-3.5 md:p-4 ${metric.tone}`}
						aside={
							<metric.icon
								className={`size-5 ${metric.iconTone}`}
								aria-hidden="true"
							/>
						}
					/>
				))}
			</div>
		</div>
	);
}
