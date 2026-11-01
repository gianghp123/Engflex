import type { LessonStatus } from "@engflex/contracts";
import { cn } from "cn";
import { Check, Lock, LockOpen, RotateCw } from "lucide-react";
import { m } from "#/paraglide/messages";

export function LessonStatusPill({
	status,
	percent = 0,
	locked = false,
	className,
}: {
	status: LessonStatus;
	percent?: number;
	locked?: boolean;
	className?: string;
}) {
	if (locked) {
		return (
			<span className={cn("chip px-2 py-0.5 text-xs", className)}>
				<Lock className="size-3.5" />
				{m["lessons.status.locked"]()}
			</span>
		);
	}
	if (status === "completed") {
		return (
			<span
				className={cn(
					"chip border-cefr-b1-border bg-cefr-b1-bg px-2 py-0.5 text-xs text-cefr-b1",
					className,
				)}
			>
				<Check className="size-3.5" />
				{m["lessons.status.completed"]()}
			</span>
		);
	}
	if (status === "in_progress") {
		return (
			<span className={cn("chip chip-accent px-2 py-0.5 text-xs", className)}>
				<RotateCw className="size-3.5" />
				{m["lessons.status.inProgress"]()} ({percent}%)
			</span>
		);
	}
	return (
		<span
			className={cn(
				"chip border-border bg-muted px-2 py-0.5 text-xs text-muted-foreground",
				className,
			)}
		>
			<LockOpen className="size-3.5" />
			{m["lessons.status.available"]()}
		</span>
	);
}
