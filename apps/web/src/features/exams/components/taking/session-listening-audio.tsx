import type { ExamTakingContent } from "@engflex/contracts";
import { type RefObject, useEffect, useRef, useState } from "react";
import { Button } from "#/components/ui/button";
import {
	clearAudioPosition,
	readAudioPosition,
	saveAudioPosition,
} from "#/features/exams/audio-position";
import { mediaSource } from "#/features/exams/review";
import { m } from "#/paraglide/messages";

export function SessionListeningAudio({
	content,
	attemptId,
	userId,
	speed,
	locked,
	audioRef,
	closed,
}: {
	content: ExamTakingContent;
	attemptId: string;
	userId: string | null | undefined;
	speed: number;
	locked: boolean;
	audioRef: RefObject<HTMLAudioElement | null>;
	closed: RefObject<boolean>;
}) {
	const src = mediaSource(content.audioAssetId, content.assets);
	const sha256 = content.assets
		.find((asset) => asset.sourceId === content.audioAssetId)
		?.sha256.toLowerCase();
	const [failed, setFailed] = useState(false);
	const position = useRef<number | null>(null);
	const positionIdentity = useRef<string | null>(null);
	const restored = useRef(false);
	const lastSave = useRef(0);
	useEffect(() => {
		const audio = audioRef.current;
		if (!audio || !src || !sha256 || !userId) return;
		const identity = `${userId}:${attemptId}:${sha256}`;
		if (positionIdentity.current !== identity) {
			positionIdentity.current = identity;
			position.current = null;
			restored.current = false;
			lastSave.current = 0;
		}
		const remember = () => {
			if (!userId) return;
			if (closed.current) {
				clearAudioPosition(userId, attemptId);
				return;
			}
			if (
				!restored.current ||
				audio.readyState < 1 ||
				!Number.isFinite(audio.currentTime)
			)
				return;
			position.current = audio.currentTime;
			saveAudioPosition(userId, attemptId, sha256, audio.currentTime);
			lastSave.current = Date.now();
		};
		const restore = () => {
			if (!Number.isFinite(audio.duration) || audio.duration <= 0) {
				setFailed(true);
				return;
			}
			if (!restored.current) {
				const saved =
					position.current ??
					(userId ? readAudioPosition(userId, attemptId, sha256) : null);
				audio.currentTime =
					saved !== null && saved < audio.duration ? saved : 0;
				position.current = audio.currentTime;
				restored.current = true;
			}
			setFailed(false);
		};
		const update = () => {
			if (restored.current && audio.readyState >= 1)
				position.current = audio.currentTime;
			if (Date.now() - lastSave.current >= 2000) remember();
		};
		audio.addEventListener("loadedmetadata", restore);
		audio.addEventListener("timeupdate", update);
		audio.addEventListener("pause", remember);
		window.addEventListener("pagehide", remember);
		if (audio.error) setFailed(true);
		else if (audio.readyState >= 1) restore();
		return () => {
			audio.pause();
			remember();
			audio.removeEventListener("loadedmetadata", restore);
			audio.removeEventListener("timeupdate", update);
			audio.removeEventListener("pause", remember);
			window.removeEventListener("pagehide", remember);
		};
	}, [attemptId, userId, src, sha256, audioRef, closed]);
	useEffect(() => {
		const audio = audioRef.current;
		if (!audio) return;
		audio.defaultPlaybackRate = speed;
		audio.playbackRate = speed;
		if (locked) audio.pause();
	}, [audioRef, speed, locked]);
	const unavailable = !src || !sha256 || failed;
	return (
		<section
			className="shrink-0 border-b px-4 py-2"
			aria-label={m["exams.toeic.take.fullListening"]()}
		>
			{/* biome-ignore lint/a11y/useMediaCaption: taking content intentionally excludes transcripts. */}
			<audio
				ref={audioRef}
				src={src ?? undefined}
				hidden={unavailable}
				controls={!locked}
				preload="metadata"
				className="h-10 w-full"
				aria-label={m["exams.toeic.take.fullListening"]()}
				onPlay={(event) => {
					if (locked || closed.current || !userId) event.currentTarget.pause();
				}}
				onError={() => {
					audioRef.current?.pause();
					setFailed(true);
				}}
			/>
			{unavailable && (
				<div className="flex items-center gap-3">
					<p role="alert" className="text-sm">
						{m["exams.toeic.take.mediaFailed"]()}
					</p>
					<Button
						variant="outline"
						size="sm"
						disabled={locked}
						onClick={() => {
							restored.current = false;
							setFailed(false);
							audioRef.current?.load();
						}}
					>
						{m["exams.toeic.take.retry"]()}
					</Button>
				</div>
			)}
		</section>
	);
}
