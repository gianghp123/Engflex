import type {
	VideoAttemptResponse,
	VideoExerciseResponse,
} from "@engflex/contracts";
import { createFileRoute } from "@tanstack/react-router";
import { cn } from "cn";
import { Clock3, Eye, EyeOff, Mic, MonitorPlay } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { type CrumbItem, useBreadcrumbs } from "#/app/breadcrumbs";
import { PageLayout } from "#/components/common/page-layout";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { Empty, EmptyTitle } from "#/components/ui/empty";
import { Toggle } from "#/components/ui/toggle";
import { ToggleGroup, ToggleGroupItem } from "#/components/ui/toggle-group";
import { BookmarkButton } from "#/features/bookmarks/bookmark-button";
import { PronunciationPanel } from "#/features/lessons/components/speaking/pronunciation-panel";
import { useRecorder } from "#/features/lessons/hooks/use-recorder";
import { PracticeToolsDialogs } from "#/features/videos/components/practice-tools-dialogs";
import { ShadowingPanel } from "#/features/videos/components/shadowing/shadowing-panel";
import { TranscriptPanel } from "#/features/videos/components/transcript-panel";
import { VideoHubSkeleton } from "#/features/videos/components/video-hub-skeleton";
import type { YouTubePlayerRef } from "#/features/videos/components/youtube-player";
import { extractYouTubeId, pairWords } from "#/features/videos/dictation";
import { formatDuration } from "#/features/videos/format";
import {
	useOpenShadowingAttempt,
	usePronounceShadowing,
	videoDetailQueryOptions,
} from "#/features/videos/queries";
import { shadowingPreferences } from "#/features/videos/store";
import { useShadowingAttempt } from "#/features/videos/use-shadowing-attempt";
import { loadOr404 } from "#/lib/router/route-loader";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/videos/$videoId/shadowing")({
	validateSearch: (search: Record<string, unknown>): { view?: View } =>
		search.view === "watch" ? { view: "watch" } : {},
	loader: ({ context: { queryClient }, params: { videoId } }) =>
		loadOr404(queryClient, videoDetailQueryOptions(videoId)),
	pendingComponent: ShadowingPending,
	component: ShadowingPage,
});

type View = "watch" | "practice";

const VIEW_TOGGLE_ITEM = cn(
	"rounded-full! px-3 text-xs font-extrabold tracking-wide uppercase",
	"data-[state=on]:bg-sidebar data-[state=on]:text-sidebar-foreground",
);

const LINE_TOGGLE =
	"h-8 gap-1.5 px-2 text-xs font-semibold text-muted-foreground data-[state=on]:bg-transparent data-[state=on]:text-foreground";

const SPEEDS = [0.75, 1, 1.25, 1.5];

function ShadowingPending() {
	return (
		<PageLayout>
			<VideoHubSkeleton />
		</PageLayout>
	);
}

function ShadowingPage() {
	const video = Route.useLoaderData();
	return <ShadowingSession key={video.id} video={video} />;
}

