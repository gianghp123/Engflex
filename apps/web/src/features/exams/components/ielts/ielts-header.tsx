import { Clock, Highlighter, RotateCcw } from "lucide-react";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { Switch } from "#/components/ui/switch";
import { m } from "#/paraglide/messages";

function formatSeconds(totalSec: number): string {
	const mm = Math.floor(totalSec / 60);
	const ss = totalSec % 60;
	return `${String(mm).padStart(2, "0")}:${String(ss).padStart(2, "0")}`;
}

export function IeltsHeader({
	testTitle,
	timeLeft,
	highlightMode,
	onToggleHighlight,
	onReset,
	onExit,
	onSubmit,
}: {
	testTitle: string;
	timeLeft: number;
	highlightMode: boolean;
	onToggleHighlight: () => void;
	onReset: () => void;
	onExit: () => void;
	onSubmit: () => void;
}) {
	return (
		<header className="flex h-14 w-full shrink-0 items-center justify-between border-b border-border/80 bg-background px-3 md:px-6">
			<div className="flex items-center gap-2.5 min-w-0">
				<div className="flex size-7 shrink-0 items-center justify-center rounded-lg bg-red-600 text-white font-bold text-xs shadow-xs">
					🇬🇧
				</div>
				<h1 className="truncate text-sm font-bold text-foreground md:text-base">
					{testTitle}
				</h1>
			</div>

			<div className="flex items-center gap-1.5 rounded-full border border-sky-200 bg-sky-50/80 px-3.5 py-1 text-xs font-mono font-bold text-sky-700 shadow-xs dark:border-sky-800 dark:bg-sky-950/40 dark:text-sky-300 md:text-sm">
				<Clock className="size-4" aria-hidden="true" />
				<span>
					{m["exams.ielts.take.remaining"]()} {formatSeconds(timeLeft)}
				</span>
			</div>

			<div className="flex items-center gap-2">
				<div className="hidden sm:flex items-center gap-2 mr-2 rounded-lg border border-border/70 bg-muted/30 px-2.5 py-1 text-xs">
					<Highlighter className="size-3.5 text-amber-500" />
					<label
						htmlFor="ielts-highlight-toggle"
						className="cursor-pointer font-medium text-foreground text-xs select-none"
					>
						{m["exams.ielts.take.highlightToggle"]()}
					</label>
					<Switch
						id="ielts-highlight-toggle"
						checked={highlightMode}
						onCheckedChange={onToggleHighlight}
					/>
				</div>

				<Button
					variant="outline"
					size="sm"
					onClick={onReset}
					className="h-8 text-xs font-semibold gap-1 hidden md:inline-flex"
					title={m["exams.ielts.take.resetButton"]()}
				>
					<RotateCcw className="size-3.5" />
					<span>{m["exams.ielts.take.resetButton"]()}</span>
				</Button>

				<Button
					type="button"
					onClick={onExit}
					className="flex h-8 items-center justify-center rounded-lg bg-rose-600 px-3 text-xs font-bold text-white shadow-xs transition-colors hover:bg-rose-700"
				>
					{m["exams.ielts.take.exitButton"]()}
				</Button>

				<SubmitButton
					type="button"
					onClick={onSubmit}
					className="flex h-8 items-center justify-center rounded-lg bg-emerald-600 px-3.5 text-xs font-bold text-white shadow-xs transition-colors hover:bg-emerald-700"
				>
					{m["exams.ielts.take.submitButton"]()}
				</SubmitButton>
			</div>
		</header>
	);
}
