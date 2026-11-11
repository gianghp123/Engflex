import type {
	DictationLineInput,
	DictationLineResult,
	VideoAttemptResponse,
	VideoExerciseResponse,
} from "@engflex/contracts";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { checkDictation, type DictationLevel } from "./dictation";
import { videoKeys } from "./queries";
import {
	completeDictationAttempt,
	saveDictationProgress,
	startDictationAttempt,
} from "./service";

export type LinePractice = DictationLineInput &
	Pick<DictationLineResult, "earned" | "total">;

function attemptLevel(result: VideoAttemptResponse["result"]): DictationLevel {
	if (result?.dictationVersion !== 1) return "hard";
	return result.level === "easy" || result.level === "normal"
		? result.level
		: "hard";
}

type Checkpoint = {
	activeIndex: number;
	practice: Record<string, LinePractice>;
};

function restoreCheckpoint(
	lines: NonNullable<VideoExerciseResponse["transcripts"]>,
	result: VideoAttemptResponse["result"],
): Checkpoint {
	const completed = new Set(result?.completedTranscriptIds);
	const practice: Checkpoint["practice"] = {};
	for (const { id, content } of lines) {
		const raw = result?.transcripts?.[id];
		const state =
			raw && typeof raw === "object" ? (raw as Partial<LinePractice>) : null;
		if (!state && !completed.has(id)) continue;
		const positions = new Set(
			checkDictation(content, "", attemptLevel(result))
				.slots.filter((slot) => slot.blankIndex !== undefined)
				.map((slot) => slot.position),
		);
		practice[id] = {
			draft: typeof state?.draft === "string" ? state.draft : "",
			answers: Array.isArray(state?.answers)
				? state.answers.map((value) => (typeof value === "string" ? value : ""))
				: undefined,
			revealed: Array.isArray(state?.revealed)
				? [...new Set(state.revealed.filter((index) => positions.has(index)))]
				: [],
			completed: completed.has(id),
			earned: typeof state?.earned === "number" ? state.earned : undefined,
			total: typeof state?.total === "number" ? state.total : undefined,
		};
	}
	const activeIndex = result?.activeTranscriptId
		? lines.findIndex((line) => line.id === result.activeTranscriptId)
		: (result?.activeSequence ?? 0);
	return {
		activeIndex:
			activeIndex >= 0 && activeIndex < lines.length ? activeIndex : 0,
		practice,
	};
}

export function useDictationAttempt(
	video: VideoExerciseResponse,
	initialAttempt: VideoAttemptResponse | null,
) {
	const [lines] = useState(() =>
		(video.transcripts ?? []).filter(
			(line) => checkDictation(line.content, "").slots.length > 0,
		),
	);
	const [checkpoint, setCheckpoint] = useState(() =>
		restoreCheckpoint(lines ?? [], initialAttempt?.result),
	);
	const saved = useRef(checkpoint);
	const attemptId = useRef(initialAttempt?.id);
	const [level, setLevel] = useState(() =>
		attemptLevel(initialAttempt?.result),
	);
	const [result, setResult] = useState(initialAttempt);
	const finished = result?.status === "completed";
	const queryClient = useQueryClient();
	const { mutate, isPending, isError, reset } = useMutation({
		mutationFn: async (next: Checkpoint) => {
			const line = lines?.[next.activeIndex];
			if (!line) return null;
			if (!attemptId.current) {
				const opened = await startDictationAttempt(video.id, level);
				attemptId.current = opened.id;
				if (attemptLevel(opened.result) !== level) return opened;
			}
			const id = attemptId.current;

			const result = await saveDictationProgress(video.id, id, {
				activeSequence: next.activeIndex,
				activeTranscriptId: line.id,
				level,
				dictationVersion: 1,
				completedTranscriptIds: lines
					.filter((item) => next.practice[item.id]?.completed)
					.map((item) => item.id),
				transcripts: Object.fromEntries(
					Object.entries(next.practice).filter(
						([key, state]) => state !== saved.current.practice[key],
					),
				),
			});
			if (result.status === "completed") return result;

			const completed = new Set(result.result?.completedTranscriptIds);
			if (lines.every(({ id }) => completed.has(id))) {
				return completeDictationAttempt(video.id, id);
			}
			return result;
		},
		onSuccess: (response, next) => {
			if (!response) return;
			saved.current = next;
			setResult(response);
			const actualLevel = attemptLevel(response.result);
			if (actualLevel !== level) {
				const restored = restoreCheckpoint(lines, response.result);
				setLevel(actualLevel);
				saved.current = restored;
				setCheckpoint(restored);
			}
			if (response.status === "completed") {
				queryClient.setQueryData(videoKeys.openDictation(video.id), null);
				void queryClient.invalidateQueries({
					queryKey: videoKeys.all,
					refetchType: "inactive",
				});
			}
		},
	});

	const { activeIndex } = checkpoint;
	const practice = { ...checkpoint.practice };
	for (const [id, state] of Object.entries(
		restoreCheckpoint(lines, result?.result).practice,
	)) {
		if (state.completed || !practice[id]) practice[id] = state;
	}
	const completedCount =
		lines?.filter((line) => practice[line.id]?.completed).length ?? 0;
	const allCompleted = Boolean(
		lines?.length && completedCount === lines.length,
	);
	const dirty =
		checkpoint !== saved.current &&
		Boolean(attemptId.current || Object.keys(checkpoint.practice).length);
	const hasUnsavedChanges = dirty && !finished;

	useEffect(() => {
		if (!hasUnsavedChanges) return;
		const preventReload = (event: BeforeUnloadEvent) => event.preventDefault();
		window.addEventListener("beforeunload", preventReload);
		return () => window.removeEventListener("beforeunload", preventReload);
	}, [hasUnsavedChanges]);

	useEffect(() => {
		if (
			finished ||
			isPending ||
			isError ||
			(!dirty && !allCompleted) ||
			(!attemptId.current && Object.keys(checkpoint.practice).length === 0)
		)
			return;
		const immediate =
			allCompleted ||
			activeIndex !== saved.current.activeIndex ||
			Object.entries(checkpoint.practice).some(
				([id, state]) =>
					(state.completed && !saved.current.practice[id]?.completed) ||
					state.revealed.length !==
						(saved.current.practice[id]?.revealed.length ?? 0),
			);
		if (immediate) {
			mutate(checkpoint);
			return;
		}
		const timer = window.setTimeout(() => mutate(checkpoint), 400);
		return () => window.clearTimeout(timer);
	}, [
		activeIndex,
		allCompleted,
		checkpoint,
		dirty,
		finished,
		isError,
		isPending,
		mutate,
	]);

	return {
		activeIndex,
		lines,
		practice,
		level,
		setLevel,
		levelLocked: Boolean(
			attemptId.current || Object.keys(checkpoint.practice).length,
		),
		finished,
		score: result?.score,
		restart: () => {
			const empty = restoreCheckpoint(lines, undefined);
			attemptId.current = undefined;
			saved.current = empty;
			setResult(null);
			setCheckpoint(empty);
			setLevel("hard");
			reset();
		},
		completedCount,
		isSaving: isPending || (hasUnsavedChanges && !isError),
		saveFailed: isError,
		retrySave: reset,
		setActiveIndex: (index: number) =>
			setCheckpoint((previous) =>
				previous.activeIndex === index
					? previous
					: { ...previous, activeIndex: index },
			),
		setPractice: (
			update: (previous: Checkpoint["practice"]) => Checkpoint["practice"],
		) =>
			setCheckpoint((previous) => ({
				...previous,
				practice: update(previous.practice),
			})),
	};
}
