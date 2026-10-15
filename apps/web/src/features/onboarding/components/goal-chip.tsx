import type { Goal } from "@engflex/contracts";
import type { LucideIcon } from "lucide-react";
import { ToggleGroupItem } from "#/components/ui/toggle-group";

export function GoalChip({
	value,
	label,
	icon: Icon,
}: {
	value: Goal;
	label: string;
	icon: LucideIcon;
}) {
	return (
		<ToggleGroupItem
			value={value}
			className="h-auto flex-col gap-2 rounded-[14px] border border-border px-4 py-3 text-sm font-semibold data-[state=on]:border-primary data-[state=on]:bg-secondary data-[state=on]:text-secondary-foreground"
		>
			<Icon className="size-4" />
			{label}
		</ToggleGroupItem>
	);
}
