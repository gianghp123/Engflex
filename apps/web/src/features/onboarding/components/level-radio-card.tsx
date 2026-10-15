import { cn } from "cn";
import type { LucideIcon } from "lucide-react";
import { RadioGroupItem } from "#/components/ui/radio-group";

export function LevelRadioCard({
	value,
	range,
	title,
	description,
	hint,
	hintIcon: HintIcon,
	selected,
}: {
	value: string;
	/** Display range from the mock, e.g. "B1–B2" (not a single CEFR value). */
	range: string;
	title: string;
	description: string;
	hint: string;
	hintIcon: LucideIcon;
	selected: boolean;
}) {
	const id = `level-${value}`;
	return (
		<div className={cn(selected && "energy-card")}>
			<div
				className={cn(
					"flex items-start gap-3 rounded-[14px] border p-4 transition",
					selected
						? "energy-card-inner border-transparent"
						: "border-border bg-card hover:border-primary/40",
				)}
			>
				<RadioGroupItem id={id} value={value} className="mt-0.5" />
				<label htmlFor={id} className="flex min-w-0 flex-col gap-1">
					<span className="flex items-center gap-2">
						<span className="text-sm font-semibold">{title}</span>
						<span className="rounded-md border border-border bg-muted px-1.5 py-0.5 text-[11px] font-semibold text-muted-foreground">
							{range}
						</span>
					</span>
					<span className="text-xs text-muted-foreground">{description}</span>
					<span className="mt-1 inline-flex items-center gap-1 text-[11px] font-medium text-muted-foreground">
						<HintIcon className="size-3" />
						{hint}
					</span>
				</label>
			</div>
		</div>
	);
}
