import type { Turn } from "@engflex/contracts";
import type {
	ConversationMessage,
	ConversationMessagePart,
} from "@pipecat-ai/client-react";
import { usePipecatConversation } from "@pipecat-ai/client-react";
import { PencilLine } from "lucide-react";
import type { ReactNode } from "react";
import { useEffect, useRef, useState } from "react";
import { Avatar, AvatarFallback } from "#/components/ui/avatar";
import { Badge } from "#/components/ui/badge";
import { Button } from "#/components/ui/button";
import { ScrollArea } from "#/components/ui/scroll-area";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "#/components/ui/tooltip";
import {
	type AppliedCorrection,
	applyCorrection,
	type Group,
} from "#/features/voice/components/transcript/apply-correction";
import { CorrectionModal } from "#/features/voice/components/transcript/correction-modal";
import type { AnalyzeControl } from "#/features/voice/components/transcript/turn-feedback";
import { TurnFeedbackPanel } from "#/features/voice/components/transcript/turn-feedback";
import {
	type CorrectionWindow,
	useCorrectionModal,
} from "#/features/voice/hooks/use-correction-modal";
import { useAnalyzeTurn, useConversation } from "#/features/voice/queries";
import { cn } from "#/lib/utils";
import { m } from "#/paraglide/messages";

/** Plain-text extraction for a pipecat message part. */
function getPartText(part: ConversationMessagePart): string {
	const text = part.text;
	if (typeof text === "string") return text;
	if (
		text !== null &&
		typeof text === "object" &&
		"spoken" in text &&
		"unspoken" in text
	) {
		const value = text as { spoken: string; unspoken: string };
		return [value.spoken, value.unspoken].filter(Boolean).join(" ");
	}
	return "";
}

function getMessageText(message: ConversationMessage): string {
	return (
		message.parts
			?.map(getPartText)
			.filter(Boolean)
			.join(" ")
			.replace(/\s+/g, " ")
			.trim() ?? ""
	);
}

export type TranscriptRow = {
	key: string;
	isUser: boolean;
	text: string;
	accessory?: ReactNode;
};

/** Shared bubble list: the live panel and the persisted review render rows
identically; only the accessory below a bubble differs. */
export function TranscriptRows({ rows }: { rows: TranscriptRow[] }) {
	return (
		<ul className="flex flex-col gap-4">
			{rows.map((row) => (
				<li
					key={row.key}
					className={cn(
						"flex min-w-0 gap-2",
						row.isUser ? "flex-row-reverse" : "flex-row",
					)}
				>
					<Avatar size="sm">
						<AvatarFallback>
							{row.isUser
								? m["voice.room.transcript.speaker.you"]().slice(0, 1)
								: m["voice.room.transcript.speaker.partner"]().slice(0, 1)}
						</AvatarFallback>
					</Avatar>
					<div
						className={cn(
							"flex min-w-0 max-w-[85%] flex-col gap-1",
							row.isUser ? "items-end" : "items-start",
						)}
					>
						<span className="text-[11px] font-medium text-muted-foreground">
							{row.isUser
								? m["voice.room.transcript.speaker.you"]()
								: m["voice.room.transcript.speaker.partner"]()}
						</span>
						<p
							className={cn(
								"rounded-xl px-3 py-2 text-sm break-words",
								row.isUser
									? "bg-primary text-primary-foreground"
									: "bg-muted text-foreground",
							)}
						>
							{row.text}
						</p>
						{row.accessory}
					</div>
				</li>
			))}
		</ul>
	);
}

/**
 * Collapse live messages into bubbles exactly like the engine collector
 * does, so group ordinal aligns with persisted positions. Pure: the
 * container and the DEV preview share it.
 */
export function buildGroups(messages: ConversationMessage[]): Group[] {
	const rows = messages
		.map((message, index) => ({
			message,
			index,
			text: getMessageText(message),
		}))
		.filter(
			(row) =>
				(row.message.role === "user" || row.message.role === "assistant") &&
				row.text.length > 0,
		);
	const groups: Group[] = [];
	let prevRole: string | null = null;
	for (const row of rows) {
		const isUser = row.message.role === "user";
		const last = groups[groups.length - 1];
		if (last && prevRole === row.message.role) {
			last.text = `${last.text} ${row.text}`.trim();
		} else {
			groups.push({
				key: `${row.message.createdAt}-${row.index}`,
				isUser,
				text: row.text,
			});
		}
		prevRole = row.message.role;
	}
	return groups;
}

