import { cn } from "cn";
import { CheckCircle2, RotateCcw } from "lucide-react";
import { useEffect, useRef } from "react";
import { Input } from "#/components/ui/input";
import { type DictationLevel, getWordHint } from "#/features/videos/dictation";
import { m } from "#/paraglide/messages";

export interface WordSlotState {
	status: "pending" | "correct" | "incorrect";
	input: string;
}

export interface DictationWordInputProps {
	words: string[];
	level: DictationLevel;
	wordStates: WordSlotState[];
	activeWordIndex: number;
	currentInput: string;
	isCompleted: boolean;
	onInputChange: (val: string) => void;
	onWordSubmit: (word: string) => void;
	onWordBackspace: () => void;
	onEnterPress?: () => void;
	onResetSentence?: () => void;
}

export function DictationWordInput({
	words,
	level,
	wordStates,
	activeWordIndex,
	currentInput,
	isCompleted,
	onInputChange,
	onWordSubmit,
	onWordBackspace,
	onEnterPress,
	onResetSentence,
}: DictationWordInputProps) {
	const inputRef = useRef<HTMLInputElement>(null);

	useEffect(() => {
		if (!isCompleted) {
			inputRef.current?.focus();
		}
	}, [isCompleted]);

	const correctCount = wordStates.filter((s) => s.status === "correct").length;
	const totalWords = words.length;

	const wordItems = words.map((word, i) => ({
		keyId: `slot-${i}`,
		index: i,
		word,
	}));

	return (
		<div className="flex flex-col gap-3">
			<div className="surface-card min-h-36 flex flex-wrap items-center content-start gap-2 rounded-2xl border-2 border-border p-4 bg-muted/30">
				{wordItems.map(({ keyId, index, word }) => {
					const state = wordStates[index];
					const isActive = index === activeWordIndex;
					const isPast = index < activeWordIndex;
					const hint = getWordHint(word, level);

					let wordContent = word;
					let styleClasses =
						"border border-border/60 bg-card text-muted-foreground";

					if (isPast && state) {
						if (state.status === "correct") {
							styleClasses =
								"border-accuracy/40 bg-accuracy/15 text-accuracy font-bold";
							wordContent = word;
						} else {
							styleClasses =
								"border-ai-coral/40 bg-ai-coral/15 text-ai-coral font-bold";
							wordContent = `${state.input || "—"} (${word})`;
						}
					} else if (isActive) {
						styleClasses =
							"border-primary bg-primary/10 text-primary font-bold shadow-[0_2px_0_var(--primary)]";
						wordContent = currentInput.trim() ? currentInput : hint || "____";
					} else {
						wordContent = hint || "____";
					}

					return (
						<span
							key={keyId}
							className={cn(
								"chip min-h-8 min-w-10 px-3 py-1 font-mono text-sm tracking-wide transition-all",
								styleClasses,
							)}
						>
							{wordContent}
						</span>
					);
				})}
			</div>

			{isCompleted ? (
				<div className="flex flex-col gap-2 rounded-2xl border-2 border-accuracy/40 bg-accuracy/10 p-3">
					<div className="flex items-center justify-between gap-2">
						<span className="flex items-center gap-1.5 text-sm font-bold text-accuracy">
							<CheckCircle2 className="size-4" />
							{m["videos.dictation.input.accuracy"]({
								correct: correctCount,
								total: totalWords,
							})}
						</span>
						{onResetSentence ? (
							<button
								type="button"
								onClick={onResetSentence}
								className="flex items-center gap-1 text-xs font-semibold text-muted-foreground hover:text-foreground cursor-pointer"
							>
								<RotateCcw className="size-3.5" />
								{m["videos.dictation.video.replay"]()}
							</button>
						) : null}
					</div>
					<p className="text-xs font-medium text-foreground">
						{m["videos.dictation.input.sentenceDone"]()}
					</p>
				</div>
			) : (
				<div className="flex flex-col gap-1.5">
					<Input
						ref={inputRef}
						type="text"
						value={currentInput}
						onChange={(e) => onInputChange(e.target.value)}
						onKeyDown={(e) => {
							if (e.key === " ") {
								e.preventDefault();
								if (currentInput.trim()) {
									onWordSubmit(currentInput.trim());
								}
							} else if (e.key === "Backspace" && currentInput === "") {
								e.preventDefault();
								onWordBackspace();
							} else if (e.key === "Enter") {
								e.preventDefault();
								if (isCompleted) {
									onEnterPress?.();
								} else if (currentInput.trim()) {
									onWordSubmit(currentInput.trim());
								}
							}
						}}
						placeholder={m["videos.dictation.input.hint"]()}
						className="field h-12 text-base font-medium px-4"
						autoComplete="off"
						autoCapitalize="off"
						spellCheck={false}
					/>
				</div>
			)}
		</div>
	);
}
