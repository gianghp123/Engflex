import { Clock3, Pause, Play, RotateCcw } from "lucide-react";
import { type RefObject, useState } from "react";
import { MoMascot } from "#/components/common/mo-mascot";
import { Button } from "#/components/ui/button";
import { Separator } from "#/components/ui/separator";
import { VideoStage } from "#/features/videos/components/video-stage";
import {
	YouTubePlayer,
	type YouTubePlayerRef,
} from "#/features/videos/components/youtube-player";
import { formatDuration } from "#/features/videos/format";
import type { Video } from "#/features/videos/types";
import { m } from "#/paraglide/messages";

export function VideoPanel({
	video,
	youtubeId,
	playerRef,
	activeSegment,
	onStart,
	onReplay,
	onPlayerReady,
	onSegmentEnd,
	onPlayPastEnd,
	onPlayingChange,
	onRateChange,
	isPlaying = false,
	children,
}: {
	video: Video;
	youtubeId?: string | null;
	playerRef?: RefObject<YouTubePlayerRef | null>;
	activeSegment?: { start: number; end: number } | null;
	onStart?: () => void;
	onReplay?: () => void;
	onPlayerReady?: () => void;
	onSegmentEnd?: () => void;
	onPlayPastEnd?: () => void;
	onPlayingChange?: (playing: boolean) => void;
	onRateChange?: (rate: number) => void;
	isPlaying?: boolean;
	children?: React.ReactNode;
}) {
	const [currentTime, setCurrentTime] = useState(0);
	const time = formatDuration(currentTime);

	return (
		<section className="surface-card flex flex-col gap-3 p-3">
			<div className="flex items-center justify-between gap-2 px-1">
				<h2 className="text-base font-extrabold tracking-wide text-foreground uppercase">
					{m["videos.dictation.video.title"]()}
				</h2>
				<span
					className="chip px-3 py-1"
					title={m["videos.dictation.video.time"]({ time })}
				>
					<Clock3 aria-hidden="true" className="size-4" />
					<span aria-hidden="true">{time}</span>
					<span className="sr-only">
						{m["videos.dictation.video.time"]({ time })}
					</span>
				</span>
			</div>

			<VideoStage video={video} onPlay={onStart}>
				{children ??
					(youtubeId ? (
						<YouTubePlayer
							ref={playerRef}
							videoId={youtubeId}
							activeSegment={activeSegment}
							onTimeUpdate={setCurrentTime}
							onReady={onPlayerReady}
							onSegmentEnd={onSegmentEnd}
							onPlayPastEnd={onPlayPastEnd}
							onPlayingChange={onPlayingChange}
							onRateChange={onRateChange}
						/>
					) : null)}
			</VideoStage>

			<p className="text-sm font-extrabold tracking-wide text-muted-foreground uppercase">
				{m["videos.dictation.video.controls"]()}
			</p>
			<div className="grid grid-cols-2 gap-3">
				<Button type="button" onClick={onStart} className="btn btn-primary">
					{isPlaying ? (
						<Pause data-icon="inline-start" />
					) : (
						<Play data-icon="inline-start" />
					)}
					{isPlaying
						? m["videos.dictation.playback.pause"]()
						: m["videos.dictation.video.start"]()}
				</Button>
				<Button
					type="button"
					variant="outline"
					onClick={onReplay}
					className="btn btn-outline"
				>
					<RotateCcw data-icon="inline-start" />
					{m["videos.dictation.video.replay"]()}
				</Button>
			</div>

			<Separator />
			<div className="flex justify-center">
				<MoMascot variant="headphones" size={104} />
			</div>
		</section>
	);
}
