import { cn } from "cn";
import {
	Check,
	ChevronLeft,
	ChevronRight,
	Eye,
	Pause,
	Play,
	RotateCw,
	X,
} from "lucide-react";
import {
	type ComponentProps,
	type ReactNode,
	useId,
	useRef,
	useState,
} from "react";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
	AlertDialogTrigger,
} from "#/components/ui/alert-dialog";
import { Button } from "#/components/ui/button";
import { Input } from "#/components/ui/input";
import { Textarea } from "#/components/ui/textarea";
import { ToggleGroup, ToggleGroupItem } from "#/components/ui/toggle-group";
import {
	PlayerTools,
	ToolbarButton,
} from "#/features/videos/components/player-tools";
import {
	type checkDictation,
	type DictationLevel,
	dictationInputWords,
	maskWord,
} from "#/features/videos/dictation";
import { m } from "#/paraglide/messages";

const levelItemClass =
	"h-6 rounded-sm! px-2 text-xs font-semibold data-[state=on]:bg-primary data-[state=on]:text-primary-foreground disabled:data-[state=on]:opacity-100";

export function DictationPanel({
	slots,
	draft,
	answers = [],
	revealed,
	completed,
	level,
	levelLocked,
	showCorrect,
	isPlaying,
	onLevel,
	onAnswers,
	onRevealAll,
	onDraftChange,
	onReveal,
	speed,
	rates,
	onPrev,
	onReplaySentence,
	onPlaySentence,
	onNextSentence,
	onSpeed,
	onSettings,
	onShortcuts,
	children,
}: {
	slots: ReturnType<typeof checkDictation>["slots"];
	draft: string;
	answers?: string[];
	revealed: number[];
	completed: boolean;
	level: DictationLevel;
	levelLocked: boolean;
	showCorrect: boolean;
	isPlaying: boolean;
	onLevel: (level: DictationLevel) => void;
	onAnswers: (answers: string[]) => void;
	onRevealAll: () => void;
	onDraftChange: (draft: string) => void;
	onReveal: (positions: number[]) => void;
	speed: number;
	rates: number[];
	onPrev?: () => void;
	onReplaySentence?: () => void;
	onPlaySentence?: () => void;
	onNextSentence?: () => void;
	onSpeed: (speed: number) => void;
	onSettings: () => void;
	onShortcuts: () => void;
	children?: ReactNode;
}) {
	const inputId = useId();
	const levelHintId = useId();
	const [composition, setComposition] = useState<string | null>(null);
	const inputs = useRef<Array<HTMLInputElement | null>>([]);
	const blankCount = slots.filter(
		(slot) => slot.blankIndex !== undefined,
	).length;
	const updateAnswers = (index: number, values: string[]) => {
		const next = Array.from({ length: blankCount }, (_, i) => answers[i] ?? "");
		for (const [offset, value] of values.entries()) {
			if (index + offset < blankCount) next[index + offset] = value;
		}
		onAnswers(next);
	};
	const wordSlots = slots.map((slot, index) => ({
		...slot,
		position: index,
		id: `word-${index}`,
		revealed: revealed.includes(index),
	}));

	return (
		<section className="surface-card flex flex-col gap-2.5 p-3">
			<ToggleGroup
				type="single"
				value={level}
				disabled={levelLocked}
				aria-describedby={levelLocked ? levelHintId : undefined}
				onValueChange={(value) => {
					if (value === "easy" || value === "normal" || value === "hard")
						onLevel(value);
				}}
				aria-label={m["videos.dictation.level.label"]()}
				className="mx-auto rounded-md border border-border bg-muted p-0.5"
			>
				<ToggleGroupItem value="easy" className={levelItemClass}>
					{m["videos.dictation.level.easy"]()}
				</ToggleGroupItem>
				<ToggleGroupItem value="normal" className={levelItemClass}>
					{m["videos.dictation.level.normal"]()}
				</ToggleGroupItem>
				<ToggleGroupItem value="hard" className={levelItemClass}>
					{m["videos.dictation.level.hard"]()}
				</ToggleGroupItem>
			</ToggleGroup>
			{levelLocked ? (
				<p
					id={levelHintId}
					className="text-center text-xs text-muted-foreground"
				>
					{m["videos.dictation.level.locked"]()}
				</p>
			) : null}
			<div className="flex flex-wrap items-center rounded-xl border border-border bg-muted/40 p-1">
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
					icon={isPlaying ? Pause : Play}
					label={
						isPlaying
							? m["videos.dictation.playback.pause"]()
							: m["videos.practice.player.play"]()
					}
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
					rates={rates}
					onSelectSpeed={onSpeed}
					onSettings={onSettings}
					onShortcuts={onShortcuts}
				/>
			</div>
			{level === "hard" && !completed ? (
				<div className="relative rounded-xl border border-primary bg-card px-3 pt-2 shadow-sm focus-within:ring-2 focus-within:ring-ring/30">
					<label htmlFor={inputId} className="block">
						<span className="text-xs font-bold uppercase tracking-wide text-muted-foreground">
							{m["videos.dictation.input.hint"]()}
						</span>
						<Textarea
							id={inputId}
							value={composition ?? draft}
							onCompositionStart={(event) =>
								setComposition(event.currentTarget.value)
							}
							onCompositionEnd={(event) => {
								setComposition(null);
								onDraftChange(event.currentTarget.value);
							}}
							onChange={(event) => {
								const value = event.currentTarget.value;
								if (composition !== null) {
									setComposition(value);
								} else {
									onDraftChange(value);
								}
							}}
							rows={4}
							className="min-h-24 resize-none border-0 bg-transparent px-0 pt-1 pr-7 pb-3 shadow-none focus-visible:ring-0 dark:bg-transparent"
							autoComplete="off"
							autoCapitalize="off"
							spellCheck={false}
						/>
					</label>
				</div>
			) : null}
			<ul className="flex flex-wrap gap-1.5">
				{wordSlots.map((slot) => {
					const index = slot.blankIndex;
					if (index === undefined) {
						return (
							<li key={slot.id} className="self-end px-1 py-1 text-sm">
								{slot.word}
							</li>
						);
					}
					const correct =
						!slot.revealed && (completed || slot.status === "correct");
					const incorrect =
						!correct && !slot.revealed && slot.status === "incorrect";
					const visible =
						completed || (correct && showCorrect) || slot.revealed;
					const status = correct
						? m["videos.dictation.feedback.correct"]()
						: slot.revealed
							? m["videos.dictation.feedback.revealed"]()
							: incorrect
								? m["videos.dictation.feedback.incorrect"]()
								: undefined;
					return (
						<li key={slot.id} className="flex flex-col items-center gap-1">
							<Button
								type="button"
								variant="ghost"
								size="icon"
								disabled={completed || slot.revealed || correct}
								onClick={() => onReveal([slot.position])}
								aria-label={m["videos.dictation.hints.reveal"]({
									index: slot.position + 1,
								})}
								className="size-5 text-muted-foreground [&_svg]:size-3.5"
							>
								<Eye />
							</Button>
							{level !== "hard" && !completed ? (
								<BlankInput
									ref={(node) => {
										inputs.current[index] = node;
									}}
									aria-label={m["videos.dictation.input.word"]({
										index: slot.position + 1,
									})}
									value={answers[index] ?? ""}
									onAnswer={(value) => updateAnswers(index, [value])}
									onKeyDown={(event) => {
										if (event.nativeEvent.isComposing) return;
										if (event.key === " ") {
											event.preventDefault();
											inputs.current[index + 1]?.focus();
										}
										if (
											event.key === "Backspace" &&
											event.currentTarget.value === ""
										) {
											const previous = inputs.current[index - 1];
											if (previous) {
												event.preventDefault();
												previous.focus();
											}
										}
									}}
									onPaste={(event) => {
										event.preventDefault();
										const words = dictationInputWords(
											event.clipboardData.getData("text"),
										);
										if (words.length) updateAnswers(index, words);
									}}
								/>
							) : null}
							<span
								className={cn(
									"inline-flex items-center gap-1 rounded-sm border border-border bg-card px-1.5 py-1 font-mono text-xs tracking-widest",
									correct &&
										"border-green-600/40 bg-green-600/10 text-green-700 dark:text-green-400",
									incorrect &&
										"border-destructive/40 bg-destructive/10 text-destructive",
									slot.revealed &&
										!correct &&
										"border-primary/40 bg-primary/10 text-primary",
								)}
							>
								{visible ? slot.word : maskWord(slot.word)}
								{correct ? (
									<Check aria-hidden="true" className="size-3" />
								) : null}
								{incorrect ? <X aria-hidden="true" className="size-3" /> : null}
								{slot.revealed && !correct ? (
									<Eye aria-hidden="true" className="size-3" />
								) : null}
								{status ? <span className="sr-only"> {status}</span> : null}
							</span>
						</li>
					);
				})}
			</ul>
			<output
				className={cn(
					"text-sm font-medium text-green-700 dark:text-green-400",
					!completed && "sr-only",
				)}
			>
				{completed ? m["videos.dictation.feedback.completed"]() : null}
			</output>
			{children}
			{!completed ? (
				<>
					<p className="text-xs text-muted-foreground">
						{m["videos.dictation.hints.note"]()}
					</p>
					<AlertDialog>
						<AlertDialogTrigger asChild>
							<Button type="button" className="btn btn-accent w-full uppercase">
								{m["videos.dictation.hints.revealAll"]()}
							</Button>
						</AlertDialogTrigger>
						<AlertDialogContent>
							<AlertDialogHeader>
								<AlertDialogTitle>
									{m["videos.dictation.hints.revealAll"]()}
								</AlertDialogTitle>
								<AlertDialogDescription>
									{m["videos.dictation.hints.confirm"]()}
								</AlertDialogDescription>
							</AlertDialogHeader>
							<AlertDialogFooter>
								<AlertDialogCancel>
									{m["common.actions.cancel"]()}
								</AlertDialogCancel>
								<AlertDialogAction onClick={onRevealAll}>
									{m["videos.dictation.hints.submit"]()}
								</AlertDialogAction>
							</AlertDialogFooter>
						</AlertDialogContent>
					</AlertDialog>
				</>
			) : null}
			<Button
				type="button"
				onClick={onNextSentence}
				disabled={!onNextSentence}
				className="btn btn-primary w-full uppercase"
			>
				{m["videos.dictation.next"]()}
				<ChevronRight data-icon="inline-end" />
			</Button>
		</section>
	);
}

function BlankInput({
	value,
	onAnswer,
	...props
}: ComponentProps<typeof Input> & {
	value: string;
	onAnswer: (value: string) => void;
}) {
	const [composition, setComposition] = useState<string | null>(null);
	return (
		<Input
			{...props}
			value={composition ?? value}
			className="w-20"
			autoComplete="off"
			autoCapitalize="off"
			spellCheck={false}
			onCompositionStart={(event) => setComposition(event.currentTarget.value)}
			onCompositionEnd={(event) => {
				setComposition(null);
				onAnswer(event.currentTarget.value);
			}}
			onChange={(event) => {
				if (composition !== null) setComposition(event.currentTarget.value);
				else onAnswer(event.currentTarget.value);
			}}
		/>
	);
}
