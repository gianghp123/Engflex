import { cn } from "cn";
import { ChevronLeft, ChevronRight, Play, RotateCw } from "lucide-react";
import { Button } from "#/components/ui/button";
import { ToggleGroup, ToggleGroupItem } from "#/components/ui/toggle-group";
import {
	PlayerTools,
	ToolbarButton,
} from "#/features/videos/components/player-tools";
import {
	DICTATION_LEVELS,
	type DictationLevel,
} from "#/features/videos/dictation";
import { m } from "#/paraglide/messages";
import { DictationWordInput, type WordSlotState } from "./dictation-word-input";
import { WordHints } from "./word-hints";

const LEVEL_LABEL: Record<DictationLevel, () => string> = {
	easy: () => m["videos.dictation.level.easy"](),
	normal: () => m["videos.dictation.level.normal"](),
	hard: () => m["videos.dictation.level.hard"](),
};

/** Middle pane: level, sentence player, answer box, word hints, next. */
export function DictationPanel({
	level,
	onLevelChange,
	speed,
	words,
	revealed,
	wordStates,
	activeWordIndex,
	currentInput,
	isCompleted,
	onInputChange,
	onWordSubmit,
	onWordBackspace,
	onEnterPress,
	onResetSentence,
	onReveal,
	onRevealAll,
	onNext,
	onPrev,
	onReplaySentence,
	onPlaySentence,
	onNextSentence,
	onSpeed,
	onSettings,
	onShortcuts,
}: {
	level: DictationLevel;
	onLevelChange: (level: DictationLevel) => void;
	speed: number;
	words: string[];
	revealed: number[];
	wordStates: WordSlotState[];
	activeWordIndex: number;
	currentInput: string;
	isCompleted: boolean;
	onInputChange: (val: string) => void;
	onWordSubmit: (word: string) => void;
	onWordBackspace: () => void;
	onEnterPress?: () => void;
	onResetSentence?: () => void;
	onReveal?: (index: number) => void;
	onRevealAll?: () => void;
	onNext?: () => void;
	onPrev?: () => void;
	onReplaySentence?: () => void;
	onPlaySentence?: () => void;
	onNextSentence?: () => void;
	onSpeed?: () => void;
	onSettings?: () => void;
	onShortcuts?: () => void;
}) {
	return (
		<section className="surface-card flex flex-col gap-3 p-3">
			<ToggleGroup
				type="single"
				value={level}
				onValueChange={(value) => {
					if (value) onLevelChange(value as DictationLevel);
				}}
				aria-label={m["videos.dictation.level.label"]()}
				className="mx-auto rounded-full border-2 border-border bg-muted p-1"
			>
				{DICTATION_LEVELS.map((item) => (
					<ToggleGroupItem
						key={item}
						value={item}
						className={cn(
							"rounded-full! px-4 text-sm font-bold",
							"data-[state=on]:bg-sidebar data-[state=on]:text-sidebar-foreground",
						)}
					>
						{LEVEL_LABEL[item]()}
					</ToggleGroupItem>
				))}
			</ToggleGroup>

			<div className="flex items-center gap-1 rounded-2xl border-2 border-border bg-accent px-2 py-1.5 shadow-[0_3px_0_var(--border)]">
				<ToolbarButton
					icon={ChevronLeft}
					label={m["videos.practice.player.prev"]()}
					onClick={onPrev}
				/>
				<ToolbarButton
					icon={RotateCw}
					label={m["videos.practice.player.replay"]()}
					onClick={onReplaySentence}
				/>
				<ToolbarButton
					icon={Play}
					label={m["videos.practice.player.play"]()}
					onClick={onPlaySentence}
				/>
				<ToolbarButton
					icon={ChevronRight}
					label={m["videos.practice.player.next"]()}
					onClick={onNextSentence}
				/>
				<span className="ml-auto" />
				<PlayerTools
					speed={speed}
					onSpeed={onSpeed}
					onSettings={onSettings}
					onShortcuts={onShortcuts}
				/>
			</div>

			<DictationWordInput
				words={words}
				level={level}
				wordStates={wordStates}
				activeWordIndex={activeWordIndex}
				currentInput={currentInput}
				isCompleted={isCompleted}
				onInputChange={onInputChange}
				onWordSubmit={onWordSubmit}
				onWordBackspace={onWordBackspace}
				onEnterPress={onEnterPress}
				onResetSentence={onResetSentence}
			/>

			<WordHints words={words} revealed={revealed} onReveal={onReveal} />
			<p className="text-xs text-muted-foreground">
				{m["videos.dictation.hints.note"]()}
			</p>

			<Button
				type="button"
				onClick={onRevealAll}
				className="btn btn-accent w-full uppercase"
			>
				{m["videos.dictation.hints.revealAll"]()}
			</Button>
			<Button
				type="button"
				onClick={onNext}
				className="btn btn-primary w-full uppercase"
			>
				{m["videos.dictation.next"]()}
				<ChevronRight data-icon="inline-end" />
			</Button>
		</section>
	);
}
