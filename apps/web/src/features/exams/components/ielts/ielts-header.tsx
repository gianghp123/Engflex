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
	pending = false,
}: {
	testTitle: string;
	timeLeft: number | null;
	highlightMode: boolean;
	onToggleHighlight: () => void;
	onReset: () => void;
	onExit: () => void;
	onSubmit: () => void;
	pending?: boolean;
}) {
	return (
		<header className="flex min-h-14 w-full shrink-0 flex-wrap items-center justify-between gap-2 border-b border-border/80 bg-background px-3 py-2 md:h-14 md:flex-nowrap md:px-6 md:py-0">
			<div className="flex w-full min-w-0 items-center gap-2.5 md:w-auto md:flex-1">
				<div className="flex size-7 shrink-0 items-center justify-center rounded-lg bg-red-600 text-white font-bold text-xs shadow-xs">
					🇬🇧
				</div>
				<h1 className="truncate text-sm font-bold text-foreground md:text-base">
					{testTitle}
				</h1>
			</div>

			<div className="flex shrink-0 items-center gap-1.5 whitespace-nowrap rounded-full border border-sky-200 bg-sky-50/80 px-3.5 py-1 text-xs font-mono font-bold text-sky-700 shadow-xs dark:border-sky-800 dark:bg-sky-950/40 dark:text-sky-300 md:text-sm">
				<Clock className="size-4" aria-hidden="true" />
				<span>
					{timeLeft === null ? (
						m["exams.ielts.take.practice"]()
					) : (
						<>
							<span className="sr-only sm:not-sr-only">
								{m["exams.ielts.take.remaining"]()}{" "}
							</span>
							{formatSeconds(timeLeft)}
						</>
					)}
				</span>
			</div>

			<div className="ml-auto flex shrink-0 items-center gap-2">
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
					disabled={pending}
					className="h-8 text-xs font-semibold gap-1"
					title={m["exams.ielts.take.resetButton"]()}
				>
					<RotateCcw className="size-3.5" />
					<span className="sr-only sm:not-sr-only">
						{m["exams.ielts.take.resetButton"]()}
					</span>
				</Button>

				<Button
					type="button"
					onClick={onExit}
					disabled={pending}
					className="flex h-8 items-center justify-center rounded-lg bg-rose-600 px-3 text-xs font-bold text-white shadow-xs transition-colors hover:bg-rose-700"
				>
					{m["exams.ielts.take.exitButton"]()}
				</Button>

				<SubmitButton
					type="button"
					onClick={onSubmit}
					pending={pending}
					className="flex h-8 items-center justify-center rounded-lg bg-emerald-600 px-3.5 text-xs font-bold text-white shadow-xs transition-colors hover:bg-emerald-700"
				>
					{m["exams.ielts.take.submitButton"]()}
				</SubmitButton>
			</div>
		</header>
	);
}
