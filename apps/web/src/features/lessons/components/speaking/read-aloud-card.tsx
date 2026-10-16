import type { Activity, PronunciationResult } from "@engflex/contracts";
import { Mic } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { MoMascot } from "#/components/common/mo-mascot";
import { PageSplit } from "#/components/common/page-layout";
import { SubmitButton } from "#/components/common/submit-button";
import { Button } from "#/components/ui/button";
import { useRecorder } from "#/features/lessons/hooks/use-recorder";
import { usePronounceAttempt } from "#/features/lessons/queries";
import { m } from "#/paraglide/messages";
import { PronunciationPanel } from "./pronunciation-panel";

/**
 * Model sentence as a tappable Mo row: the whole row toggles the model
 * audio (idle `headphones`, playing `talk`). Keyed by `audioUrl` at the
 * call site so advancing items resets the playing state and stops audio.
 */
function ModelRow({ text, audioUrl }: { text: string; audioUrl: string }) {
	const ref = useRef<HTMLAudioElement | null>(null);
	const [playing, setPlaying] = useState(false);
	return (
		<div className="flex flex-col gap-1">
			{/* biome-ignore lint/a11y/useMediaCaption: model utterance only — the exact sentence renders as text in the bubble beside this player. */}
			<audio
				ref={ref}
				src={audioUrl}
				preload="metadata"
				className="hidden"
				onEnded={() => setPlaying(false)}
			/>
			<button
				type="button"
				onClick={() => {
					const audio = ref.current;
					if (!audio) return;
					if (playing) {
						audio.pause();
						setPlaying(false);
					} else {
						audio.currentTime = 0;
						void audio.play().then(
							() => setPlaying(true),
							() => setPlaying(false),
						);
					}
				}}
				aria-label={m["lessons.speaking.listenModel"]()}
				className="flex w-full cursor-pointer items-center gap-3 rounded-2xl text-left focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
			>
				<MoMascot variant={playing ? "talk" : "headphones"} size={44} />
				<span className="bubble bubble-partner min-w-0 flex-1 text-[15px]">
					“{text}”
				</span>
			</button>
		</div>
	);
}

/**
 * Read-aloud activity: one item at a time — listen to the model, record,
 * upload once, review the score, then retry the same item or advance.
 * The final item's Next calls `onComplete`.
 */
export function SpeakingActivity({
	activity,
	attemptId,
	onComplete,
}: {
	activity: Activity;
	attemptId: string | null;
	onComplete?: () => void;
}) {
	const items = activity.speaking?.items ?? [];
	const [index, setIndex] = useState(0);
	const item = items[index];
	const { state, start, stop, reset } = useRecorder();
	const attempt = usePronounceAttempt();
	const [result, setResult] = useState<PronunciationResult | null>(null);
	// One mutation per attempt: the ready effect below fires once per blob.
	const submittedRef = useRef(false);

	const recording = state.status === "recording";
	const scored = result !== null;
	// Pre-score upload guard only: the ready blob is mid-flight (or about to
	// be). This never gates the scored screen — Retry/Next key off
	// `isPending` alone, so success always leaves them enabled. The Record
	// button binds `attempt.isPending` via `SubmitButton pending` and keeps
	// the ready/null guards as plain `disabled`.
	const isLast = index >= items.length - 1;
	const readyBlob = state.status === "ready" ? state.blob : null;
	const [selfAudioUrl, setSelfAudioUrl] = useState<string | null>(null);

	useEffect(() => {
		if (state.status !== "ready" || submittedRef.current) return;
		if (!attemptId) return;
		submittedRef.current = true;
		attempt.mutate(
			{
				activityId: activity.id,
				attemptId,
				itemIndex: index,
				audio: state.blob,
				mime: state.mime,
			},
			{
				onSuccess: (data) => setResult(data),
				onError: () => {
					// Back to idle so Record is reachable again; the
					// `uploadFailed` alert (attempt.isError) stays visible
					// until the next attempt. Without this the item bricks:
					// Record disabled, Retry hidden, no recovery control.
					submittedRef.current = false;
					reset();
				},
			},
		);
	}, [state, activity.id, index, attemptId, attempt.mutate, reset]);

	// Object URL for self-playback: created once per scored blob, revoked on
	// cleanup (retry/next/item change) so no URL outlives its recording.
	useEffect(() => {
		if (readyBlob === null || result === null) return;
		const url = URL.createObjectURL(readyBlob);
		setSelfAudioUrl(url);
		return () => {
			URL.revokeObjectURL(url);
			setSelfAudioUrl(null);
		};
	}, [readyBlob, result]);

	if (!item) return null;

	const retry = () => {
		submittedRef.current = false;
		attempt.reset();
		setResult(null);
		reset();
	};

	const next = () => {
		if (isLast) {
			onComplete?.();
			return;
		}
		submittedRef.current = false;
		attempt.reset();
		setResult(null);
		reset();
		setIndex((current) => current + 1);
	};

	return (
		<PageSplit
			main={
				<div className="surface-card flex flex-col gap-4 p-5">
					<div className="flex flex-col gap-1">
						<h2 className="text-lg font-bold text-foreground">
							{m["lessons.speaking.title"]()}
						</h2>
						<p className="text-sm text-muted-foreground">
							{m["lessons.speaking.instruction"]()}
						</p>
					</div>
					<ModelRow
						key={item.modelAudioUrl}
						text={item.text}
						audioUrl={item.modelAudioUrl}
					/>
					<div className="flex flex-wrap items-center gap-2">
						{!recording && !scored ? (
							<SubmitButton
								type="button"
								size="sm"
								className="btn btn-primary"
								disabled={state.status === "ready" || attemptId === null}
								pending={attempt.isPending}
								onClick={() => void start()}
							>
								{m["lessons.speaking.record"]()}
							</SubmitButton>
						) : null}
						{recording ? (
							<Button
								type="button"
								size="sm"
								className="btn btn-primary"
								disabled={attempt.isPending}
								onClick={stop}
							>
								{m["lessons.speaking.stop"]()}
							</Button>
						) : null}
						{scored ? (
							<>
								<Button
									type="button"
									variant="outline"
									size="sm"
									className="btn btn-outline"
									disabled={attempt.isPending}
									onClick={retry}
								>
									{m["lessons.speaking.retry"]()}
								</Button>
								<Button
									type="button"
									size="sm"
									className="btn btn-primary"
									disabled={attempt.isPending}
									onClick={next}
								>
									{m["lessons.speaking.next"]()}
								</Button>
							</>
						) : null}
						<span className="ml-auto text-xs font-medium text-muted-foreground">
							{index + 1}/{items.length}
						</span>
					</div>
					{state.status === "error" ? (
						<p role="alert" className="text-sm font-medium text-ai-coral">
							{state.message}
						</p>
					) : null}
					{attempt.isError ? (
						<p role="alert" className="text-sm font-medium text-ai-coral">
							{m["lessons.speaking.uploadFailed"]()}
						</p>
					) : null}
				</div>
			}
			aside={
				result ? (
					<PronunciationPanel
						result={result}
						referenceText={item.text}
						selfAudioUrl={selfAudioUrl}
					/>
				) : (
					<div className="surface-card flex min-h-40 flex-col items-center justify-center gap-2 p-5 text-center">
						<Mic aria-hidden="true" className="size-6 text-muted-foreground" />
						<p className="text-sm font-medium text-muted-foreground">
							{m["lessons.speaking.scoreEmpty"]()}
						</p>
					</div>
				)
			}
		/>
	);
}
