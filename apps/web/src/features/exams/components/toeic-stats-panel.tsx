import type { ExamAttemptStats } from "@engflex/contracts";
import { TrendingUp } from "lucide-react";
import { StatCard } from "#/components/common/stat-card";
import { m } from "#/paraglide/messages";

export function ToeicStatsPanel({
	stats,
	pending = false,
}: {
	stats?: ExamAttemptStats;
	pending?: boolean;
}) {
	const score = (value: number | null | undefined) =>
		value == null
			? m["exams.attempt.unavailable"]()
			: m["exams.toeic.hub.analytics.score"]({ value: Math.round(value) });
	return (
		<section className="flex flex-col gap-3">
			<h2 className="flex items-center gap-2 text-base font-bold">
				<TrendingUp className="size-5 text-emerald-600" aria-hidden="true" />
				{m["exams.toeic.hub.analytics.title"]()}
			</h2>
			<p className="text-xs text-muted-foreground">
				{m["exams.toeic.hub.analytics.scope"]()}
			</p>
			<div className="grid grid-cols-2 gap-3">
				<StatCard
					pending={pending}
					title={m["exams.toeic.hub.analytics.completedTests"]()}
					value={stats?.completedTests ?? 0}
				/>
				<StatCard
					pending={pending}
					title={m["exams.toeic.hub.analytics.highestScore"]()}
					value={score(stats?.highestScore)}
				/>
				<StatCard
					pending={pending}
					title={m["exams.toeic.hub.analytics.averageScore"]()}
					value={score(stats?.averageScore)}
				/>
				<StatCard
					pending={pending}
					title={m["exams.toeic.hub.analytics.latestScore"]()}
					value={score(stats?.latestScore)}
				/>
			</div>
		</section>
	);
}
