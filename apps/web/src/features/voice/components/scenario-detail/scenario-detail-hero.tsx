import type { Scenario } from "@engflex/contracts";
import { ArrowRight, Clock3 } from "lucide-react";
import { CefrBadge } from "#/components/common/cefr-badge";
import { SubmitButton } from "#/components/common/submit-button";
import { personaInitials } from "#/features/voice/persona";
import { m } from "#/paraglide/messages";

/** Hero card: badges, title, objective, partner tile, start CTA. */
export function ScenarioDetailHero({
	scenario,
	onStart,
	pending = false,
}: {
	scenario: Scenario;
	onStart: () => void;
	pending?: boolean;
}) {
	const persona = scenario.persona;
	return (
		<section className="surface-card flex flex-col gap-5 p-5 md:p-6">
			<div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
				<div className="min-w-0 flex-1">
					<div className="flex flex-wrap items-center gap-2">
						{scenario.topic?.shortName ? (
							<span className="chip px-2.5 py-1 text-[11px] font-bold">
								{scenario.topic.shortName}
							</span>
						) : null}
						<span className="inline-flex items-center gap-1 text-[11px] font-medium text-muted-foreground">
							<Clock3 className="size-3.5" />
							{m["voice.detail.upTo"]({ max: scenario.maxDuration })}
						</span>
						<span className="ml-auto shrink-0">
							<CefrBadge value={scenario.cefrLevel} />
						</span>
					</div>
					<h1 className="mt-3 text-2xl font-bold tracking-tight text-foreground md:text-3xl">
						{scenario.title}
					</h1>
					<p className="mt-2 text-[15px] font-medium text-muted-foreground">
						{scenario.objective}
					</p>
					{persona ? (
						<div className="mt-4 inline-flex items-center gap-3">
							<span className="tile bg-secondary text-xs font-bold text-secondary-foreground">
								{personaInitials(persona.roleTitle)}
							</span>
							<span className="min-w-0">
								<span className="block truncate text-sm font-bold text-foreground">
									{persona.name} · {persona.roleTitle}
								</span>
								{persona.personality ? (
									<span className="block text-[11px] font-medium text-muted-foreground">
										{persona.personality}
									</span>
								) : null}
							</span>
						</div>
					) : null}
				</div>
				<SubmitButton
					type="button"
					className="btn btn-primary shrink-0"
					onClick={onStart}
					pending={pending}
				>
					{m["voice.detail.start"]()}
					<ArrowRight data-icon="inline-end" />
				</SubmitButton>
			</div>
		</section>
	);
}
