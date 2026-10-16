import { Mic, Play, Square } from "lucide-react";
import { Button } from "#/components/ui/button";
import { Label } from "#/components/ui/label";
import { Switch } from "#/components/ui/switch";
import { PlayerTools } from "#/features/videos/components/player-tools";
import { VideoStage } from "#/features/videos/components/video-stage";
import type { Video } from "#/features/videos/types";
import { m } from "#/paraglide/messages";
import { WordIpaRow } from "./word-ipa-row";

/**
 * Main shadowing pane: video, playback switches, the current sentence with
 * IPA and the record controls. Recording and playback are the caller's.
 */
export function ShadowingPanel({
	video,
	practice,
	words,
	autoPause,
	onAutoPauseChange,
	largeVideo,
	onLargeVideoChange,
	speed,
	recording,
	hasRecording,
	onStart,
	onRecord,
	onPlayRecording,
	onSpeed,
	onSettings,
	onShortcuts,
}: {
	video: Video;
	/** false = watch-only view: the practice block is hidden. */
	practice: boolean;
	words: { word: string; ipa?: string }[];
	autoPause: boolean;
	onAutoPauseChange: (value: boolean) => void;
	largeVideo: boolean;
	onLargeVideoChange: (value: boolean) => void;
	speed: number;
	recording: boolean;
	hasRecording: boolean;
	onStart?: () => void;
	onRecord?: () => void;
	onPlayRecording?: () => void;
	onSpeed?: () => void;
	onSettings?: () => void;
	onShortcuts?: () => void;
}) {
	return (
		<section className="surface-card flex flex-col gap-4 p-3">
			<VideoStage video={video} onPlay={onStart} />

			<div className="grid items-center gap-3 sm:grid-cols-[1fr_auto_1fr]">
				<div className="flex flex-col gap-2 px-1">
					<div className="flex items-center gap-2">
						<Switch
							id="shadowing-auto-pause"
							checked={autoPause}
							onCheckedChange={onAutoPauseChange}
						/>
						<Label htmlFor="shadowing-auto-pause" className="font-semibold">
							{m["videos.shadowing.controls.autoPause"]()}
						</Label>
					</div>
					<div className="flex items-center gap-2">
						<Switch
							id="shadowing-large-video"
							checked={largeVideo}
							onCheckedChange={onLargeVideoChange}
						/>
						<Label htmlFor="shadowing-large-video" className="font-semibold">
							{m["videos.shadowing.controls.largeVideo"]()}
						</Label>
					</div>
				</div>
				<Button
					type="button"
					onClick={onStart}
					className="btn btn-primary min-w-48 uppercase"
				>
					<Play data-icon="inline-start" />
					{m["videos.shadowing.controls.start"]()}
				</Button>
				<div className="flex justify-end">
					<PlayerTools
						speed={speed}
						onSpeed={onSpeed}
						onSettings={onSettings}
						onShortcuts={onShortcuts}
					/>
				</div>
			</div>

			{practice ? (
				<>
					<WordIpaRow words={words} />
					<div className="flex flex-wrap items-center justify-center gap-3">
						<Button
							type="button"
							variant="outline"
							onClick={onPlayRecording}
							disabled={!hasRecording}
							className="btn btn-outline min-w-48 uppercase"
						>
							<Play data-icon="inline-start" />
							{m["videos.shadowing.record.playback"]()}
						</Button>
						<Button
							type="button"
							onClick={onRecord}
							aria-pressed={recording}
							className={
								recording
									? "btn btn-danger min-w-48 uppercase"
									: "btn btn-primary min-w-48 uppercase"
							}
						>
							{recording ? (
								<Square data-icon="inline-start" />
							) : (
								<Mic data-icon="inline-start" />
							)}
							{recording
								? m["videos.shadowing.record.stop"]()
								: m["videos.shadowing.record.start"]()}
						</Button>
					</div>
				</>
			) : null}
		</section>
	);
}
