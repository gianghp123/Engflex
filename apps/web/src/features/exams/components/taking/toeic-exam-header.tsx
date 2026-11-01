import { Clock } from "lucide-react";
import { Logo } from "#/components/common/logo";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";

function formatSeconds(totalSec: number): string {
	const mm = Math.floor(totalSec / 60);
	const ss = totalSec % 60;
	return `${String(mm).padStart(2, "0")}:${String(ss).padStart(2, "0")}`;
}

export function ToeicExamHeader({
	testTitle,
	pending,
	practice,
	locked,
	timeLeft,
	speed,
	onCycleSpeed,
	onSubmitClick,
	onExitClick,
}: {
	testTitle: string;
	pending: boolean;
	practice: boolean;
	locked: boolean;
	timeLeft: number;
	speed: number;
	onCycleSpeed: () => void;
	onSubmitClick: () => void;
	onExitClick: () => void;
}) {
	return (
		<header className="flex h-14 w-full shrink-0 items-center justify-between border-b border-border/80 bg-background px-4 md:px-6">
			{/* Left: Test Branding */}
			<div className="flex items-center gap-2.5 min-w-0">
				<Logo variant="mark" className="shrink-0" />
				<h1 className="truncate text-sm font-bold text-foreground md:text-base">
					{testTitle}
				</h1>
			</div>

			{/* Center: Countdown Timer Badge */}
			<div className="flex items-center gap-1.5 rounded-full border border-sky-200 bg-sky-50/80 px-3.5 py-1 text-xs font-mono font-bold text-sky-700 shadow-xs dark:border-sky-800 dark:bg-sky-950/40 dark:text-sky-300 md:text-sm">
				<Clock className="size-4" aria-hidden="true" />
				<span>
					{practice
						? m["exams.toeic.take.elapsed"]()
						: m["exams.toeic.take.remaining"]()}{" "}
					{formatSeconds(timeLeft)}
				</span>
			</div>

			{/* Right: Actions */}
			<div className="flex items-center gap-2">
				<button
					type="button"
					onClick={onCycleSpeed}
					className="flex h-8 items-center justify-center rounded-full border border-border/80 bg-muted/40 px-3 text-xs font-bold text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
					title={m["exams.toeic.take.speed"]()}
				>
					{speed}x
				</button>

				<SubmitButton
					pending={pending}
					disabled={locked}
					type="button"
					onClick={onSubmitClick}
					className="flex h-8 items-center justify-center rounded-lg bg-emerald-600 px-3.5 text-xs font-bold text-white shadow-xs transition-colors hover:bg-emerald-700"
				>
					{m["exams.toeic.take.submitButton"]()}
				</SubmitButton>

				<Button
					disabled={locked}
					type="button"
					onClick={onExitClick}
					className="flex h-8 items-center justify-center rounded-lg bg-rose-600 px-3.5 text-xs font-bold text-white shadow-xs transition-colors hover:bg-rose-700"
				>
					{m["exams.toeic.take.exitButton"]()}
				</Button>
			</div>
		</header>
	);
}
