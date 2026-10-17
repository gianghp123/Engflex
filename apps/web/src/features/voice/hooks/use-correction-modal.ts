import type { TranscriptResult } from "@engflex/contracts";
import { usePipecatClientMicControl } from "@pipecat-ai/client-react";
import { useCallback, useRef, useState } from "react";
import {
	type TranscriptAction,
	transcribeAudio,
	transcriptCommand,
} from "#/features/voice/service";
import { ApiError } from "#/lib/api/api";
import { m } from "#/paraglide/messages";

/**
 * The correction modal.
 *
 * The learner opens it; nothing opens it for them. From the moment it opens
 * the WebRTC mic track is muted, so the live pipeline hears silence and the
 * modal is the only thing they can act on until they send or dismiss.
 *
 * `text` is the field, and the field is the one true copy. It is seeded by
 * `review`, replaced whenever a recorded re-speak transcribes, and editable
 * at all times. The engine keeps no copy of its own.
 */
export type CorrectionWindow = {
	/** The field: the transcript under review, typed or spoken. */
	text: string;
	/** True while a re-speak is being recorded. */
	recording: boolean;
	/** True while a recording is being transcribed. */
	uploading: boolean;
	submitting: boolean;
	error: string | null;
};

const SUPPORTED_MIME = ["audio/webm", "audio/mp4"] as const;

function pickMime(): { mime: string; filename: string } | null {
	if (typeof MediaRecorder === "undefined") return null;
	for (const mime of SUPPORTED_MIME) {
		if (MediaRecorder.isTypeSupported(mime)) {
			return {
				mime,
				filename: mime === "audio/mp4" ? "retake.mp4" : "retake.webm",
			};
		}
	}
	return null;
}

/** E2E seam, read once per recording: the Playwright fake mic cannot feed
 * MediaRecorder, so the test injects fixture bytes and the upload path stays
 * identical. A module function rather than an inline read, so the recording
 * callback has no hidden dependency on the global. */
function readE2eFixture(): ArrayBuffer | undefined {
	if (typeof window === "undefined") return undefined;
	return (window as unknown as { __E2E_RETAKE_BYTES__?: ArrayBuffer })
		.__E2E_RETAKE_BYTES__;
}

export function useCorrectionModal(
	conversationId: string | null,
	options: { onCorrected?: (text: string) => void } = {},
) {
	const { onCorrected } = options;
	const { enableMic } = usePipecatClientMicControl();
	const [window, setWindow] = useState<CorrectionWindow | null>(null);
	const recorder = useRef<MediaRecorder | null>(null);
	const chunks = useRef<Blob[]>([]);
	// What stopRecording uploads: the real recording's container, or the
	// fixture's honest one on the e2e path (bytes and label must agree, or
	// Deepgram rejects the audio).
	const pending = useRef<{ mime: string; filename: string } | null>(null);

	const restoreMic = useCallback(() => {
		recorder.current?.stream.getTracks().forEach((t) => {
			t.stop();
		});
		recorder.current = null;
		enableMic(true);
	}, [enableMic]);

	const close = useCallback(() => {
		restoreMic();
		setWindow(null);
	}, [restoreMic]);

	// The engine is the only source of truth for the reviewed turn, so open
	// asks it and takes its answer. A 404 means the session ended underneath
	// us, which closes the modal rather than stranding the learner. The mic
	// is muted only once the modal will actually show.
	const open = useCallback(async () => {
		if (!conversationId) return;
		setWindow({
			text: "",
			recording: false,
			uploading: false,
			submitting: true,
			error: null,
		});
		try {
			const result = await transcriptCommand(conversationId, "review");
			enableMic(false);
			setWindow({
				text: result.text ?? "",
				recording: false,
				uploading: false,
				submitting: false,
				error: null,
			});
		} catch {
			setWindow(null);
		}
	}, [conversationId, enableMic]);

	const run = useCallback(
		async (action: TranscriptAction, text?: string) => {
			if (!conversationId) return;
			setWindow((w) => (w ? { ...w, submitting: true, error: null } : w));
			try {
				const result: TranscriptResult = await transcriptCommand(
					conversationId,
					action,
					text,
				);
				if (action === "send") {
					onCorrected?.(result.text ?? text ?? "");
				}
				close();
			} catch (err) {
				if (err instanceof ApiError && err.status === 404) {
					close();
					return;
				}
				setWindow((w) =>
					w ? { ...w, submitting: false, error: messageOf(err) } : w,
				);
			}
		},
		[conversationId, close, onCorrected],
	);

	// A recorded re-speak replaces the field — the learner re-spoke precisely
	// because they did not want what was there. On any failure the field is
	// untouched and the error is shown; nothing auto-sends.
	const stopRecording = useCallback(async () => {
		// No recorder and no chunks means nothing to send. The e2e seam fills
		// chunks without a recorder, so chunks — not the recorder — decide.
		if (
			(recorder.current === null && chunks.current.length === 0) ||
			!conversationId
		)
			return;
		const picked = pending.current ?? pickMime();
		setWindow((w) =>
			w ? { ...w, recording: false, uploading: true, error: null } : w,
		);
		let blob: Blob;
		const rec = recorder.current;
		if (rec) {
			blob = await new Promise<Blob>((resolve) => {
				rec.onstop = () =>
					resolve(new Blob(chunks.current, { type: picked?.mime ?? "" }));
				rec.stop();
			});
			recorder.current = null;
		} else {
			blob = new Blob(chunks.current, { type: picked?.mime ?? "audio/wav" });
		}
		pending.current = null;
		try {
			const result = await transcribeAudio(
				conversationId,
				blob,
				picked?.filename ?? "retake.webm",
			);
			setWindow((w) => (w ? { ...w, text: result.text, uploading: false } : w));
		} catch (err) {
			setWindow((w) =>
				w ? { ...w, uploading: false, error: messageOf(err) } : w,
			);
		}
	}, [conversationId]);

	const startRecording = useCallback(async () => {
		const picked = pickMime();
		const e2e = readE2eFixture();
		if (e2e) {
			chunks.current = [new Blob([e2e], { type: "audio/wav" })];
			pending.current = { mime: "audio/wav", filename: "retake.wav" };
			setWindow((w) => (w ? { ...w, recording: true, error: null } : w));
			return;
		}
		if (!picked) {
			setWindow((w) =>
				w ? { ...w, error: m["voice.room.recovery.recordUnsupported"]() } : w,
			);
			return;
		}
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
			chunks.current = [];
			const rec = new MediaRecorder(stream, { mimeType: picked.mime });
			rec.ondataavailable = (e) => {
				if (e.data.size > 0) chunks.current.push(e.data);
			};
			recorder.current = rec;
			rec.start();
			setWindow((w) => (w ? { ...w, recording: true, error: null } : w));
		} catch {
			setWindow((w) =>
				w ? { ...w, error: m["voice.room.recovery.micUnavailable"]() } : w,
			);
		}
	}, []);

	return {
		window,
		open,
		close,
		// The field is the field: every keystroke lands in the one place the
		// engine will read it back from, so there is no second copy to drift.
		edit: useCallback((text: string) => {
			setWindow((w) => (w ? { ...w, text } : w));
		}, []),
		record: startRecording,
		stop: useCallback(() => void stopRecording(), [stopRecording]),
		send: useCallback((text: string) => void run("send", text), [run]),
		dismiss: useCallback(() => void run("dismiss"), [run]),
	};
}

function messageOf(err: unknown): string {
	if (err instanceof ApiError) return err.message;
	return "Something went wrong. Please try again.";
}
