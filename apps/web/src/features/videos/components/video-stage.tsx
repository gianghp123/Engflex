import { cn } from "cn";
import { Play } from "lucide-react";
import { Button } from "#/components/ui/button";
import type { Video } from "#/features/videos/types";
import { m } from "#/paraglide/messages";

/** Poster + play overlay. The real player (iframe/<video>) is the caller's. */
export function VideoStage({
	video,
	onPlay,
	className,
}: {
	video: Video;
	onPlay?: () => void;
	className?: string;
}) {
	return (
		<div
			className={cn(
				"relative aspect-video overflow-hidden rounded-xl border-2 border-primary bg-muted shadow-[0_4px_0_var(--border)]",
				className,
			)}
		>
			{video.thumbnailUrl ? (
				<img
					src={video.thumbnailUrl}
					alt={video.title}
					className="size-full object-cover"
				/>
			) : (
				<div
					aria-hidden="true"
					className="size-full bg-linear-to-br from-primary to-accuracy"
				/>
			)}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute inset-0 bg-black/25"
			/>
			<Button
				type="button"
				onClick={onPlay}
				aria-label={m["videos.practice.video.play"]()}
				className="btn btn-primary absolute top-1/2 left-1/2 size-14 -translate-x-1/2 -translate-y-1/2 rounded-full p-0"
			>
				<Play />
			</Button>
		</div>
	);
}
