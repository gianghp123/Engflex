import { ToggleGroupItem } from "#/components/ui/toggle-group";

export function TopicChip({ value, label }: { value: string; label: string }) {
	return (
		<ToggleGroupItem
			value={value}
			className="h-auto rounded-full border border-border px-3 py-1.5 text-xs font-medium data-[state=on]:border-primary data-[state=on]:bg-secondary data-[state=on]:text-secondary-foreground"
		>
			{label}
		</ToggleGroupItem>
	);
}
