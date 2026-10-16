import { Clock3, Play, RotateCcw } from "lucide-react";
import { MoMascot } from "#/components/common/mo-mascot";
import { Button } from "#/components/ui/button";
import { Separator } from "#/components/ui/separator";
import { VideoStage } from "#/features/videos/components/video-stage";
import { formatDuration } from "#/features/videos/format";
import type { Video } from "#/features/videos/types";
import { m } from "#/paraglide/messages";

/** Left pane: video stage + start/replay controls. Playback is the caller's. */
export function VideoPanel({
	video,
	currentTime,
	onStart,
	onReplay,
	children,
}: {
	video: Video;
	currentTime: number;
	onStart?: () => void;
	onReplay?: () => void;
	children?: React.ReactNode;
}) {
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
				{children}
			</VideoStage>

			<p className="text-sm font-extrabold tracking-wide text-muted-foreground uppercase">
				{m["videos.dictation.video.controls"]()}
			</p>
			<div className="grid grid-cols-2 gap-3">
				<Button type="button" onClick={onStart} className="btn btn-primary">
					<Play data-icon="inline-start" />
					{m["videos.dictation.video.start"]()}
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