function ShadowingSession({ video }: { video: VideoExerciseResponse }) {
	const hasLines = Boolean(video.transcripts?.length);
	const attempt = useOpenShadowingAttempt(video.id, hasLines);
	if (
		hasLines &&
		(attempt.isPending || (!attempt.isFetchedAfterMount && attempt.isFetching))
	) {
		return <ShadowingPending />;
	}
	if (hasLines && attempt.isError) {
		return (
			<PageLayout>
				<Empty className="surface-card">
					<EmptyTitle>{m["videos.shadowing.attempt.loadFailed"]()}</EmptyTitle>
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
		<ShadowingWorkspace video={video} initialAttempt={attempt.data ?? null} />
	);
}

function ShadowingWorkspace({
	video,
	initialAttempt,
}: {
	video: VideoExerciseResponse;
	initialAttempt: VideoAttemptResponse | null;
}) {
	const transcripts = video.transcripts ?? [];

	const playerRef = useRef<YouTubePlayerRef>(null);
	const youtubeId = extractYouTubeId(video.videoUrl ?? "");

	const { view = "practice" } = Route.useSearch();
	const navigate = Route.useNavigate();
	const [settings, setSettings] = shadowingPreferences.usePreferences();
	const {
		activeIndex,
		setActiveIndex,
		completedCount,
		isSaving,
		saveFailed,
		retrySave,
		ensureAttemptId,
	} = useShadowingAttempt(video, initialAttempt);
	const {
		showTranscript,
		showPhonetic,
		showTranslation,
		autoPause,
		largeVideo,
		speed,
	} = settings;
	const setShowTranscript = (update: (value: boolean) => boolean) =>
		setSettings({ showTranscript: update(showTranscript) });
	const setShowPhonetic = (value: boolean) =>
		setSettings({ showPhonetic: value });
	const setShowTranslation = (value: boolean) =>
		setSettings({ showTranslation: value });
	const setAutoPause = (value: boolean) => setSettings({ autoPause: value });
	const setLargeVideo = (value: boolean) => setSettings({ largeVideo: value });
	const [toolsOpen, setToolsOpen] = useState<"settings" | "shortcuts" | null>(
		null,
	);
	useEffect(() => {
		playerRef.current?.setSpeed(speed);
	}, [speed]);
	const [currentTime, setCurrentTime] = useState(0);
	const lastTimeUpdateRef = useRef(0);

	const {
		state: recorderState,
		start: startRecording,
		stop: stopRecording,
		reset: resetRecording,
	} = useRecorder(m["videos.shadowing.record.micDenied"]());
	const recording = recorderState.status === "recording";
	const hasRecording = recorderState.status === "ready";
	const pronounce = usePronounceShadowing();
	const [recordingUrl, setRecordingUrl] = useState<string | null>(null);
	useEffect(() => {
		if (recorderState.status !== "ready") return;
		const url = URL.createObjectURL(recorderState.blob);
		setRecordingUrl(url);
		return () => {
			URL.revokeObjectURL(url);
			setRecordingUrl(null);
		};
	}, [recorderState]);
	// A new line starts a fresh take — the previous recording (if any) is
	// discarded rather than carried over to a sentence it wasn't spoken for.
	// biome-ignore lint/correctness/useExhaustiveDependencies: fires on line change only; resetRecording/pronounce.reset are stable
	useEffect(() => {
		resetRecording();
		pronounce.reset();
	}, [activeIndex]);
	useEffect(() => resetRecording, [resetRecording]);
	const handleRecord = () => {
		if (recording) {
			stopRecording();
		} else {
			pronounce.reset();
			void startRecording();
		}
	};
	const handlePlayRecording = () => {
		if (!recordingUrl) return;
		new Audio(recordingUrl).play().catch(() => {});
	};

	// Auto-submits the moment a take is ready — the same "stop recording →
	// score" flow as lesson read-aloud, just against a transcript line
	// instead of a speaking item's reference text.
	// biome-ignore lint/correctness/useExhaustiveDependencies: fires once per ready blob; the rest are read fresh inside the async callback, not meant to re-trigger the effect
	useEffect(() => {
		if (recorderState.status !== "ready") return;
		const currentLine = transcripts[activeIndex];
		if (!currentLine) return;
		void (async () => {
			const attemptId = await ensureAttemptId();
			pronounce.mutate({
				videoId: video.id,
				attemptId,
				transcriptId: currentLine.id,
				audio: recorderState.blob,
				mime: recorderState.mime,
			});
		})();
	}, [recorderState]);

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

	const line = transcripts[activeIndex];
	const words = line ? pairWords(line.content, line.phonetic) : [];

	const handleTimeUpdate = (t: number) => {
		const now = Date.now();
		if (now - lastTimeUpdateRef.current >= 500) {
			lastTimeUpdateRef.current = now;
			setCurrentTime(Math.floor(t));
		}
	};

	const time = formatDuration(currentTime);
	const twoColumns = showTranscript && !largeVideo;

	const handlePlayerReady = () => {
		playerRef.current?.setSpeed(speed);
		if (activeIndex > 0 && line) {
			playerRef.current?.seekTo(line.startTimestamp);
		}
	};

	const selectLine = (newIndex: number) => {
		setActiveIndex(newIndex);
		const targetLine = transcripts[newIndex];
		if (targetLine) {
			if (autoPause) {
				playerRef.current?.playSegment(
					targetLine.startTimestamp,
					targetLine.endTimestamp,
				);
			} else {
				playerRef.current?.seekTo(targetLine.startTimestamp);
				playerRef.current?.play();
			}
		}
	};

	const handleNextSentence = () => {
		if (activeIndex < transcripts.length - 1) {
			selectLine(activeIndex + 1);
		}
	};

	const handlePlaySentence = () => {
		if (!line) {
			playerRef.current?.play();
			return;
		}
		const current = playerRef.current?.getCurrentTime() ?? 0;
		if (autoPause && current >= line.endTimestamp - 0.3) {
			handleNextSentence();
			return;
		}
		if (current < line.startTimestamp - 0.5) {
			if (autoPause) {
				playerRef.current?.playSegment(line.startTimestamp, line.endTimestamp);
			} else {
				playerRef.current?.seekTo(line.startTimestamp);
				playerRef.current?.play();
			}
			return;
		}
		if (autoPause) {
			playerRef.current?.playSegment(current, line.endTimestamp);
		} else {
			playerRef.current?.play();
		}
	};

	const handleSpeedCycle = () => {
		const nextIndex = (SPEEDS.indexOf(speed) + 1) % SPEEDS.length;
		const nextSpeed = SPEEDS[nextIndex];
		setSettings({ speed: nextSpeed });
		playerRef.current?.setSpeed(nextSpeed);
	};

	return (
		<PageLayout>
			<PracticeToolsDialogs
				mode="shadowing"
				open={toolsOpen}
				onOpenChange={setToolsOpen}
				settings={settings}
				onChange={setSettings}
				onReset={shadowingPreferences.reset}
			/>
			<div className="flex flex-wrap items-center justify-end gap-3">
				<BookmarkButton type="video" id={video.id} />
				<span className="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
					<Clock3 aria-hidden="true" className="size-4" />
					<span aria-hidden="true">{time}</span>
					<span className="sr-only">
						{m["videos.shadowing.toolbar.time"]({ time })}
					</span>
				</span>
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
						? m["videos.shadowing.toolbar.hideTranscript"]()
						: m["videos.shadowing.toolbar.showTranscript"]()}
				</Button>
				<ToggleGroup
					type="single"
					value={view}
					onValueChange={(value) => {
						if (value) {
							void navigate({
								search: value === "watch" ? { view: "watch" } : {},
								replace: true,
							});
						}
					}}
					aria-label={m["videos.shadowing.toolbar.view"]()}
					className="rounded-full border-2 border-border bg-card p-1 shadow-[0_3px_0_var(--border)]"
				>
					<ToggleGroupItem value="watch" className={VIEW_TOGGLE_ITEM}>
						<MonitorPlay />
						{m["videos.shadowing.toolbar.watch"]()}
					</ToggleGroupItem>
					<ToggleGroupItem value="practice" className={VIEW_TOGGLE_ITEM}>
						<Mic />
						{m["videos.shadowing.toolbar.practice"]()}
					</ToggleGroupItem>
				</ToggleGroup>
			</div>

			<div
				className={cn(
					"grid items-start gap-3",
					twoColumns && "xl:grid-cols-[minmax(0,8fr)_minmax(0,4fr)]",
				)}
			>
				<ShadowingPanel
					video={video}
					youtubeId={youtubeId}
					playerRef={playerRef}
					activeSegment={
						line
							? {
									start: line.startTimestamp,
									end: line.endTimestamp,
								}
							: null
					}
					onPlayerReady={handlePlayerReady}
					onTimeUpdate={handleTimeUpdate}
					onPlayPastEnd={handleNextSentence}
					practice={view === "practice"}
					words={words}
					autoPause={autoPause}
					onAutoPauseChange={setAutoPause}
					largeVideo={largeVideo}
					onLargeVideoChange={setLargeVideo}
					speed={speed}
					onSpeed={handleSpeedCycle}
					onSettings={() => setToolsOpen("settings")}
					onShortcuts={() => setToolsOpen("shortcuts")}
					onStart={handlePlaySentence}
					recording={recording}
					hasRecording={hasRecording}
					onRecord={handleRecord}
					onPlayRecording={handlePlayRecording}
				/>
				{showTranscript ? (
					<TranscriptPanel
						transcripts={transcripts}
						activeIndex={activeIndex}
						progress={
							transcripts.length
								? (completedCount / transcripts.length) * 100
								: 0
						}
						masked={false}
						showPhonetic={showPhonetic}
						showTranslation={showTranslation}
						onSelect={selectLine}
						actions={
							<>
								<Toggle
									size="sm"
									pressed={showPhonetic}
									onPressedChange={setShowPhonetic}
									className={LINE_TOGGLE}
								>
									{showPhonetic ? <Eye /> : <EyeOff />}
									{m["videos.shadowing.transcript.phonetic"]()}
								</Toggle>
								<Toggle
									size="sm"
									pressed={showTranslation}
									onPressedChange={setShowTranslation}
									className={LINE_TOGGLE}
								>
									{showTranslation ? <Eye /> : <EyeOff />}
									{m["videos.shadowing.transcript.translation"]()}
								</Toggle>
							</>
						}
					/>
				) : null}
			</div>
			{pronounce.data && line ? (
				<PronunciationPanel
					result={pronounce.data}
					referenceText={line.content}
					selfAudioUrl={recordingUrl}
				/>
			) : null}
			{pronounce.isError ? (
				<div className="flex flex-wrap items-center justify-end gap-2">
					<span role="alert">{m["videos.shadowing.record.scoreFailed"]()}</span>
				</div>
			) : null}
			{recorderState.status === "error" ? (
				<div className="flex flex-wrap items-center justify-end gap-2">
					<span role="alert">{recorderState.message}</span>
				</div>
			) : null}
			{saveFailed ? (
				<div className="flex flex-wrap items-center justify-end gap-2">
					<span role="alert">{m["videos.shadowing.attempt.saveFailed"]()}</span>
					<SubmitButton pending={isSaving} onClick={retrySave}>
						{m["common.actions.retry"]()}
					</SubmitButton>
				</div>
			) : null}
			<output className="block text-right text-sm text-muted-foreground">
				{isSaving ? m["videos.shadowing.attempt.saving"]() : null}
			</output>
		</PageLayout>
	);
}
