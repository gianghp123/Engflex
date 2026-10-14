import { cn } from "cn";
import type { LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
import { CardShell } from "#/components/common/card-shell";
import { Skeleton } from "#/components/ui/skeleton";

const TONE_STROKE = {
	primary: "stroke-primary",
	accuracy: "stroke-accuracy",
	violet: "stroke-accent-violet",
} as const;

export function ProgressRing({
	value,
	tone = "primary",
}: {
	value: number;
	tone?: keyof typeof TONE_STROKE;
}) {
	const radius = 26;
	const circumference = 2 * Math.PI * radius;
	const pct = Math.min(100, Math.max(0, value));
	return (
		<svg viewBox="0 0 64 64" className="size-16 -rotate-90" aria-hidden="true">
			<circle
				cx="32"
				cy="32"
				r={radius}
				strokeWidth="6"
				className="fill-none stroke-muted"
			/>
			<circle
				cx="32"
				cy="32"
				r={radius}
				strokeWidth="6"
				strokeLinecap="round"
				className={cn("fill-none", TONE_STROKE[tone])}
				strokeDasharray={circumference}
				strokeDashoffset={circumference * (1 - pct / 100)}
			/>
		</svg>
	);
}

export function StatCard({
	label,
	icon: Icon,
	value,
	footer,
	tone = "primary",
	ringValue,
	className,
	pending = false,
}: {
	label: string;
	icon?: LucideIcon;
	value: ReactNode;
	footer?: ReactNode;
	tone?: keyof typeof TONE_STROKE;
	ringValue?: number;
	className?: string;
	pending?: boolean;
}) {
	const hasRing = typeof ringValue === "number";
	const skeleton = (
		<>
			<div className="flex min-w-0 flex-col gap-1">
				<Skeleton className="h-3 w-20" />
				<Skeleton className="h-7 w-12" />
				{hasRing ? <Skeleton className="h-3 w-16" /> : null}
			</div>
			{hasRing ? (
				<Skeleton className="size-16 rounded-full" />
			) : (
				<Skeleton className="size-9 rounded-lg" />
			)}
		</>
	);
	return (
		<CardShell
			pending={pending}
			skeleton={skeleton}
			className={cn("flex items-center justify-between gap-3 p-4", className)}
		>
			<div className="flex min-w-0 flex-col gap-1">
				<span className="text-[11px] font-semibold text-muted-foreground">
					{label}
				</span>
				<span className="text-2xl font-bold text-foreground">{value}</span>
				{footer ? (
					<span className="text-xs text-muted-foreground">{footer}</span>
				) : null}
			</div>
			{hasRing ? (
				<ProgressRing value={ringValue} tone={tone} />
			) : Icon ? (
				<span
					className={cn(
						"tile",
						tone === "accuracy" && "bg-accuracy-tint text-accuracy",
						tone === "violet" && "bg-secondary text-accent-violet",
						tone === "primary" && "bg-secondary text-primary",
					)}
				>
					<Icon className="size-5" />
				</span>
			) : null}
		</CardShell>
	);
}
