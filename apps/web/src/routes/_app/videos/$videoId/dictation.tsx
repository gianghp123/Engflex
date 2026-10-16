import {
	createFileRoute,
	Link,
	notFound,
	useNavigate,
} from "@tanstack/react-router";
import { cn } from "cn";
import { ArrowLeft, Eye, EyeOff } from "lucide-react";
import { useRef, useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { type CrumbItem, useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import {
	Empty,
	EmptyDescription,
	EmptyHeader,
	EmptyTitle,
} from "#/components/ui/empty";
import { DictationPanel } from "#/features/videos/components/dictation/dictation-panel";
import { DictationSummaryDialog } from "#/features/videos/components/dictation/dictation-summary-dialog";
import type { WordSlotState } from "#/features/videos/components/dictation/dictation-word-input";
import { VideoPanel } from "#/features/videos/components/dictation/video-panel";
import { TranscriptPanel } from "#/features/videos/components/transcript-panel";
import { VideoHubSkeleton } from "#/features/videos/components/video-hub-skeleton";
import {
	YouTubePlayer,
	type YouTubePlayerRef,
} from "#/features/videos/components/youtube-player";
import {
	cleanWordForComparison,
	type DictationLevel,
	extractYouTubeId,
	splitWords,
} from "#/features/videos/dictation";
import { videoDetailQueryOptions } from "#/features/videos/queries";
import { DockGuide } from "#/features/vocabulary/components/dock-guide";
import { TextActionDock } from "#/features/vocabulary/components/text-action-dock";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/videos/$videoId/dictation")({
	loader: async ({ context: { queryClient }, params: { videoId } }) => {
		try {
			return await queryClient.query(videoDetailQueryOptions(videoId));
		} catch {
			throw notFound();
		}
	},
	pendingComponent: DictationPending,
	component: DictationPage,
});

function DictationPending() {
	return (
		<PageLayout>
			<VideoHubSkeleton />
		</PageLayout>
	);
}

const GRID = {
	both: "xl:grid-cols-[minmax(0,6fr)_minmax(0,5fr)_minmax(0,3.5fr)]",
	media: "xl:grid-cols-[minmax(0,7fr)_minmax(0,5fr)]",
	transcript: "xl:grid-cols-[minmax(0,7fr)_minmax(0,5fr)]",
	none: "mx-auto w-full max-w-3xl",
} as const;

const SPEEDS = [0.75, 1, 1.25, 1.5];

interface SentenceProgressState {
	completed: boolean;
	wordStates: WordSlotState[];
}

