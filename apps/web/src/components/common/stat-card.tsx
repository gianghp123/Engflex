import { cn } from "cn";
import type { ReactNode } from "react";
import { CardShell } from "#/components/common/card-shell";
import { Skeleton } from "#/components/ui/skeleton";

const TONE_STROKE = {
	primary: "stroke-primary",
	accuracy: "stroke-accuracy",
	violet: "stroke-accent-violet",
} as const;

/** A standalone widget, not a card prop: a caller passes it as `aside`. */
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

/**
 * The house stat card. Three things are fixed so they cannot drift between
 * pages — the title's typography, the metric's scale (`size`), and the
 * `CardShell` pending contract — and everything else belongs to the caller:
 * `aside` is the right column (ring, icon tile, badge…), `children` is the
 * body under the metric (bar, caption, CTA…). The frame never switches on what
 * those widgets are, so it does not grow a prop per caller.
 */
export function StatCard({
	title,
	value,
	size = "md",
	aside,
	children,
	skeleton,
	pending = false,
	className,
}: {
	title: ReactNode;
	value: ReactNode;
	/** `md` = `text-2xl`; `lg` = `stat-display` (the vocabulary hero number). */
	size?: "md" | "lg";
	/** Right column: the frame owns its placement, the caller owns its content. */
	aside?: ReactNode;
	/** Body under the metric. A child with `mt-auto` pins itself to the card foot. */
	children?: ReactNode;
	/**
	 * Overrides the placeholder when `aside`/`children` are an unusual shape, so
	 * the loading frame still matches the loaded one (`CardShell` rule).
	 */
	skeleton?: ReactNode;
	pending?: boolean;
	className?: string;
}) {
	const placeholder = (
		<>
			<div className="flex min-w-0 flex-1 flex-col gap-2 self-stretch">
				<Skeleton className="h-3 w-20" />
				<Skeleton className={size === "lg" ? "h-9 w-24" : "h-7 w-16"} />
				{children ? <Skeleton className="h-3 w-24" /> : null}
			</div>
			{aside ? <Skeleton className="size-9 rounded-lg" /> : null}
		</>
	);

	return (
		<CardShell
			pending={pending}
			skeleton={skeleton ?? placeholder}
			className={cn("flex items-center justify-between gap-3 p-5", className)}
		>
			{/* `self-stretch` keeps the column full height inside the centred row,
			    so a `mt-auto` child still lands at the card foot. */}
			<div className="flex min-w-0 flex-1 flex-col gap-2 self-stretch">
				<span className="text-[11px] font-bold text-muted-foreground">
					{title}
				</span>
				<span
					className={
						size === "lg"
							? "stat-display"
							: "text-2xl font-bold text-foreground"
					}
				>
					{value}
				</span>
				{children}
			</div>
			{aside ? <div className="flex shrink-0 items-center">{aside}</div> : null}
		</CardShell>
	);
}
