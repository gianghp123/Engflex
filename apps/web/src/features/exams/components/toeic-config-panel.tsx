import { EXAM_MODE_VALUES, type ExamMode } from "@engflex/contracts";
import { SubmitButton } from "#/components/common/submit-button";
import { enumOptions } from "#/lib/enum-options";
import { m } from "#/paraglide/messages";
import type { ToeicPartItem } from "./toeic-parts-breakdown";

export interface ToeicConfigPanelProps {
	durationMinutes: number;
	parts: ToeicPartItem[];
	selectedMode: ExamMode;
	selectedParts: number[];
	pending?: boolean;
	disabled?: boolean;
	locked?: boolean;
	onSelectMode?: (mode: ExamMode) => void;
	onTogglePart?: (partNumber: number) => void;
	onSelectAll?: () => void;
	onDeselectAll?: () => void;
	onStartTest?: () => void;
}

const MODE_OPTIONS = enumOptions<ExamMode>(EXAM_MODE_VALUES, {
	exam: () => m["exams.toeic.test.config.examModeTitle"](),
	practice: () => m["exams.toeic.test.config.practiceModeTitle"](),
});
const MODE_DESCRIPTIONS: Record<ExamMode, () => string> = {
	exam: () => m["exams.toeic.test.config.examModeDesc"](),
	practice: () => m["exams.toeic.test.config.practiceModeDesc"](),
};

export function ToeicConfigPanel({
	durationMinutes,
	parts,
	selectedMode,
	selectedParts,
	pending = false,
	disabled = false,
	locked = false,
	onSelectMode,
	onTogglePart,
	onSelectAll,
	onDeselectAll,
	onStartTest,
}: ToeicConfigPanelProps) {
	const questionCount = parts
		.filter((part) => selectedParts.includes(part.partNumber))
		.reduce((sum, part) => sum + part.questionsCount, 0);
	return (
		<fieldset
			disabled={locked}
			className="flex min-w-0 flex-col gap-6 rounded-2xl border border-border/80 bg-card p-5 shadow-sm md:p-6"
		>
			<div className="flex flex-col gap-2.5">
				<h3 className="text-xs font-bold tracking-wider text-muted-foreground uppercase">
					{m["exams.toeic.test.config.modeTitle"]()}
				</h3>
				<div className="grid grid-cols-1 gap-2.5 sm:grid-cols-2">
					{MODE_OPTIONS.map((mode) => (
						<button
							key={mode.value}
							type="button"
							aria-pressed={selectedMode === mode.value}
							onClick={() => onSelectMode?.(mode.value)}
							className={`flex flex-col items-start rounded-xl p-3.5 text-left transition-all ${selectedMode === mode.value ? "border-2 border-foreground/90 bg-muted/20 shadow-sm" : "border border-border/80 hover:bg-muted/30"}`}
						>
							<span className="text-sm font-bold text-foreground">
								{mode.label()}
							</span>
							<span className="mt-1 text-xs text-muted-foreground">
								{MODE_DESCRIPTIONS[mode.value]()}
							</span>
						</button>
					))}
				</div>
			</div>
			<fieldset
				disabled={selectedMode === "exam"}
				className="flex min-w-0 flex-col gap-3"
			>
				<legend className="sr-only">
					{m["exams.toeic.test.config.partsSelectionTitle"]()}
				</legend>
				<div className="flex items-center justify-between">
					<h3 className="text-sm font-bold text-foreground">
						{m["exams.toeic.test.config.partsSelectionTitle"]()}
					</h3>
					<div className="flex items-center gap-2 text-xs font-semibold text-muted-foreground">
						<button
							type="button"
							onClick={onSelectAll}
							className="transition-colors hover:text-foreground disabled:opacity-50"
						>
							{m["exams.toeic.test.config.selectAll"]()}
						</button>
						<span>•</span>
						<button
							type="button"
							onClick={onDeselectAll}
							className="transition-colors hover:text-foreground disabled:opacity-50"
						>
							{m["exams.toeic.test.config.deselectAll"]()}
						</button>
					</div>
				</div>
				{(["listening", "reading"] as const).map((section) => (
					<div key={section} className="flex flex-col gap-2">
						<span className="text-[11px] font-bold tracking-wider text-muted-foreground uppercase">
							{section === "listening"
								? m["exams.toeic.test.breakdown.listening"]()
								: m["exams.toeic.test.breakdown.reading"]()}
						</span>
						<div className="flex flex-col gap-2">
							{parts
								.filter((part) => part.section === section)
								.map((part) => (
									<label
										key={part.partNumber}
										className="flex cursor-pointer items-center gap-2.5 text-xs text-foreground select-none"
									>
										<input
											type="checkbox"
											checked={selectedParts.includes(part.partNumber)}
											onChange={() => onTogglePart?.(part.partNumber)}
											className="size-4 rounded border-border text-primary accent-primary"
										/>
										<span>
											{m["exams.toeic.test.config.partOption"]({
												part: part.partNumber,
												type: part.type,
												count: part.questionsCount,
											})}
										</span>
									</label>
								))}
						</div>
					</div>
				))}
			</fieldset>
			<div className="rounded-xl border border-border/60 bg-muted/40 p-3 text-center text-xs font-semibold text-foreground md:text-sm">
				{selectedMode === "exam"
					? m["exams.toeic.test.config.summary"]({
							parts: selectedParts.length,
							questions: questionCount,
							duration: durationMinutes,
						})
					: m["exams.toeic.test.config.practiceSummary"]({
							parts: selectedParts.length,
							questions: questionCount,
						})}
			</div>
			{selectedMode === "exam" && (
				<p className="whitespace-pre-line text-center text-xs text-muted-foreground">
					{m["exams.toeic.test.config.autoSubmitNote"]()}
				</p>
			)}
			<SubmitButton
				type="button"
				pending={pending}
				disabled={disabled}
				onClick={onStartTest}
				className="w-full rounded-xl bg-emerald-600 py-3.5 text-sm font-bold text-white shadow-sm transition-colors hover:bg-emerald-700 md:text-base cursor-pointer"
			>
				{m["exams.toeic.test.config.startButton"]()}
			</SubmitButton>
		</fieldset>
	);
}