function DictationPage() {
	const navigate = useNavigate();
	const video = Route.useLoaderData();
	const transcripts = video?.transcripts ?? [];

	const playerRef = useRef<YouTubePlayerRef>(null);

	const [showMedia, setShowMedia] = useState(true);
	const [showTranscript, setShowTranscript] = useState(true);
	const [level, setLevel] = useState<DictationLevel>("normal");
	const [speed, setSpeed] = useState(1);
	const [currentTime, setCurrentTime] = useState(0);

	const [activeIndex, setActiveIndex] = useState(0);
	const [activeWordIndex, setActiveWordIndex] = useState(0);
	const [currentInput, setCurrentInput] = useState("");
	const [revealed, setRevealed] = useState<number[]>([]);
	const [sentenceStates, setSentenceStates] = useState<
		Record<number, SentenceProgressState>
	>({});
	const [showSummary, setShowSummary] = useState(false);

	const crumbs: CrumbItem[] = [
		{ label: m["nav.item.videos"](), to: APP_ROUTES.VIDEOS.LIST },
		...(video?.category
			? [
					{
						label: video.category.name,
						to: APP_ROUTES.VIDEOS.CATEGORY,
						params: { categorySlug: video.category.slug },
					},
				]
			: []),
		{ label: video?.title ?? "" },
	];
	useBreadcrumbs(crumbs);

	if (!video) return null;

	if (transcripts.length === 0) {
		return (
			<PageLayout>
				<Empty className="surface-card items-center text-center">
					<MoMascot variant="confused" size={72} />
					<EmptyHeader>
						<EmptyTitle>{m["videos.dictation.empty.title"]()}</EmptyTitle>
						<EmptyDescription>
							{m["videos.dictation.empty.body"]()}
						</EmptyDescription>
					</EmptyHeader>
					<Button asChild variant="outline" className="btn btn-outline mt-4">
						<Link to={APP_ROUTES.VIDEOS.LIST}>
							<ArrowLeft data-icon="inline-start" />
							{m["videos.dictation.empty.back"]()}
						</Link>
					</Button>
				</Empty>
			</PageLayout>
		);
	}

	const currentLine = transcripts[activeIndex];
	const words = splitWords(currentLine?.content ?? "");
	const currentSentenceState = sentenceStates[activeIndex] ?? {
		completed: false,
		wordStates: [],
	};
	const isSentenceCompleted = currentSentenceState.completed;

	const youtubeId = extractYouTubeId(video.videoUrl);

	const handlePlaySentence = () => {
		if (!currentLine) return;
		playerRef.current?.playSegment(
			currentLine.startTimestamp,
			currentLine.endTimestamp,
		);
	};

	const handleReplaySentence = () => {
		if (!currentLine) return;
		playerRef.current?.playSegment(
			currentLine.startTimestamp,
			currentLine.endTimestamp,
		);
	};

	const handleSpeedCycle = () => {
		const nextIndex = (SPEEDS.indexOf(speed) + 1) % SPEEDS.length;
		const nextSpeed = SPEEDS[nextIndex];
		setSpeed(nextSpeed);
		playerRef.current?.setSpeed(nextSpeed);
	};

	const selectLine = (newIndex: number) => {
		setActiveIndex(newIndex);
		const targetState = sentenceStates[newIndex];
		const targetWords = splitWords(transcripts[newIndex]?.content ?? "");
		if (targetState?.completed) {
			setActiveWordIndex(targetWords.length);
			setCurrentInput("");
		} else {
			const existing = targetState?.wordStates ?? [];
			setActiveWordIndex(existing.length);
			setCurrentInput("");
		}
		setRevealed([]);

		const targetLine = transcripts[newIndex];
		if (targetLine) {
			playerRef.current?.seekTo(targetLine.startTimestamp);
		}
	};

	const handleWordSubmit = (wordText: string) => {
		if (isSentenceCompleted || activeWordIndex >= words.length) return;
		const targetWord = words[activeWordIndex];
		const isCorrect =
			cleanWordForComparison(wordText) === cleanWordForComparison(targetWord);

		const updatedWordStates = [...currentSentenceState.wordStates];
		updatedWordStates[activeWordIndex] = {
			status: isCorrect ? "correct" : "incorrect",
			input: wordText,
		};

		const nextWordIdx = activeWordIndex + 1;
		const done = nextWordIdx >= words.length;

		setSentenceStates((prev) => ({
			...prev,
			[activeIndex]: {
				completed: done,
				wordStates: updatedWordStates,
			},
		}));

		setActiveWordIndex(nextWordIdx);
		setCurrentInput("");
	};

	const handleWordBackspace = () => {
		if (isSentenceCompleted || activeWordIndex <= 0) return;
		const prevIdx = activeWordIndex - 1;
		const prevInput = currentSentenceState.wordStates[prevIdx]?.input ?? "";

		const updatedWordStates = [...currentSentenceState.wordStates];
		updatedWordStates[prevIdx] = { status: "pending", input: "" };

		setSentenceStates((prev) => ({
			...prev,
			[activeIndex]: {
				completed: false,
				wordStates: updatedWordStates,
			},
		}));

		setActiveWordIndex(prevIdx);
		setCurrentInput(prevInput);
	};

	const handleResetSentence = () => {
		setSentenceStates((prev) => ({
			...prev,
			[activeIndex]: { completed: false, wordStates: [] },
		}));
		setActiveWordIndex(0);
		setCurrentInput("");
		setRevealed([]);
		handlePlaySentence();
	};

	const handleReveal = (index: number) => {
		if (isSentenceCompleted) return;
		setRevealed((curr) =>
			curr.includes(index) ? curr.filter((i) => i !== index) : [...curr, index],
		);

		const targetWord = words[index];
		const updatedWordStates = [...currentSentenceState.wordStates];
		updatedWordStates[index] = {
			status: "incorrect",
			input: targetWord,
		};

		let nextIdx = activeWordIndex;
		if (index === activeWordIndex) {
			nextIdx = activeWordIndex + 1;
		}

		const done = nextIdx >= words.length;

		setSentenceStates((prev) => ({
			...prev,
			[activeIndex]: {
				completed: done,
				wordStates: updatedWordStates,
			},
		}));

		setActiveWordIndex(nextIdx);
		setCurrentInput("");
	};

	const handleRevealAll = () => {
		if (isSentenceCompleted) return;
		setRevealed(words.map((_, i) => i));

		const updatedWordStates = words.map((w, i) => {
			const existing = currentSentenceState.wordStates[i];
			if (existing && existing.status === "correct") return existing;
			return { status: "incorrect" as const, input: w };
		});

		setSentenceStates((prev) => ({
			...prev,
			[activeIndex]: {
				completed: true,
				wordStates: updatedWordStates,
			},
		}));

		setActiveWordIndex(words.length);
		setCurrentInput("");
	};

	const handleNextSentence = () => {
		if (activeIndex < transcripts.length - 1) {
			selectLine(activeIndex + 1);
			const nextLine = transcripts[activeIndex + 1];
			if (nextLine) {
				playerRef.current?.playSegment(
					nextLine.startTimestamp,
					nextLine.endTimestamp,
				);
			}
		} else {
			setShowSummary(true);
		}
	};

	const handlePrevSentence = () => {
		if (activeIndex > 0) {
			selectLine(activeIndex - 1);
			const prevLine = transcripts[activeIndex - 1];
			if (prevLine) {
				playerRef.current?.playSegment(
					prevLine.startTimestamp,
					prevLine.endTimestamp,
				);
			}
		}
	};

	const completedCount = Object.values(sentenceStates).filter(
		(s) => s.completed,
	).length;
	const overallProgress =
		transcripts.length > 0 ? (completedCount / transcripts.length) * 100 : 0;

	let totalWordsInLesson = 0;
	let totalCorrectWords = 0;
	for (let i = 0; i < transcripts.length; i++) {
		const sentenceWords = splitWords(transcripts[i]?.content ?? "");
		totalWordsInLesson += sentenceWords.length;
		const st = sentenceStates[i];
		if (st) {
			totalCorrectWords += st.wordStates.filter(
				(ws) => ws.status === "correct",
			).length;
		}
	}
	const totalIncorrectWords = Math.max(
		0,
		totalWordsInLesson - totalCorrectWords,
	);
	const overallAccuracy =
		totalWordsInLesson > 0 ? (totalCorrectWords / totalWordsInLesson) * 100 : 0;

	const handleRestart = () => {
		setSentenceStates({});
		setActiveIndex(0);
		setActiveWordIndex(0);
		setCurrentInput("");
		setRevealed([]);
		setShowSummary(false);
		const firstLine = transcripts[0];
		if (firstLine) {
			playerRef.current?.seekTo(firstLine.startTimestamp);
		}
	};

	const layout =
		showMedia && showTranscript
			? GRID.both
			: showMedia
				? GRID.media
				: showTranscript
					? GRID.transcript
					: GRID.none;

	return (
		<PageLayout>
			<div className="flex flex-wrap items-center justify-end gap-2">
				<div className="mr-auto">
					<DockGuide />
				</div>
				<Button
					type="button"
					variant="outline"
					aria-pressed={!showMedia}
					onClick={() => setShowMedia((value) => !value)}
					className="btn btn-outline"
				>
					{showMedia ? (
						<EyeOff data-icon="inline-start" />
					) : (
						<Eye data-icon="inline-start" />
					)}
					{showMedia
						? m["videos.dictation.toolbar.hideMedia"]()
						: m["videos.dictation.toolbar.showMedia"]()}
				</Button>
				<Button
					type="button"
					variant="outline"
					aria-pressed={!showTranscript}
					onClick={() => setShowTranscript((value) => !value)}
					className="btn btn-outline"
				>
					{showTranscript ? (
						<EyeOff data-icon="inline-start" />
					) : (
						<Eye data-icon="inline-start" />
					)}
					{showTranscript
						? m["videos.dictation.toolbar.hideTranscript"]()
						: m["videos.dictation.toolbar.showTranscript"]()}
				</Button>
			</div>

			<TextActionDock source={{ type: "video", id: video?.id }}>
				<div className={cn("grid items-start gap-3", layout)}>
					{showMedia ? (
						<VideoPanel
							video={video}
							currentTime={currentTime}
							onStart={handlePlaySentence}
							onReplay={handleReplaySentence}
						>
							{youtubeId ? (
								<YouTubePlayer
									ref={playerRef}
									videoId={youtubeId}
									onTimeUpdate={setCurrentTime}
								/>
							) : null}
						</VideoPanel>
					) : null}

					<DictationPanel
						level={level}
						onLevelChange={setLevel}
						speed={speed}
						words={words}
						revealed={revealed}
						wordStates={currentSentenceState.wordStates}
						activeWordIndex={activeWordIndex}
						currentInput={currentInput}
						isCompleted={isSentenceCompleted}
						onInputChange={setCurrentInput}
						onWordSubmit={handleWordSubmit}
						onWordBackspace={handleWordBackspace}
						onEnterPress={handleNextSentence}
						onResetSentence={handleResetSentence}
						onReveal={handleReveal}
						onRevealAll={handleRevealAll}
						onPrev={handlePrevSentence}
						onNextSentence={handleNextSentence}
						onNext={handleNextSentence}
						onPlaySentence={handlePlaySentence}
						onReplaySentence={handleReplaySentence}
						onSpeed={handleSpeedCycle}
					/>

					{showTranscript ? (
						<TranscriptPanel
							transcripts={transcripts}
							activeIndex={activeIndex}
							progress={overallProgress}
							onSelect={selectLine}
						/>
					) : null}
				</div>
			</TextActionDock>

			<DictationSummaryDialog
				open={showSummary}
				accuracy={overallAccuracy}
				correctWords={totalCorrectWords}
				incorrectWords={totalIncorrectWords}
				onRestart={handleRestart}
				onBackToHub={() => navigate({ to: APP_ROUTES.VIDEOS.LIST })}
			/>
		</PageLayout>
	);
}
