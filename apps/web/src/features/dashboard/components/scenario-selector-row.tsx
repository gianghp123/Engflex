import { cn } from "cn";
import { Coins } from "lucide-react";

export function ScenarioSelectorRow({
	title,
	status,
	selected,
	onSelect,
	costText,
}: {
	title: string;
	status: string;
	selected: boolean;
	onSelect: () => void;
	costText?: string;
}) {
	return (
		<button
			type="button"
			aria-pressed={selected}
			onClick={onSelect}
			className={cn(
				"flex w-full items-center justify-between gap-3 rounded-[16px] border-2 px-3 py-2 text-left text-xs transition",
				selected
					? "border-primary bg-accent"
					: "border-border bg-card hover:border-primary",
			)}
		>
			<span className="flex min-w-0 flex-col">
				<span className="block truncate font-medium text-foreground">
					{title}
				</span>
				{costText ? (
					<span className="mt-0.5 flex items-center gap-1 text-[11px] font-medium text-muted-foreground">
						<Coins className="size-3.5" />
						{costText}
					</span>
				) : null}
			</span>
			<span
				className={cn(
					"shrink-0 text-[11px] font-semibold",
					selected ? "text-primary" : "text-muted-foreground",
				)}
			>
				{status}
			</span>
		</button>
	);
}
