import { createFileRoute, notFound } from "@tanstack/react-router";
import { cn } from "cn";
import { Eye, EyeOff } from "lucide-react";
import { useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { type CrumbItem, useBreadcrumbs } from "#/app/breadcrumbs";
import { CefrBadge } from "#/components/common/cefr-badge";
import { PageLayout } from "#/components/common/page-layout";
import { Button } from "#/components/ui/button";
import { DictationPanel } from "#/features/videos/components/dictation/dictation-panel";
import { VideoPanel } from "#/features/videos/components/dictation/video-panel";
import { TranscriptPanel } from "#/features/videos/components/transcript-panel";
import { type DictationLevel, splitWords } from "#/features/videos/dictation";
import { getVideoById } from "#/features/videos/fixtures";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/videos/$videoId/dictation")({
	loader: ({ params: { videoId } }) => {
		const video = getVideoById(videoId);
		if (!video) throw notFound();
		return video;
	},
	component: DictationPage,
});

/** Column template per visible-pane combination (media, transcript). */
const GRID = {
	both: "xl:grid-cols-[minmax(0,6fr)_minmax(0,5fr)_minmax(0,3.5fr)]",
	media: "xl:grid-cols-[minmax(0,7fr)_minmax(0,5fr)]",
	transcript: "xl:grid-cols-[minmax(0,7fr)_minmax(0,5fr)]",
	none: "mx-auto w-full max-w-3xl",
} as const;

function DictationPage() {
	const video = Route.useLoaderData();
	const transcripts = video.transcripts ?? [];

	// UI-only state: playback, checking and scoring are wired by the caller.
	const [showMedia, setShowMedia] = useState(true);
	const [showTranscript, setShowTranscript] = useState(true);
	const [level, setLevel] = useState<DictationLevel>("hard");
	const [activeIndex, setActiveIndex] = useState(0);
	const [answer, setAnswer] = useState("");
	const [revealed, setRevealed] = useState<number[]>([]);

	const crumbs: CrumbItem[] = [
		{ label: m["nav.item.videos"](), to: APP_ROUTES.VIDEOS.LIST },
		...(video.category ? [{ label: video.category.name }] : []),
		{ label: video.title },
	];
	useBreadcrumbs(crumbs);

	const words = splitWords(transcripts[activeIndex]?.content ?? "");
	const selectLine = (index: number) => {
		setActiveIndex(index);
		setAnswer("");
		setRevealed([]);
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
			<div className="flex flex-wrap items-center justify-end gap-3">
				{video.cefrLevel ? (
					<span className="chip mr-auto px-3 py-1">
						<CefrBadge value={video.cefrLevel} />
					</span>
				) : null}
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
				{showMedia ? <VideoPanel video={video} currentTime={0} /> : null}
				<DictationPanel
					level={level}
					onLevelChange={setLevel}
					speed={1}
					words={words}
					revealed={revealed}
					onReveal={(index) =>
						setRevealed((current) =>
							current.includes(index)
								? current.filter((item) => item !== index)
								: [...current, index],
						)
					}
					onRevealAll={() => setRevealed(words.map((_, index) => index))}
					answer={answer}
					onAnswerChange={setAnswer}
					onPrev={() => selectLine(Math.max(0, activeIndex - 1))}
					onNextSentence={() =>
						selectLine(Math.min(transcripts.length - 1, activeIndex + 1))
					}
					onNext={() =>
						selectLine(Math.min(transcripts.length - 1, activeIndex + 1))
					}
				/>
				{showTranscript ? (
					<TranscriptPanel
						transcripts={transcripts}
						activeIndex={activeIndex}
						progress={0}
						onSelect={selectLine}
					/>
				) : null}
			</div>
		</PageLayout>
	);
}
