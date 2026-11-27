import type {
	VideoAttemptResponse,
	VideoExerciseResponse,
} from "@engflex/contracts";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRef, useState } from "react";
import { videoKeys } from "./queries";
import {
	completeDictationAttempt,
	saveDictationProgress,
	startShadowingAttempt,
} from "./service";

type Checkpoint = {
	activeIndex: number;
	completed: Set<string>;
};

function restoreCheckpoint(
	lines: NonNullable<VideoExerciseResponse["transcripts"]>,
	result: VideoAttemptResponse["result"],
): Checkpoint {
	const completed = new Set(result?.completedTranscriptIds);
	const activeIndex = result?.activeTranscriptId
		? lines.findIndex((line) => line.id === result.activeTranscriptId)
		: (result?.activeSequence ?? 0);
	return {
		activeIndex:
			activeIndex >= 0 && activeIndex < lines.length ? activeIndex : 0,
		completed,
	};
}

/**
 * Tracks shadowing progress by line: which transcript lines the learner has
 * practiced through and where they left off, so the session resumes instead
 * of restarting at line 1. "Completed" means "practiced", not "scored above
 * some bar": a line is marked practiced the moment the learner navigates
 * away from it via {@link setActiveIndex}, mirroring dictation's model
 * where activeIndex is "the line currently being worked" — independent of
 * the AI pronunciation score a recording on that line may have earned (see
 * usePronounceShadowing), which the backend persists per line but never
 * gates practiced-status on.
 */
export function useShadowingAttempt(
	video: VideoExerciseResponse,
	initialAttempt: VideoAttemptResponse | null,
) {
	const lines = video.transcripts ?? [];
	const [checkpoint, setCheckpoint] = useState(() =>
		restoreCheckpoint(lines, initialAttempt?.result),
	);
	const saved = useRef(checkpoint);
	const attemptId = useRef(initialAttempt?.id);
	const [result, setResult] = useState(initialAttempt);
	const finished = result?.status === "completed";
	const queryClient = useQueryClient();

	/** Lazily opens the attempt on first use — recording a line before ever
	 *  navigating away from it hits this same path as the checkpoint save. */
	async function ensureAttemptId(): Promise<string> {
		if (!attemptId.current) {
			const opened = await startShadowingAttempt(video.id);
			attemptId.current = opened.id;
		}
		return attemptId.current;
	}

	const { mutate, isPending, isError, reset } = useMutation({
		mutationFn: async (next: Checkpoint) => {
			const id = await ensureAttemptId();
			const completedIds = [...next.completed];
			const result = await saveDictationProgress(video.id, id, {
				activeSequence: next.activeIndex,
				activeTranscriptId: lines[next.activeIndex]?.id,
				completedTranscriptIds: completedIds,
				transcripts: {},
			});
			if (result.status === "completed") return result;
			if (lines.length > 0 && completedIds.length >= lines.length) {
				return completeDictationAttempt(video.id, id);
			}
			return result;
		},
		onSuccess: (response, next) => {
			saved.current = next;
			setResult(response);
			if (response.status === "completed") {
				queryClient.setQueryData(videoKeys.openShadowing(video.id), null);
				void queryClient.invalidateQueries({
					queryKey: videoKeys.all,
					refetchType: "inactive",
				});
			}
		},
	});

	const completedCount = checkpoint.completed.size;

	/**
	 * Moves to `index`, marking the line being left (the current
	 * activeIndex) as practiced. Call this instead of a plain local
	 * setState for every line navigation — first load has nothing to mark
	 * since there's no prior line yet.
	 */
	function setActiveIndex(index: number) {
		if (finished) return;
		setCheckpoint((previous) => {
			if (previous.activeIndex === index) return previous;
			const leaving = lines[previous.activeIndex];
			const completed = leaving
				? new Set(previous.completed).add(leaving.id)
				: previous.completed;
			const next: Checkpoint = { activeIndex: index, completed };
			mutate(next);
			return next;
		});
	}

	return {
		activeIndex: checkpoint.activeIndex,
		completed: checkpoint.completed,
		completedCount,
		finished,
		setActiveIndex,
		ensureAttemptId,
		isSaving: isPending,
		saveFailed: isError,
		retrySave: () => mutate(checkpoint),
		restart: () => {
			const empty = restoreCheckpoint(lines, undefined);
			attemptId.current = undefined;
			saved.current = empty;
			setResult(null);
			setCheckpoint(empty);
			reset();
		},
	};
}
