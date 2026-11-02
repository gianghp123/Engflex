import type {
	SaveVideoProgress,
	VideoAttemptResponse,
} from "@engflex/contracts";
import { useRef, useState } from "react";
import { ApiError } from "#/lib/api/api";
import {
	getOpenVideoAttempt,
	saveVideoProgress,
	startVideoAttempt,
} from "./service";

// One writer per workspace. Intermediate pending snapshots are replaced by
// the latest draft; each write uses the revision acknowledged by the server.
export function useVideoProgressQueue(
	videoId: string,
	initial: VideoAttemptResponse | null,
) {
	const attempt = useRef(initial);
	const pending = useRef<Omit<SaveVideoProgress, "revision"> | null>(null);
	const running = useRef<Promise<boolean> | null>(null);
	const blocked = useRef(false);
	const [error, setError] = useState<unknown>(null);
	const [busy, setBusy] = useState(false);

	const save = (
		draft: Omit<SaveVideoProgress, "revision">,
	): Promise<boolean> => {
		if (blocked.current) return Promise.resolve(false);
		pending.current = draft;
		if (running.current) return running.current;
		setBusy(true);
		running.current = (async () => {
			try {
				attempt.current ??= await startVideoAttempt(videoId, "dictation");
				while (pending.current) {
					const next = pending.current;
					pending.current = null;
					attempt.current = await saveVideoProgress(
						videoId,
						attempt.current.id,
						{
							...next,
							revision: attempt.current.revision,
						},
					);
				}
				setError(null);
				return true;
			} catch (cause) {
				pending.current = null;
				blocked.current = true;
				setError(cause);
				return false;
			} finally {
				running.current = null;
				setBusy(false);
			}
		})();
		return running.current;
	};

	const reload = async () => {
		setBusy(true);
		try {
			const latest = await getOpenVideoAttempt(videoId, "dictation");
			attempt.current = latest;
			blocked.current = false;
			setError(null);
			return latest;
		} catch (cause) {
			setError(cause);
			return null;
		} finally {
			setBusy(false);
		}
	};

	const recover = async () => {
		setBusy(true);
		try {
			let latest: VideoAttemptResponse;
			try {
				latest = await getOpenVideoAttempt(videoId, "dictation");
			} catch (cause) {
				if (
					cause instanceof ApiError &&
					cause.status === 404 &&
					!attempt.current
				) {
					blocked.current = false;
					setError(null);
					return true;
				}
				throw cause;
			}
			const previous = attempt.current;
			if (
				previous
					? latest.id !== previous.id || latest.revision !== previous.revision
					: latest.revision !== 1
			) {
				throw new ApiError(
					409,
					"attempt changed; inspect saved progress before continuing",
				);
			}
			attempt.current = latest;
			blocked.current = false;
			setError(null);
			return true;
		} catch (cause) {
			setError(cause);
			return false;
		} finally {
			setBusy(false);
		}
	};

	const restart = async () => {
		if (running.current) await running.current;
		attempt.current = null;
		blocked.current = false;
		setError(null);
	};
	return { save, reload, recover, restart, attempt, error, busy };
}
