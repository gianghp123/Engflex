import type { Scenario } from "@engflex/contracts";
import { Link } from "@tanstack/react-router";
import { APP_ROUTES } from "#/app/app-route";
import { MoMascot } from "#/components/common/mo-mascot";
import {
	HistoryRow,
	HistoryRowSkeleton,
} from "#/features/voice/components/history/history-row";
import { useScenarioRuns } from "#/features/voice/queries";
import { m } from "#/paraglide/messages";

const RUNS_PREVIEW = 3;

/** Target vocabulary plus the scenario's previous runs (history API). */
export function ScenarioDetailSide({ scenario }: { scenario: Scenario }) {
	const runs = useScenarioRuns(scenario.id, RUNS_PREVIEW);
	const rows = runs.data?.items ?? [];
	const total = runs.data?.pagination.total ?? 0;

	return (
		<div className="flex flex-col gap-6">
			{scenario.details.vocab && scenario.details.vocab.length > 0 ? (
				<section className="surface-card flex flex-col gap-3 p-5">
					<div className="flex items-center justify-between gap-2">
						<h2 className="text-base font-bold text-foreground">
							{m["voice.detail.vocab.title"]()}
						</h2>
						<span className="text-[11px] font-medium text-muted-foreground">
							{m["voice.detail.vocab.count"]({
								count: scenario.details.vocab.length,
							})}
						</span>
					</div>
					<div className="flex flex-wrap gap-2">
						{scenario.details.vocab.map((term) => (
							<span key={term} className="chip">
								{term}
							</span>
						))}
					</div>
				</section>
			) : null}

			<section className="surface-card flex flex-col gap-3 p-5">
				<h2 className="text-base font-bold text-foreground">
					{m["voice.detail.runs.title"]()}
				</h2>
				{runs.isPending ? (
					<div className="flex flex-col divide-y">
						<HistoryRowSkeleton />
						<HistoryRowSkeleton />
					</div>
				) : runs.isError ? (
					<p className="text-xs text-muted-foreground">
						{m["voice.history.error"]()}
					</p>
				) : rows.length === 0 ? (
					<div className="flex flex-col items-center gap-2 text-center">
						<MoMascot variant="rest" size={56} />
						<p className="text-sm font-bold text-foreground">
							{m["voice.detail.runs.emptyTitle"]()}
						</p>
						<p className="text-xs text-muted-foreground">
							{m["voice.detail.runs.emptyBody"]()}
						</p>
					</div>
				) : (
					<>
						<div className="flex flex-col divide-y">
							{rows.map((conversation) => (
								<HistoryRow key={conversation.id} conversation={conversation} />
							))}
						</div>
						{total > rows.length ? (
							<Link
								to={APP_ROUTES.VOICE.HISTORY.LIST}
								search={{ scenarioId: scenario.id }}
								className="text-sm font-bold text-primary hover:underline"
							>
								{m["voice.history.viewAll"]()}
							</Link>
						) : null}
					</>
				)}
			</section>
		</div>
	);
}
