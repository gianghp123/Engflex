import { useAuth } from "@clerk/tanstack-react-start";
import type {
	VideoAttemptResponse,
	VideoExerciseResponse,
} from "@engflex/contracts";
import { useQuery, useQueryClient } from "@tanstack/react-query";
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
import type { YouTubePlayerRef } from "#/features/videos/components/youtube-player";
import {
	cleanWordForComparison,
	type DictationLevel,
	extractYouTubeId,
	splitWords,
} from "#/features/videos/dictation";
import {
	openVideoAttemptQueryOptions,
	useCompleteVideoAttempt,
	useSaveVideoProgress,
	useStartVideoAttempt,
	videoDetailQueryOptions,
	videoKeys,
} from "#/features/videos/queries";
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

function parseInitialSentenceStates(
	transcripts: { id: string }[],
	result?: VideoAttemptResponse["result"],
): Record<number, SentenceProgressState> {
	if (!result?.transcripts) return {};
	const map: Record<number, SentenceProgressState> = {};
	transcripts.forEach((t, idx) => {
		const raw = result.transcripts?.[t.id] as
			| { completed?: boolean; wordStates?: WordSlotState[] }
			| undefined;
		if (raw) {
			map[idx] = {
				completed: raw.completed ?? false,
				wordStates: raw.wordStates ?? [],
			};
		}
	});
	return map;
}

function DictationPage() {
	const video = Route.useLoaderData();
	const { isLoaded } = useAuth();

	const openQuery = useQuery({
		...openVideoAttemptQueryOptions(video?.id ?? "", "dictation"),
		enabled: isLoaded && !!video?.id,
	});

	if (!video) return null;

	if (!isLoaded || openQuery.isPending) {
		return <DictationPending />;
	}

	return (
		<DictationWorkspace
			key={video.id}
			video={video}
			initialAttempt={openQuery.data ?? null}
		/>
	);
}

