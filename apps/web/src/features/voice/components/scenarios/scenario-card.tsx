import type { Scenario } from "@engflex/contracts";
import { cn } from "cn";
import { Clock3 } from "lucide-react";
import { CefrBadge } from "#/components/common/cefr-badge";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";

const TILE_TONES = [
	"bg-primary text-primary-foreground",
	"bg-secondary text-secondary-foreground",
	"bg-accent-violet text-primary-foreground",
] as const;

function personaInitials(roleTitle: string): string {
	const words = roleTitle.split(/\s+/).filter(Boolean);
	const letters = words.map((word) =>
		word.replace(/^[^A-Za-z]*/, "").charAt(0),
	);
	return (letters.slice(0, 2).join("") || "?").toUpperCase();
}

export function ScenarioCard({
	scenario,
	onStart,
}: {
	scenario: Scenario;
	onStart: () => void;
}) {
	const tone =
		TILE_TONES[(scenario.persona?.name.length ?? 1) % TILE_TONES.length];

	return (
		<div className="surface-card flex h-full flex-col gap-3 p-5">
			<div className="flex items-center gap-2">
				{scenario.persona ? (
					<>
						<span className={cn("tile text-xs font-bold", tone)}>
							{personaInitials(scenario.persona.roleTitle)}
						</span>
						<span className="min-w-0">
							<span className="block truncate text-xs font-bold text-foreground">
								{scenario.persona.name} · {scenario.persona.roleTitle}
							</span>
							<span className="block text-[11px] font-medium text-muted-foreground">
								{scenario.persona.personality}
							</span>
						</span>
					</>
				) : null}
				<span className="ml-auto shrink-0">
					<CefrBadge value={scenario.cefrLevel} />
				</span>
			</div>
			<h3 className="text-base font-bold text-foreground">{scenario.title}</h3>
			<p className="text-[15px] font-medium text-muted-foreground">
				{scenario.objective}
			</p>
			<div className="mt-auto flex items-center justify-between gap-2 pt-1">
				<span className="inline-flex items-center gap-1 text-[11px] font-medium text-muted-foreground">
					<Clock3 className="size-3.5" />
					{m["voice.card.maxSession"]({ max: scenario.maxDuration })}
				</span>
				<Button
					type="button"
					size="sm"
					className="btn btn-primary"
					onClick={onStart}
				>
					{m["voice.card.start"]()}
				</Button>
			</div>
		</div>
	);
}