/** Fixture bridge for the DEV preview: persisted turns as live messages. */
export function turnsToMessages(turns: Turn[]): ConversationMessage[] {
	return turns.map((turn) => ({
		role: turn.role === "user" ? "user" : "assistant",
		createdAt: turn.createdAt,
		parts: [{ text: turn.text, final: true, createdAt: turn.createdAt }],
	}));
}

/**
 * The transcript view: pure rendering over already-resolved data. No pipecat
 * provider, no network, no modal — the container injects all of that, and
 * the DEV preview injects fixtures into the same container instead.
 * Tweak layout here and see it in both places.
 */
function TranscriptPanelView({
	className,
	groups,
	applied,
	persistedTurns,
	analyzePending,
	analyzeFailed,
	onAnalyze,
	canReview,
	onReview,
}: {
	className?: string;
	groups: Group[];
	applied: AppliedCorrection | null;
	persistedTurns: Turn[];
	analyzePending: boolean;
	analyzeFailed: boolean;
	onAnalyze: (position: number) => void;
	canReview: boolean;
	onReview: () => void;
}) {
	const bottomRef = useRef<HTMLDivElement>(null);
	const shown = applyCorrection(groups, applied);
	const persistedByPosition = new Map(
		persistedTurns.map((turn) => [turn.position, turn]),
	);

	// Scroll on what is displayed, not on what pipecat holds, so dropping a
	// stale reply does not leave the view parked past the end.
	const rowCount = shown.length;
	useEffect(() => {
		if (rowCount === 0) return;
		bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
	}, [rowCount]);

	const lastUserIndex = (() => {
		for (let i = shown.length - 1; i >= 0; i--) {
			if (shown[i].isUser) return i;
		}
		return -1;
	})();

	return (
		<div
			className={cn(
				"surface-card flex h-full min-h-0 flex-col gap-6 py-6",
				className,
			)}
		>
			<div className="flex items-center gap-2 px-6 text-base leading-none font-semibold">
				{m["voice.room.transcript.title"]()}
				<Badge variant="secondary">{shown.length}</Badge>
			</div>
			<div className="min-h-0 flex-1 px-6 pb-6">
				<ScrollArea className="h-full max-h-[55vh] min-h-0 overflow-x-clip pr-3 lg:max-h-none">
					{shown.length === 0 ? (
						<p className="py-8 text-center text-sm text-muted-foreground">
							{m["voice.room.transcript.empty"]()}
						</p>
					) : (
						<TranscriptRows
							rows={shown.map((group, groupIndex) => {
								// Global position: group ordinal, aligned with the
								// persisted turns. The Analyze button appears only
								// once the row exists server-side, so a transient
								// mismatch can delay it but never misattach feedback.
								const position = groupIndex + 1;
								const persistedTurn = persistedByPosition.get(position);
								const showAnalyze = group.isUser && persistedTurn !== undefined;
								// Review belongs to the turn being corrected, so it
								// sits under that bubble rather than in the header
								// where it read as a panel-level action.
								const showReview = canReview && groupIndex === lastUserIndex;
								return {
									key: group.key,
									isUser: group.isUser,
									text: group.text,
									accessory:
										showReview || showAnalyze ? (
											<div className="flex flex-wrap items-start gap-1.5">
												{showReview && (
													<Tooltip>
														<TooltipTrigger asChild>
															<Button
																type="button"
																variant="outline"
																size="icon-sm"
																aria-label={m[
																	"voice.room.recovery.reviewTranscript"
																]()}
																onClick={onReview}
															>
																<PencilLine />
															</Button>
														</TooltipTrigger>
														<TooltipContent>
															{m["voice.room.recovery.reviewTranscript"]()}
														</TooltipContent>
													</Tooltip>
												)}
												{showAnalyze && (
													<TurnFeedbackPanel
														feedback={persistedTurn.feedback?.payload}
														utterance={group.text}
														pending={analyzePending}
														failed={analyzeFailed}
														onAnalyze={() => onAnalyze(position)}
													/>
												)}
											</div>
										) : undefined,
								};
							})}
						/>
					)}
					<div ref={bottomRef} />
				</ScrollArea>
			</div>
		</div>
	);
}

/**
 * The transcript: one component for the room and the DEV preview. The room
 * passes `conversationId` and gets the live conversation, the persisted
 * poll, analyze, and the engine-backed correction modal. The preview passes
 * fixtures instead — same bubbles, same buttons, same modal, zero network.
 */
