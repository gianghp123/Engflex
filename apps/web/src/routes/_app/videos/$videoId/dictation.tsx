import type {
	VideoAttemptResponse,
	VideoExerciseResponse,
} from "@engflex/contracts";
import { createFileRoute, Link, notFound } from "@tanstack/react-router";
import { cn } from "cn";
import { ArrowLeft, Eye, EyeOff } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { type CrumbItem, useBreadcrumbs } from "#/app/breadcrumbs";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
} from "#/components/ui/dialog";
import {
	Empty,
	EmptyDescription,
	EmptyHeader,
	EmptyTitle,
} from "#/components/ui/empty";
import { BookmarkButton } from "#/features/bookmarks/bookmark-button";
import { DictationPanel } from "#/features/videos/components/dictation/dictation-panel";
import {
	DictationNotes,
	DictationWordResults,
} from "#/features/videos/components/dictation/dictation-review";
import { VideoPanel } from "#/features/videos/components/dictation/video-panel";
import { PracticeToolsDialogs } from "#/features/videos/components/practice-tools-dialogs";
import { TranscriptPanel } from "#/features/videos/components/transcript-panel";
import { VideoHubSkeleton } from "#/features/videos/components/video-hub-skeleton";
import type { YouTubePlayerRef } from "#/features/videos/components/youtube-player";
import { checkDictation, extractYouTubeId } from "#/features/videos/dictation";
import { playDictationFeedback } from "#/features/videos/dictation-sound";
import {
	useOpenDictationAttempt,
	videoDetailQueryOptions,
} from "#/features/videos/queries";
import {
	DICTATION_SPEEDS,
	dictationPreferences,
} from "#/features/videos/store";
import {
	type LinePractice,
	useDictationAttempt,
} from "#/features/videos/use-dictation-attempt";
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

const EMPTY_PRACTICE: LinePractice = {
	draft: "",
	revealed: [],
	completed: false,
};

function DictationPage() {
	const video = Route.useLoaderData();
	return <DictationSession key={video.id} video={video} />;
}

function DictationSession({ video }: { video: VideoExerciseResponse }) {
	const hasLines = Boolean(video.transcripts?.length);
	const attempt = useOpenDictationAttempt(video.id, hasLines);
	if (
		hasLines &&
		(attempt.isPending || (!attempt.isFetchedAfterMount && attempt.isFetching))
	) {
		return <DictationPending />;
	}
	if (hasLines && attempt.isError) {
		return (
			<PageLayout>
				<Empty className="surface-card">
					<EmptyTitle>{m["videos.dictation.attempt.loadFailed"]()}</EmptyTitle>
					<SubmitButton
						pending={attempt.isFetching}
						onClick={() => void attempt.refetch()}
					>
						{m["common.actions.retry"]()}
					</SubmitButton>
				</Empty>
			</PageLayout>
		);
	}
	return (
		<DictationWorkspace video={video} initialAttempt={attempt.data ?? null} />
	);
}

