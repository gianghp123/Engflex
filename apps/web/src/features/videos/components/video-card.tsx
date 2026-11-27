import { Link, useNavigate } from "@tanstack/react-router";
import { cn } from "cn";
import { Lock } from "lucide-react";
import { useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { useCanAccessPremium } from "#/features/payments/queries";
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
	const navigate = useNavigate();
	const canAccessPremium = useCanAccessPremium();
	const premiumBlocked = video.isPremium && !canAccessPremium;
	const handleSelectMode = (mode: PracticeMode) => {
		setOpen(false);
		if (onSelectMode) {
			onSelectMode(mode);
			return;
		}
		if (mode === "dictation") {
			void navigate({
				to: APP_ROUTES.VIDEOS.DICTATION,
				params: { videoId: video.id },
			});
		} else if (mode === "shadowing") {
			void navigate({
				to: APP_ROUTES.VIDEOS.SHADOWING,
				params: { videoId: video.id },
			});
		}
	};
	return (
		<>
			<article
				className={cn(
					"group surface-card-interactive relative flex h-full cursor-pointer flex-col overflow-hidden",
					video.isPremium &&
						"bg-linear-to-b from-card from-55% to-pro-tint [--card-edge-hover:var(--pro)] [--card-edge:var(--pro-border)]",
				)}
			>
				<VideoThumbnail video={video} />
				<div className="flex flex-1 flex-col gap-3 px-4 pt-3 pb-4">
					<h3
						title={video.title}
						className={cn(
							"line-clamp-2 min-h-[42px] text-[15px] leading-[21px] font-bold",
							video.isPremium ? "text-pro" : "text-foreground",
						)}
					>
						{video.title}
					</h3>
					<div className="mt-auto flex items-center justify-between gap-2">
						{premiumBlocked ? (
							<span className="inline-flex items-center gap-1 text-xs font-bold text-muted-foreground">
								<Lock className="size-3.5" />
								{m["common.premium.upgradeAction"]()}
							</span>
						) : (
							<>
								<PracticeChip
									label={m["videos.card.practice.dictation"]()}
									done={video.practice?.dictation}
								/>
								<PracticeChip
									label={m["videos.card.practice.shadowing"]()}
									done={video.practice?.shadowing}
								/>
							</>
						)}
					</div>
				</div>
				{premiumBlocked ? (
					<Link
						to={APP_ROUTES.PAYMENTS.PRICING}
						aria-label={video.title}
						className="absolute inset-0 cursor-pointer rounded-[22px] focus-visible:outline-3 focus-visible:outline-offset-3 focus-visible:outline-ring"
					/>
				) : (
					<button
						type="button"
						aria-label={video.title}
						aria-haspopup="dialog"
						onClick={() => setOpen(true)}
						className="absolute inset-0 cursor-pointer rounded-[22px] focus-visible:outline-3 focus-visible:outline-offset-3 focus-visible:outline-ring"
					/>
				)}
			</article>
			<PracticeModeDialog
				open={open}
				onOpenChange={setOpen}
				onSelect={handleSelectMode}
			/>
		</>
	);
}