export function TranscriptPanel({
	className,
	messages: messagesOverride,
	conversationId,
	persistedTurns: persistedTurnsOverride,
	allowReview,
	analyze: analyzeOverride,
	fixtureText,
}: {
	className?: string;
	messages?: ConversationMessage[];
	conversationId?: string;
	/** Skip the persisted poll and use these turns (preview fixtures). */
	persistedTurns?: Turn[];
	/** Show the edit button without a live session (preview). */
	allowReview?: boolean;
	/** Skip the analyze mutation and use this control (preview noop). */
	analyze?: AnalyzeControl;
	/**
	 * Run the correction modal on local state seeded with this text instead
	 * of the engine (preview). Only honored without a `conversationId`:
	 * open seeds the field, typing edits it, record/stop flip their flags
	 * for visual tweaking, and send runs the real applied-derivation below.
	 */
	fixtureText?: string;
}) {
	const live = usePipecatConversation();
	const messages = messagesOverride ?? live.messages;

	// Live persisted turns, polled while a session is running. Positions are
	// assigned by finalize order on both sides, so group ordinal == position.
	// Fixtures skip the poll entirely (no network).
	const useFixtures = persistedTurnsOverride !== undefined;
	const persisted = useConversation(conversationId ?? "", {
		refetchInterval: conversationId && !useFixtures ? 2000 : undefined,
		enabled: !useFixtures,
	});
	const analyzeLive = useAnalyzeTurn(conversationId ?? "");
	const analyze: AnalyzeControl = analyzeOverride ?? {
		pending: analyzeLive.isPending,
		failed: analyzeLive.isError,
		onAnalyze: (position) => analyzeLive.mutate(position),
	};
	// The last correction the engine accepted, plus how many bubbles existed
	// when it landed. Recorded here because the panel owns what is displayed.
	const [applied, setApplied] = useState<AppliedCorrection | null>(null);
	const groups = buildGroups(messages);
	// This closure is recreated every render, so `groups` below is the
	// pre-send list: the tail snapshot is what the send actually saw.
	const recordCorrection = (text: string) => {
		let lastUser = -1;
		for (let i = groups.length - 1; i >= 0; i--) {
			if (groups[i].isUser) {
				lastUser = i;
				break;
			}
		}
		setApplied({
			text,
			boundary: groupCountRef.current,
			staleTail: groups
				.slice(lastUser + 1)
				.map((g) => ({ key: g.key, text: g.text })),
		});
	};
	const liveCorrection = useCorrectionModal(conversationId ?? null, {
		onCorrected: recordCorrection,
	});
	// Preview-only modal: same CorrectionModal, local state, no engine. The
	// send path is shared — the corrected bubble and the regenerated-reply
	// derivation render exactly as they do live.
	const [fixtureWindow, setFixtureWindow] = useState<CorrectionWindow | null>(
		null,
	);
	const useFixtureModal = fixtureText !== undefined && !conversationId;
	const correction = useFixtureModal
		? {
				window: fixtureWindow,
				open: async () => {
					setFixtureWindow({
						text: fixtureText,
						recording: false,
						uploading: false,
						submitting: false,
						error: null,
					});
				},
				close: () => setFixtureWindow(null),
				edit: (text: string) =>
					setFixtureWindow((w) => (w ? { ...w, text } : w)),
				record: () =>
					setFixtureWindow((w) => (w ? { ...w, recording: true } : w)),
				stop: () =>
					setFixtureWindow((w) =>
						w ? { ...w, recording: false, uploading: false } : w,
					),
				send: (text: string) => {
					recordCorrection(text);
					setFixtureWindow(null);
				},
				dismiss: () => setFixtureWindow(null),
			}
		: liveCorrection;

	// The boundary is captured when the learner sends, so it must read the
	// bubbles as they were at that moment — not as they are after the
	// correction has already been applied.
	const shownLen = applyCorrection(groups, applied).length;
	const groupCountRef = useRef(shownLen);
	useEffect(() => {
		groupCountRef.current = shownLen;
	}, [shownLen]);

	// The learner reviews their most recent turn. Nothing decides for them
	// that a turn was misheard, so this is available whenever there is
	// something to review; the engine resolves the target itself.
	const canReview = Boolean(
		(conversationId || allowReview) && groups.some((g) => g.isUser),
	);
	const persistedTurns = persistedTurnsOverride ?? persisted.data?.turns ?? [];

	return (
		<>
			{correction.window && (
				<CorrectionModal
					window={correction.window}
					onSubmit={correction.send}
					onEdit={correction.edit}
					onRecord={correction.record}
					onStop={correction.stop}
					onDismiss={correction.dismiss}
				/>
			)}
			<TranscriptPanelView
				className={className}
				groups={groups}
				applied={applied}
				persistedTurns={persistedTurns}
				analyzePending={analyze.pending}
				analyzeFailed={analyze.failed}
				onAnalyze={analyze.onAnalyze}
				canReview={canReview}
				onReview={() => void correction.open()}
			/>
		</>
	);
}