function DictationWorkspace({
	video,
	initialAttempt,
}: {
	video: VideoExerciseResponse;
	initialAttempt: VideoAttemptResponse | null;
}) {
	const playerRef = useRef<YouTubePlayerRef>(null);
	const submittedSounds = useRef(new Set<string>());
	const [playing, setPlaying] = useState(false);
	const [rates, setRates] = useState(DICTATION_SPEEDS);
	const [detailId, setDetailId] = useState<string | null>(null);

	const [settings, setSettings] = dictationPreferences.usePreferences();
	const { showMedia, showTranscript, speed } = settings;
	const isPlaying = showMedia && playing;
	const setShowMedia = (update: (value: boolean) => boolean) =>
		setSettings({ showMedia: update(showMedia) });
	const setShowTranscript = (update: (value: boolean) => boolean) =>
		setSettings({ showTranscript: update(showTranscript) });
	const [toolsOpen, setToolsOpen] = useState<"settings" | "shortcuts" | null>(
		null,
	);
	useEffect(() => {
		playerRef.current?.setSpeed(speed);
	}, [speed]);

	const {
		activeIndex,
		lines: transcripts,
		level,
		levelLocked,
		setLevel,
		finished,
		score,
		restart,
		setActiveIndex,
		practice,
		setPractice,
		completedCount,
		isSaving,
		saveFailed,
		retrySave,
	} = useDictationAttempt(video, initialAttempt);

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

	const currentLine = transcripts[activeIndex] ?? transcripts[0];
	const currentPractice = practice[currentLine.id] ?? EMPTY_PRACTICE;
	const currentCheck = checkDictation(
		currentLine.content,
		currentPractice.draft,
		level,
		currentPractice.answers,
	);
	const detailsIndex = transcripts.findIndex((line) => line.id === detailId);
	const detailLine = transcripts[detailsIndex];
	const detailPractice = detailLine ? practice[detailLine.id] : undefined;
	const progress = (completedCount / transcripts.length) * 100;
	const updatePractice = (next: LinePractice) => {
		if (currentPractice.completed) return;
		setPractice((previous) => ({
			...previous,
			[currentLine.id]: next,
		}));
		if (
			next.completed &&
			settings.soundFeedback &&
			!submittedSounds.current.has(currentLine.id)
		) {
			submittedSounds.current.add(currentLine.id);
			playDictationFeedback(Boolean(next.revealAll || next.revealed.length));
		}
	};
	const updateAnswer = (answer: Pick<LinePractice, "draft" | "answers">) => {
		const completed = checkDictation(
			currentLine.content,
			answer.draft,
			level,
			answer.answers,
		).complete;
		updatePractice({ ...currentPractice, ...answer, completed });
	};
	const revealAll = () => {
		const remaining = currentCheck.slots.filter(
			(slot) => slot.blankIndex !== undefined && slot.status !== "correct",
		);
		updatePractice({
			...currentPractice,
			revealed: [
				...new Set([
					...currentPractice.revealed,
					...remaining.map((slot) => slot.position),
				]),
			],
			revealAll: true,
			completed: true,
		});
	};
	const youtubeId = extractYouTubeId(video.videoUrl);

	const handlePlayerReady = () => {
		const supported = playerRef.current?.getAvailableRates() ?? [];
		const available = DICTATION_SPEEDS.filter((rate) =>
			supported.includes(rate),
		);
		setRates(available.length ? available : [1]);
		playerRef.current?.setSpeed(available.includes(speed) ? speed : 1);
		playerRef.current?.seekTo(currentLine.startTimestamp);
	};

	const selectLine = (newIndex: number) => {
		setActiveIndex(newIndex);
		const targetLine = transcripts[newIndex];
		if (targetLine) {
			playerRef.current?.playSegment(
				targetLine.startTimestamp,
				targetLine.endTimestamp,
			);
		}
	};

	const handleNextSentence = () => {
		if (activeIndex < transcripts.length - 1) {
			selectLine(activeIndex + 1);
		}
	};

	const handlePrevSentence = () => {
		if (activeIndex > 0) {
			selectLine(activeIndex - 1);
		}
	};

	const handlePlaySentence = () => {
		if (isPlaying) {
			playerRef.current?.pause();
			return;
		}
		const current = playerRef.current?.getCurrentTime() ?? 0;
		if (
			current >= currentLine.endTimestamp - 0.1 ||
			current < currentLine.startTimestamp
		) {
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

	const effectiveSpeed = rates.includes(speed) ? speed : 1;

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
			{finished ? (
				<div className="surface-card flex flex-wrap items-center gap-3 p-3">
					<output>
						{score == null
							? m["videos.dictation.review.noScore"]()
							: m["videos.dictation.review.finalScore"]({ score })}
					</output>
					<span>
						{m["videos.dictation.review.hintCount"]({
							count: Object.values(practice).reduce(
								(total, state) => total + state.revealed.length,
								0,
							),
						})}
					</span>
					<Button
						type="button"
						variant="outline"
						onClick={() => {
							restart();
							submittedSounds.current.clear();
							setDetailId(null);
							playerRef.current?.pause();
							playerRef.current?.seekTo(transcripts[0].startTimestamp);
						}}
					>
						{m["videos.dictation.review.restart"]()}
					</Button>
				</div>
			) : null}
			{saveFailed ? (
				<div className="flex flex-wrap items-center justify-end gap-2">
					<span role="alert">{m["videos.dictation.attempt.saveFailed"]()}</span>
					<SubmitButton pending={isSaving} onClick={retrySave}>
						{m["common.actions.retry"]()}
					</SubmitButton>
				</div>
			) : null}
			<div className="flex flex-wrap items-center justify-end gap-2">
				<output className="text-sm text-muted-foreground">
					{isSaving ? m["videos.dictation.attempt.saving"]() : null}
				</output>
				<BookmarkButton type="video" id={video.id} />
				<Button
					type="button"
					variant="outline"
					aria-pressed={!showMedia}
					onClick={() => {
						setPlaying(false);
						setShowMedia((value) => !value);
					}}
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
						onPlayingChange={setPlaying}
						onRateChange={(rate) => {
							if (DICTATION_SPEEDS.includes(rate)) setSettings({ speed: rate });
						}}
						isPlaying={isPlaying}
					/>
				) : null}

				<DictationPanel
					key={`${currentLine.id}:${level}`}
					slots={currentCheck.slots}
					{...currentPractice}
					level={level}
					levelLocked={levelLocked}
					onLevel={setLevel}
					showCorrect={settings.showCorrect}
					isPlaying={isPlaying}
					onDraftChange={(draft) => updateAnswer({ draft })}
					onAnswers={(answers) => updateAnswer({ draft: "", answers })}
					onRevealAll={revealAll}
					onReveal={(positions) =>
						updatePractice({
							...currentPractice,
							revealed: [
								...new Set([...currentPractice.revealed, ...positions]),
							],
						})
					}
					speed={effectiveSpeed}
					rates={rates}
					onPrev={activeIndex > 0 ? handlePrevSentence : undefined}
					onNextSentence={
						activeIndex < transcripts.length - 1
							? handleNextSentence
							: undefined
					}
					onPlaySentence={handlePlaySentence}
					onReplaySentence={handleReplaySentence}
					onSpeed={(rate) => setSettings({ speed: rate })}
					onSettings={() => setToolsOpen("settings")}
					onShortcuts={() => setToolsOpen("shortcuts")}
				>
					{currentPractice.completed ? (
						<DictationNotes
							line={currentLine}
							translation={settings.showTranslation}
							phonetic={settings.showPhonetic}
						/>
					) : null}
				</DictationPanel>

				{showTranscript ? (
					<TranscriptPanel
						transcripts={transcripts}
						activeIndex={activeIndex}
						progress={progress}
						onSelect={selectLine}
						dictationStatus={(id) => ({
							submitted: practice[id]?.completed ?? false,
							hinted: Boolean(practice[id]?.revealed.length),
							earned: practice[id]?.earned,
							total: practice[id]?.total,
						})}
						onDetails={(index) => setDetailId(transcripts[index].id)}
					/>
				) : null}
			</div>

			<Dialog
				open={Boolean(detailLine && detailPractice?.completed)}
				onOpenChange={(open) => {
					if (!open) setDetailId(null);
				}}
			>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{m["videos.dictation.review.details"]()}</DialogTitle>
					</DialogHeader>
					{detailLine && detailPractice ? (
						<>
							<DictationWordResults
								slots={checkDictation(detailLine.content, "", level).slots}
								revealed={detailPractice.revealed}
								scored={detailPractice.total !== undefined}
							/>
							<DictationNotes
								line={detailLine}
								translation={settings.showTranslation}
								phonetic={settings.showPhonetic}
							/>
							<Button
								type="button"
								disabled={!showMedia || !youtubeId}
								onClick={() => {
									setDetailId(null);
									selectLine(detailsIndex);
								}}
							>
								{m["videos.practice.player.replay"]()}
							</Button>
						</>
					) : null}
				</DialogContent>
			</Dialog>

			<PracticeToolsDialogs
				mode="dictation"
				open={toolsOpen}
				onOpenChange={setToolsOpen}
				settings={settings}
				onChange={setSettings}
				onReset={dictationPreferences.reset}
				rates={rates}
			/>
		</PageLayout>
	);
}
