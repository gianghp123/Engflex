import { cn } from "cn";
import {
	Clapperboard,
	Clock3,
	Crown,
	Headphones,
	type LucideIcon,
	Youtube,
} from "lucide-react";
import { useState } from "react";
import { CefrBadge } from "#/components/common/cefr-badge";
import { formatDuration, formatViews } from "#/features/videos/format";
import type { Video } from "#/features/videos/types";
import { m } from "#/paraglide/messages";

function OverlayChip({
	icon: Icon,
	text,
	label,
	className,
}: {
	icon: LucideIcon;
	text: string;
	label: string;
	className?: string;
}) {
	return (
		<span
			title={label}
			className={cn(
				"inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs leading-4 font-bold",
				className,
			)}
		>
			<Icon className="size-3.5" aria-hidden="true" />
			<span aria-hidden="true">{text}</span>
			<span className="sr-only">{label}</span>
		</span>
	);
}

const GLASS = "bg-black/55 text-white backdrop-blur-sm";

export function VideoThumbnail({ video }: { video: Video }) {
	const src = video.thumbnailUrl;
	const [failedSrc, setFailedSrc] = useState<string | null>(null);
	const broken = src === "" || failedSrc === src;

	return (
		<div className="relative aspect-video overflow-hidden bg-muted">
			{broken ? (
				<div
					aria-hidden="true"
					className="flex size-full items-center justify-center bg-linear-to-br from-primary to-accuracy"
				>
					<Clapperboard className="size-10 text-white/85" />
				</div>
			) : (
				<img
					src={src}
					alt={video.title}
					loading="lazy"
					decoding="async"
					ref={(el) => {
						if (el?.complete && el.naturalWidth === 0) setFailedSrc(src);
					}}
					onError={() => setFailedSrc(src)}
					className="size-full object-cover transition-transform duration-300 ease-out group-hover:scale-[1.06] motion-reduce:transition-none motion-reduce:group-hover:scale-100"
				/>
			)}
			<div
				aria-hidden="true"
				className="pointer-events-none absolute inset-0 bg-linear-to-b from-black/30 via-transparent via-45% to-black/35"
			/>

			<div className="absolute inset-x-2.5 top-2.5 flex items-start justify-between gap-2">
				<div className="flex flex-wrap items-center gap-1.5">
					{video.isPro ? (
						<OverlayChip
							icon={Crown}
							text={m["videos.card.pro"]()}
							label={m["videos.card.pro"]()}
							className="bg-secondary text-secondary-foreground shadow-[0_2px_0_rgba(0,0,0,0.25)]"
						/>
					) : null}
					{video.views !== undefined ? (
						<OverlayChip
							icon={Headphones}
							text={formatViews(video.views)}
							label={m["videos.card.views"]({
								count: formatViews(video.views),
							})}
							className={GLASS}
						/>
					) : null}
				</div>
				{video.cefrLevel ? (
					<span className="shrink-0 rounded-full bg-card">
						<CefrBadge value={video.cefrLevel} />
					</span>
				) : null}
			</div>

			<div className="absolute inset-x-2.5 bottom-2.5 flex items-end justify-between gap-2">
				{video.source === "youtube" ? (
					<OverlayChip
						icon={Youtube}
						text={m["videos.card.source.youtube"]()}
						label={m["videos.card.source.youtube"]()}
						className="bg-white/95 text-red-600"
					/>
				) : (
					<span />
				)}
				{video.duration > 0 ? (
					<OverlayChip
						icon={Clock3}
						text={formatDuration(video.duration)}
						label={m["videos.card.duration"]({
							time: formatDuration(video.duration),
						})}
						className={GLASS}
					/>
				) : null}
			</div>
		</div>
	);
}
