import type { ConversationObjective } from "@engflex/contracts";
import { m } from "#/paraglide/messages";

/** Pre-session checklist: the scenario's ordered tasks, read off `objectives[].id` order. */
export function ObjectivesRow({
	objectives,
}: {
	objectives: ConversationObjective[];
}) {
	const ordered = [...objectives].sort((a, b) => a.id - b.id);
	if (ordered.length === 0) return null;
	return (
		<section className="surface-card flex flex-col gap-3 p-5">
			<h2 className="text-base font-bold text-foreground">
				{m["voice.room.objectives.title"]()}
			</h2>
			<ol className="flex list-decimal flex-col gap-2 pl-5 marker:font-bold marker:text-primary">
				{ordered.map((objective) => (
					<li
						key={objective.id}
						className="flex flex-col gap-0.5 text-[15px] font-medium text-foreground"
					>
						<span>{objective.label}</span>
						{objective.hint ? (
							<span className="text-xs font-normal text-muted-foreground">
								{m["voice.room.objectives.hint"]({ hint: objective.hint })}
							</span>
						) : null}
					</li>
				))}
			</ol>
		</section>
	);
}
