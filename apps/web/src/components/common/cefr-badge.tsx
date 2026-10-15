import type { CEFR, ScenarioDifficulty } from "@engflex/contracts";
import { cn } from "cn";

/** Token-mapped CEFR chip + level dots. Total over CEFR | ScenarioDifficulty
 *  so a new member is a compile error; A1/A2 use the neutral tokens and C2
 *  reuses the C1 tone (no dedicated tokens exist for them). */
const CEFR_STYLES: Record<CEFR | ScenarioDifficulty, string> = {
	A1: "border-border bg-muted text-muted-foreground",
	A2: "border-border bg-muted text-muted-foreground",
	B1: "border-cefr-b1-border bg-cefr-b1-bg text-cefr-b1",
	"B1+": "border-cefr-b1p-border bg-cefr-b1p-bg text-cefr-b1p",
	B2: "border-cefr-b2-border bg-cefr-b2-bg text-cefr-b2",
	C1: "border-cefr-c1-border bg-cefr-c1-bg text-cefr-c1",
	C2: "border-cefr-c1-border bg-cefr-c1-bg text-cefr-c1",
};

const CEFR_DOTS: Record<CEFR | ScenarioDifficulty, number> = {
	A1: 0,
	A2: 0,
	B1: 0,
	"B1+": 1,
	B2: 2,
	C1: 3,
	C2: 3,
};

export function CefrBadge({
	value,
	className,
}: {
	value: CEFR | ScenarioDifficulty;
	className?: string;
}) {
	const filled = CEFR_DOTS[value] ?? 0;
	return (
		<span
			className={cn(
				"chip shrink-0 px-2 py-0.5 text-xs",
				CEFR_STYLES[value] ?? "border-border bg-muted text-muted-foreground",
				className,
			)}
		>
			<span className="level-dots" aria-hidden="true">
				{[0, 1, 2].map((i) => (
					<span
						key={i}
						className={cn(
							"size-[9px] rounded-full",
							i < filled ? "bg-primary" : "bg-border",
						)}
					/>
				))}
			</span>
			{value}
		</span>
	);
}
