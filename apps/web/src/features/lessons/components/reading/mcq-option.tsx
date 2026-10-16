import { cn } from "cn";
import { Check } from "lucide-react";
import { RadioGroupItem } from "#/components/ui/radio-group";

export function McqOption({
	id,
	value,
	label,
	text,
	selected,
	disabled = false,
	tone = "default",
}: {
	id: string;
	value: string;
	label: string;
	text: string;
	selected: boolean;
	disabled?: boolean;
	tone?: "default" | "correct" | "wrong";
}) {
	return (
		<div
			className={cn(
				"flex items-center gap-3 rounded-[16px] border-2 p-3 transition focus-within:border-primary",
				tone === "correct" && "border-accuracy bg-accuracy-tint",
				tone === "wrong" && "border-ai-coral bg-card",
				tone === "default" &&
					(selected
						? "border-primary bg-accent"
						: "border-border bg-card hover:border-primary"),
				selected && tone === "default" && "shadow-edge-sm",
				disabled ? "cursor-default" : "cursor-pointer",
			)}
		>
			<RadioGroupItem
				id={id}
				value={value}
				disabled={disabled}
				className="sr-only"
			/>
			<span
				aria-hidden="true"
				className={cn(
					"inline-flex size-9 shrink-0 items-center justify-center rounded-full text-sm font-bold",
					tone === "correct"
						? "bg-accuracy text-primary-foreground"
						: tone === "wrong"
							? "bg-ai-coral text-primary-foreground"
							: selected
								? "bg-primary text-primary-foreground"
								: "bg-muted text-muted-foreground",
				)}
			>
				{label}
			</span>
			<label
				htmlFor={id}
				className={cn(
					"min-w-0 flex-1 text-[15px] font-medium text-foreground",
					disabled ? "cursor-default" : "cursor-pointer",
				)}
			>
				{text}
			</label>
			{selected && tone === "default" ? (
				<Check aria-hidden="true" className="size-4 shrink-0 text-primary" />
			) : null}
		</div>
	);
}
