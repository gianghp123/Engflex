import { cn } from "cn";
import { ChevronLeft, ChevronRight, Mic, Play, RotateCw } from "lucide-react";
import { Button } from "#/components/ui/button";
import { Textarea } from "#/components/ui/textarea";
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
	onReveal,
	onRevealAll,
	answer,
	onAnswerChange,
	onNext,
	onPrev,
	onReplaySentence,
	onPlaySentence,
	onNextSentence,
	onSpeed,
	onSettings,
	onShortcuts,
	onMic,
}: {
	level: DictationLevel;
	onLevelChange: (level: DictationLevel) => void;
	speed: number;
	words: string[];
	revealed: number[];
	onReveal?: (index: number) => void;
	onRevealAll?: () => void;
	answer: string;
	onAnswerChange: (value: string) => void;
	onNext?: () => void;
	onPrev?: () => void;
	onReplaySentence?: () => void;
	onPlaySentence?: () => void;
	onNextSentence?: () => void;
	onSpeed?: () => void;
	onSettings?: () => void;
	onShortcuts?: () => void;
	onMic?: () => void;
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

			<div className="relative">
				<label
					htmlFor="dictation-answer"
					className="absolute top-3 left-4 text-xs font-extrabold tracking-wide text-muted-foreground uppercase"
				>
					{m["videos.dictation.answer.label"]()}
				</label>
				<Textarea
					id="dictation-answer"
					value={answer}
					onChange={(event) => onAnswerChange(event.target.value)}
					placeholder={m["videos.dictation.answer.placeholder"]()}
					rows={5}
					className="field min-h-40 pt-9 pb-14 text-[15px]"
				/>
				<Button
					type="button"
					variant="outline"
					size="icon"
					onClick={onMic}
					aria-label={m["videos.dictation.answer.mic"]()}
					title={m["videos.dictation.answer.mic"]()}
					className="absolute right-3 bottom-3 rounded-full shadow-[0_3px_0_var(--border)]"
				>
					<Mic />
				</Button>
			</div>

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
