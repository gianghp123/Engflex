import { useCallback, useRef, useState } from "react";
import { m } from "#/paraglide/messages";

export type RecorderState =
	| { status: "idle" }
	| { status: "recording" }
	| { status: "ready"; blob: Blob; mime: string }
	| { status: "error"; message: string };

const SUPPORTED_MIME = ["audio/webm", "audio/mp4"] as const;

function pickMime(): string | null {
	if (typeof MediaRecorder === "undefined") return null;
	for (const mime of SUPPORTED_MIME) {
		if (MediaRecorder.isTypeSupported(mime)) return mime;
	}
	return null;
}

/**
 * Minimal MediaRecorder wrapper for read-aloud/shadowing capture.
 *
 * `start` resolves once recording has begun (or an `error` state is set);
 * `stop` resolves the recorded chunks into a `ready` blob asynchronously via
 * `onstop`. Microphone denial or a missing `MediaRecorder` lands in `error`
 * with `micDeniedMessage` — this hook never throws.
 */
export function useRecorder(
	micDeniedMessage: string = m["lessons.speaking.micDenied"](),
): {
	state: RecorderState;
	start: () => Promise<void>;
	stop: () => void;
	reset: () => void;
} {
	const [state, setState] = useState<RecorderState>({ status: "idle" });
	const recorder = useRef<MediaRecorder | null>(null);
	const chunks = useRef<Blob[]>([]);
	const mime = useRef<string>("");

	const start = useCallback(async () => {
		const picked = pickMime();
		if (!picked || typeof navigator === "undefined") {
			setState({ status: "error", message: micDeniedMessage });
			return;
		}
		try {
			const stream = await navigator.mediaDevices.getUserMedia({
				audio: true,
			});
			chunks.current = [];
			mime.current = picked;
			const rec = new MediaRecorder(stream, { mimeType: picked });
			rec.ondataavailable = (event) => {
				if (event.data.size > 0) chunks.current.push(event.data);
			};
			recorder.current = rec;
			rec.start();
			setState({ status: "recording" });
		} catch {
			setState({ status: "error", message: micDeniedMessage });
		}
	}, [micDeniedMessage]);

	const stop = useCallback(() => {
		const rec = recorder.current;
		if (!rec || rec.state === "inactive") return;
		const type = mime.current;
		rec.onstop = () => {
			const blob = new Blob(chunks.current, { type });
			rec.stream.getTracks().forEach((track) => {
				track.stop();
			});
			recorder.current = null;
			setState({ status: "ready", blob, mime: type });
		};
		rec.stop();
	}, []);

	const reset = useCallback(() => {
		const rec = recorder.current;
		if (rec) {
			// Detach callbacks first so a stale onstop can never fire after
			// reset and resurrect a ready blob for the next item.
			rec.onstop = null;
			rec.ondataavailable = null;
			rec.stream.getTracks().forEach((track) => {
				track.stop();
			});
			recorder.current = null;
		}
		chunks.current = [];
		setState({ status: "idle" });
	}, []);

	return { state, start, stop, reset };
}
