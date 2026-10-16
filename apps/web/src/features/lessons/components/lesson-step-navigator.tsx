import type { Activity, ActivityType } from "@engflex/contracts";
import { useStore } from "@tanstack/react-store";
import { cn } from "cn";
import { CircleCheck } from "lucide-react";
import { PART_META } from "#/features/lessons/parts";
import { lessonsStore } from "#/features/lessons/store";
import { m } from "#/paraglide/messages";

export function LessonStepNavigator({
	lessonId,
	activities,
	currentPart,
	onSelect,
}: {
	lessonId: string;
	activities: Activity[];
	currentPart: ActivityType;
	onSelect: (part: ActivityType) => void;
}) {
	const completed = useStore(
		lessonsStore,
		(state) => state.completedParts[lessonId] ?? [],
	);

	return (
		<nav className="flex flex-wrap items-center gap-2">
			{activities.map((activity) => {
				const part = activity.type;
				const done = completed.includes(activity.partNumber);
				const active = part === currentPart;
				return (
					<button
						key={activity.id}
						type="button"
						onClick={() => onSelect(part)}
						aria-current={active ? "step" : undefined}
						className={cn(
							"inline-flex cursor-pointer items-center gap-1.5 rounded-full border-2 px-3 py-1 text-xs font-bold transition",
							active
								? "border-primary bg-primary text-primary-foreground shadow-edge-primary-md"
								: "border-border bg-card text-muted-foreground hover:border-primary",
						)}
					>
						{done && !active ? (
							<CircleCheck className="size-3.5 text-accuracy" />
						) : null}
						<span>
							{activity.partNumber}. {PART_META[part].label()}
						</span>
						<span className="sr-only">
							{done
								? m["lessons.step.done"]()
								: active
									? m["lessons.step.active"]()
									: m["lessons.step.upcoming"]()}
						</span>
					</button>
				);
			})}
		</nav>
	);
}
