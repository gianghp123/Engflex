import { cn } from "cn";
import { Pause, Play, RotateCcw } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { Button } from "#/components/ui/button";
import { m } from "#/paraglide/messages";

const BARS = 28;

function formatSeconds(totalSeconds: number | null): string {
	if (totalSeconds === null || !Number.isFinite(totalSeconds)) return "–:––";
	const floored = Math.max(0, Math.floor(totalSeconds));
	return `${Math.floor(floored / 60)}:${String(floored % 60).padStart(2, "0")}`;
}

/**
 * Real audio player backed by `<audio>`. Progress follows `timeupdate`, the
 * total comes from `loadedmetadata` (falling back to `durationMs` when the
 * file exposes no duration). A load error swaps the controls for a message.
 */
export function AudioPlayer({
	src,
	durationMs,
}: {
	src: string;
	durationMs?: number;
}) {
	const ref = useRef<HTMLAudioElement>(null);
	const [playing, setPlaying] = useState(false);
	const [failed, setFailed] = useState(false);
	const [slow, setSlow] = useState(false);
	const [current, setCurrent] = useState(0);
	const [total, setTotal] = useState<number | null>(null);

	useEffect(() => {
		const audio = ref.current;
		if (audio) audio.playbackRate = slow ? 0.8 : 1;
	}, [slow]);

	const fallbackTotal =
		total ?? (durationMs !== undefined ? durationMs / 1000 : null);
	const progress =
		fallbackTotal && fallbackTotal > 0
			? Math.min(1, current / fallbackTotal)
			: 0;
	const filledBars = Math.round(progress * BARS);

	if (failed) {
		return (
			<div className="flex flex-col gap-2 rounded-[16px] bg-accent p-4 shadow-edge-sm">
				<p role="alert" className="text-sm font-medium text-ai-coral">
					{m["lessons.listening.audioFailed"]()}
				</p>
			</div>
		);
	}

	return (
		<div className="flex flex-col gap-2 rounded-[16px] bg-accent p-4 shadow-edge-sm">
			{/* biome-ignore lint/a11y/useMediaCaption: dialogue audio only — the full transcript renders as text in ListeningBriefCard beside this player. */}
			<audio
				ref={ref}
				src={src}
				preload="metadata"
				onTimeUpdate={(event) => setCurrent(event.currentTarget.currentTime)}
				onLoadedMetadata={(event) => setTotal(event.currentTarget.duration)}
				onEnded={() => setPlaying(false)}
				onError={() => {
					setFailed(true);
					setPlaying(false);
				}}
			/>
			<div className="flex items-center gap-2">
				<Button
					type="button"
					aria-label={
						playing ? m["common.audio.pause"]() : m["common.audio.play"]()
					}
					onClick={() => {
						const audio = ref.current;
						if (!audio) return;
						if (playing) {
							audio.pause();
							setPlaying(false);
						} else {
							audio.playbackRate = slow ? 0.8 : 1;
							void audio.play().then(
								() => setPlaying(true),
								() => {
									setFailed(true);
									setPlaying(false);
								},
							);
						}
					}}
					className="btn btn-primary h-12 w-12 rounded-full p-0"
				>
					{playing ? <Pause /> : <Play />}
				</Button>
				<Button
					type="button"
					variant="outline"
					size="sm"
					className="btn btn-outline px-2"
					aria-label={m["lessons.listening.restart"]()}
					title={m["lessons.listening.restart"]()}
					onClick={() => {
						const audio = ref.current;
						if (!audio) return;
						audio.currentTime = 0;
						setCurrent(0);
						audio.playbackRate = slow ? 0.8 : 1;
						void audio.play().then(
							() => setPlaying(true),
							() => {
								setFailed(true);
								setPlaying(false);
							},
						);
					}}
				>
					<RotateCcw data-icon="inline-start" />
					<span className="sr-only">{m["lessons.listening.restart"]()}</span>
				</Button>
				<span className="ml-auto flex items-center gap-1">
					<button
						type="button"
						aria-pressed={!slow}
						onClick={() => setSlow(false)}
						className={cn(
							"chip px-2 py-0 text-[11px]",
							!slow && "chip-selected",
						)}
					>
						1.0x
					</button>
					<button
						type="button"
						aria-pressed={slow}
						onClick={() => setSlow(true)}
						className={cn(
							"chip px-2 py-0 text-[11px]",
							slow && "chip-selected",
						)}
					>
						0.8x
					</button>
				</span>
			</div>
			<div
				className="flex h-12 items-center gap-1 rounded-[16px] bg-card px-3"
				aria-hidden="true"
			>
				{BARS
					? Array.from({ length: BARS }, (_, index) => {
							const height = 8 + ((index * 7) % 20);
							return (
								<span
									// biome-ignore lint/suspicious/noArrayIndexKey: fixed-length decorative waveform — the bars never reorder and each bar's height derives from its own index, so index IS the identity here.
									key={index}
									className={cn(
										"w-1 shrink-0 rounded-full",
										index < filledBars ? "bg-secondary" : "bg-primary",
									)}
									style={{ height, opacity: index < filledBars ? 1 : 0.45 }}
								/>
							);
						})
					: null}
			</div>
			<div className="flex items-center justify-between text-[11px] font-medium text-muted-foreground">
				<span>{formatSeconds(current)}</span>
				<span>
					{m["common.audio.total"]()} {formatSeconds(fallbackTotal)}
				</span>
			</div>
		</div>
	);
}