function DictationWorkspace({
	video,
	initialAttempt,
}: {
	video: VideoExerciseResponse;
	initialAttempt: VideoAttemptResponse | null;
}) {
	const navigate = useNavigate();
	const queryClient = useQueryClient();
	const transcripts = video.transcripts ?? [];

	const playerRef = useRef<YouTubePlayerRef>(null);

	const [currentAttemptId, setCurrentAttemptId] = useState<string | null>(
		initialAttempt?.id ?? null,
	);
	const [showMedia, setShowMedia] = useState(true);
	const [showTranscript, setShowTranscript] = useState(true);
	const [level, setLevel] = useState<DictationLevel>(() => {
		const lvl = initialAttempt?.result?.level;
		return lvl === "easy" || lvl === "hard" ? lvl : "normal";
	});
	const [speed, setSpeed] = useState(1);

	const [activeIndex, setActiveIndex] = useState(() => {
		const seq = initialAttempt?.result?.activeSequence ?? 0;
		return seq >= 0 && seq < transcripts.length ? seq : 0;
	});

	const [sentenceStates, setSentenceStates] = useState<
		Record<number, SentenceProgressState>
	>(() => parseInitialSentenceStates(transcripts, initialAttempt?.result));

	const [activeWordIndex, setActiveWordIndex] = useState(() => {
		const targetWords = splitWords(transcripts[activeIndex]?.content ?? "");
		const targetState = sentenceStates[activeIndex];
		if (targetState?.completed) {
			return targetWords.length;
		}
		return targetState?.wordStates?.length ?? 0;
	});

	const [currentInput, setCurrentInput] = useState("");
	const [revealed, setRevealed] = useState<number[]>([]);
	const [showSummary, setShowSummary] = useState(false);

	const startAttempt = useStartVideoAttempt();
	const saveProgress = useSaveVideoProgress();
	const completeAttempt = useCompleteVideoAttempt();

	const persistProgress = async (
		targetIndex: number,
		statesToSave: Record<number, SentenceProgressState>,
		currentLevel = level,
	) => {
		let attemptId = currentAttemptId;
		if (!attemptId) {
			try {
				const res = await startAttempt.mutateAsync({
					videoId: video.id,
					type: "dictation",
				});
				attemptId = res.id;
				setCurrentAttemptId(res.id);
			} catch {
				return;
			}
		}

		const transcriptsPayload: Record<string, unknown> = {};
		const completedIds: string[] = [];
		transcripts.forEach((t, idx) => {
			const st = statesToSave[idx];
			if (st) {
				transcriptsPayload[t.id] = {
					completed: st.completed,
					wordStates: st.wordStates,
				};
				if (st.completed) completedIds.push(t.id);
			}
		});

		try {
			await saveProgress.mutateAsync({
				videoId: video.id,
				attemptId,
				req: {
					activeSequence: targetIndex,
					activeTranscriptId: transcripts[targetIndex]?.id,
					level: currentLevel,
					completedTranscriptIds: completedIds,
					transcripts: transcriptsPayload,
				},
			});
		} catch {}
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

	const handleFinish = async () => {
		setShowSummary(true);
		let attemptId = currentAttemptId;
		if (!attemptId) {
			try {
				const res = await startAttempt.mutateAsync({
					videoId: video.id,
					type: "dictation",
				});
				attemptId = res.id;
				setCurrentAttemptId(res.id);
			} catch {
				return;
			}
		}

		await persistProgress(activeIndex, sentenceStates);

		try {
			await completeAttempt.mutateAsync({
				videoId: video.id,
				attemptId,
				req: {
					score: Math.round(overallAccuracy),
				},
			});
			void queryClient.invalidateQueries({
				queryKey: videoKeys.detail(video.id),
			});
			void queryClient.invalidateQueries({
				queryKey: videoKeys.all,
			});
		} catch {}
	};

	const crumbs: CrumbItem[] = [
		{ label: m["nav.item.videos"](), to: APP_ROUTES.VIDEOS.LIST },
		...(video.category
			? [
					{
						label: video.category.name,
						to: APP_ROUTES.VIDEOS.CATEGORY,
						params: { categorySlug: video.category.slug },
					},
				]
			: []),
		{ label: video.title },
	];
	useBreadcrumbs(crumbs);

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

	const handlePlayerReady = () => {
		if (activeIndex > 0 && currentLine) {
			playerRef.current?.seekTo(currentLine.startTimestamp);
		}
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
			playerRef.current?.playSegment(
				targetLine.startTimestamp,
				targetLine.endTimestamp,
			);
		}
		void persistProgress(newIndex, sentenceStates);
	};

	const handleNextSentence = () => {
		if (activeIndex < transcripts.length - 1) {
			selectLine(activeIndex + 1);
		} else {
			void handleFinish();
		}
	};

	const handlePrevSentence = () => {
		if (activeIndex > 0) {
			selectLine(activeIndex - 1);
		}
	};

	const handlePlaySentence = () => {
		if (!currentLine) return;
		const current = playerRef.current?.getCurrentTime() ?? 0;
		if (current >= currentLine.endTimestamp - 0.3) {
			handleNextSentence();
			return;
		}
		if (current < currentLine.startTimestamp - 0.5) {
			playerRef.current?.playSegment(
				currentLine.startTimestamp,
				currentLine.endTimestamp,
			);
			return;
		}
		playerRef.current?.playSegment(current, currentLine.endTimestamp);
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

		const nextStates = {
			...sentenceStates,
			[activeIndex]: {
				completed: done,
				wordStates: updatedWordStates,
			},
		};

		setSentenceStates(nextStates);
		setActiveWordIndex(nextWordIdx);
		setCurrentInput("");

		if (done) {
			void persistProgress(activeIndex, nextStates);
		}
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
		const nextStates = {
			...sentenceStates,
			[activeIndex]: { completed: false, wordStates: [] },
		};
		setSentenceStates(nextStates);
		setActiveWordIndex(0);
		setCurrentInput("");
		setRevealed([]);
		handlePlaySentence();
		void persistProgress(activeIndex, nextStates);
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

		const nextStates = {
			...sentenceStates,
			[activeIndex]: {
				completed: done,
				wordStates: updatedWordStates,
			},
		};

		setSentenceStates(nextStates);
		setActiveWordIndex(nextIdx);
		setCurrentInput("");

		if (done) {
			void persistProgress(activeIndex, nextStates);
		}
	};

	const handleRevealAll = () => {
		if (isSentenceCompleted) return;
		setRevealed(words.map((_, i) => i));

		const updatedWordStates = words.map((w, i) => {
			const existing = currentSentenceState.wordStates[i];
			if (existing && existing.status === "correct") return existing;
			return { status: "incorrect" as const, input: w };
		});

		const nextStates = {
			...sentenceStates,
			[activeIndex]: {
				completed: true,
				wordStates: updatedWordStates,
			},
		};

		setSentenceStates(nextStates);
		setActiveWordIndex(words.length);
		setCurrentInput("");
		void persistProgress(activeIndex, nextStates);
	};

	const handleRestart = async () => {
		setSentenceStates({});
		setActiveIndex(0);
		setActiveWordIndex(0);
		setCurrentInput("");
		setRevealed([]);
		setShowSummary(false);
		setCurrentAttemptId(null);
		const firstLine = transcripts[0];
		if (firstLine) {
			playerRef.current?.seekTo(firstLine.startTimestamp);
		}
		try {
			const res = await startAttempt.mutateAsync({
				videoId: video.id,
				type: "dictation",
			});
			setCurrentAttemptId(res.id);
		} catch {}
	};

	const handleLevelChange = (newLevel: DictationLevel) => {
		setLevel(newLevel);
		void persistProgress(activeIndex, sentenceStates, newLevel);
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

			<div className={cn("grid items-start gap-3", layout)}>
				{showMedia ? (
					<VideoPanel
						video={video}
						youtubeId={youtubeId}
						playerRef={playerRef}
						activeSegment={
							currentLine
								? {
										start: currentLine.startTimestamp,
										end: currentLine.endTimestamp,
									}
								: null
						}
						onPlayerReady={handlePlayerReady}
						onStart={handlePlaySentence}
						onReplay={handleReplaySentence}
						onPlayPastEnd={handleNextSentence}
					/>
				) : null}

				<DictationPanel
					level={level}
					onLevelChange={handleLevelChange}
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
