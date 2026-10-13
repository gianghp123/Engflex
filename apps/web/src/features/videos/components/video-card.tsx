import { cn } from "cn";
import { useState } from "react";
import type { Video } from "#/features/videos/types";
import { m } from "#/paraglide/messages";
import { PracticeChip } from "./practice-chip";
import { type PracticeMode, PracticeModeDialog } from "./practice-mode-dialog";
import { VideoThumbnail } from "./video-thumbnail";

export function VideoCard({
	video,
	onSelectMode,
}: {
	video: Video;
	onSelectMode?: (mode: PracticeMode) => void;
}) {
	const [open, setOpen] = useState(false);
	return (
		<>
			<article
				className={cn(
					"group surface-card-interactive relative flex h-full cursor-pointer flex-col overflow-hidden",
					video.isPro &&
						"bg-linear-to-b from-card from-55% to-pro-tint [--card-edge-hover:var(--pro)] [--card-edge:var(--pro-border)]",
				)}
			>
				<VideoThumbnail video={video} />
				<div className="flex flex-1 flex-col gap-3 px-4 pt-3 pb-4">
					<h3
						title={video.title}
						className={cn(
							"line-clamp-2 min-h-[42px] text-[15px] leading-[21px] font-bold",
							video.isPro ? "text-pro" : "text-foreground",
						)}
					>
						{video.title}
					</h3>
					<div className="mt-auto flex items-center justify-between gap-2">
						<PracticeChip
							label={m["videos.card.practice.dictation"]()}
							done={video.practice?.dictation}
						/>
						<PracticeChip
							label={m["videos.card.practice.shadowing"]()}
							done={video.practice?.shadowing}
						/>
					</div>
				</div>
				<button
					type="button"
					aria-label={video.title}
					aria-haspopup="dialog"
					onClick={() => setOpen(true)}
					className="absolute inset-0 cursor-pointer rounded-[22px] focus-visible:outline-3 focus-visible:outline-offset-3 focus-visible:outline-ring"
				/>
			</article>
			<PracticeModeDialog
				open={open}
				onOpenChange={setOpen}
				onSelect={onSelectMode}
			/>
		</>
	);
}
