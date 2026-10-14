import type { Scenario } from "@engflex/contracts";
import { MoMascot } from "#/components/common/mo-mascot";
import { m } from "#/paraglide/messages";

/** Target vocabulary plus the runs shell (empty in v1 — no history yet). */
export function ScenarioDetailSide({ scenario }: { scenario: Scenario }) {
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

			<section className="surface-card flex flex-col items-center gap-2 p-5 text-center">
				<MoMascot variant="rest" size={56} />
				<h2 className="text-base font-bold text-foreground">
					{m["voice.detail.runs.title"]()}
				</h2>
				<p className="text-sm font-bold text-foreground">
					{m["voice.detail.runs.emptyTitle"]()}
				</p>
				<p className="text-xs text-muted-foreground">
					{m["voice.detail.runs.emptyBody"]()}
				</p>
			</section>
		</div>
	);
}
