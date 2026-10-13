import { createFileRoute, notFound } from "@tanstack/react-router";
import { cn } from "cn";
import { Clock3, Eye, EyeOff, Mic, MonitorPlay } from "lucide-react";
import { useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { type CrumbItem, useBreadcrumbs } from "#/app/breadcrumbs";
import { CefrBadge } from "#/components/common/cefr-badge";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { Toggle } from "#/components/ui/toggle";
import { ToggleGroup, ToggleGroupItem } from "#/components/ui/toggle-group";
import { ShadowingPanel } from "#/features/videos/components/shadowing/shadowing-panel";
import { TranscriptPanel } from "#/features/videos/components/transcript-panel";
import { pairWords } from "#/features/videos/dictation";
import { getVideoById } from "#/features/videos/fixtures";
import { formatDuration } from "#/features/videos/format";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/videos/$videoId/shadowing")({
	loader: ({ params: { videoId } }) => {
		const video = getVideoById(videoId);
		if (!video) throw notFound();
		return video;
	},
	component: ShadowingPage,
});

type View = "watch" | "practice";

const VIEW_TOGGLE_ITEM = cn(
	"rounded-full! px-3 text-xs font-extrabold tracking-wide uppercase",
	"data-[state=on]:bg-sidebar data-[state=on]:text-sidebar-foreground",
);

const LINE_TOGGLE =
	"h-8 gap-1.5 px-2 text-xs font-semibold text-muted-foreground data-[state=on]:bg-transparent data-[state=on]:text-foreground";

function ShadowingPage() {
	const video = Route.useLoaderData();
	const transcripts = video.transcripts ?? [];

	// UI-only state: playback, recording and scoring are wired by the caller.
	const [view, setView] = useState<View>("practice");
	const [showTranscript, setShowTranscript] = useState(true);
	const [showPhonetic, setShowPhonetic] = useState(false);
	const [showTranslation, setShowTranslation] = useState(true);
	const [autoPause, setAutoPause] = useState(true);
	const [largeVideo, setLargeVideo] = useState(false);
	const [activeIndex, setActiveIndex] = useState(0);

	const crumbs: CrumbItem[] = [
		{ label: m["nav.item.videos"](), to: APP_ROUTES.VIDEOS.LIST },
		...(video.category ? [{ label: video.category.name }] : []),
		{ label: video.title },
	];
	useBreadcrumbs(crumbs);

	const line = transcripts[activeIndex];
	const words = line ? pairWords(line.content, line.phonetic) : [];
	const time = formatDuration(0);
	const twoColumns = showTranscript && !largeVideo;

	return (
		<PageLayout>
			<div className="flex flex-wrap items-center justify-end gap-3">
				{video.cefrLevel ? (
					<span className="chip mr-auto px-3 py-1">
						<CefrBadge value={video.cefrLevel} />
					</span>
				) : null}
				<span className="flex items-center gap-2 text-sm font-semibold text-muted-foreground">
					<MoMascot variant="headphones" size={28} />
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
						if (value) setView(value as View);
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
					practice={view === "practice"}
					words={words}
					autoPause={autoPause}
					onAutoPauseChange={setAutoPause}
					largeVideo={largeVideo}
					onLargeVideoChange={setLargeVideo}
					speed={1}
					recording={false}
					hasRecording={false}
				/>
				{showTranscript ? (
					<TranscriptPanel
						transcripts={transcripts}
						activeIndex={activeIndex}
						progress={0}
						masked={false}
						showPhonetic={showPhonetic}
						showTranslation={showTranslation}
						onSelect={setActiveIndex}
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
		</PageLayout>
	);
}
