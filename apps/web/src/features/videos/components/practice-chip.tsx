import { cn } from "cn";
import { CircleCheck, CircleX } from "lucide-react";
import { m } from "#/paraglide/messages";

export function PracticeChip({
	label,
	done,
}: {
	label: string;
	done?: boolean;
}) {
	const Icon = done ? CircleCheck : CircleX;
	return (
		<span
			className={cn(
				"inline-flex items-center gap-1.5 text-sm font-bold",
				done ? "text-primary" : "text-foreground",
			)}
		>
			<span aria-hidden={done !== undefined}>{label}</span>
			{done !== undefined ? (
				<>
					<Icon
						aria-hidden="true"
						className={cn(
							"size-4",
							done ? "text-primary" : "text-muted-foreground/70",
						)}
					/>
					<span className="sr-only">
						{done
							? m["videos.card.practice.done"]({ mode: label })
							: m["videos.card.practice.todo"]({ mode: label })}
					</span>
				</>
			) : null}
		</span>
	);
}
