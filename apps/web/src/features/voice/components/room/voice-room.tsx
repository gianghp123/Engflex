import type { Turn } from "@engflex/contracts";
import { PipecatAppBase } from "@pipecat-ai/voice-ui-kit";
import type { ComponentProps, RefObject } from "react";
import { EndedReview } from "#/features/voice/components/room/ended-review";
import { RoomErrorListener } from "#/features/voice/components/room/room-error-listener";
import { SessionModal } from "#/features/voice/components/room/session-modal";
import {
	SparringGuideButton,
	SparringGuideDialog,
} from "#/features/voice/components/room/sparring-guide-dialog";
import { VoicePanel } from "#/features/voice/components/room/voice-panel";
import { VoiceRoomLayout } from "#/features/voice/components/room/voice-room-layout";
import { VoiceRoomSkeleton } from "#/features/voice/components/room/voice-room-skeleton";
import { PersistedTranscript } from "#/features/voice/components/transcript/persisted-transcript";
import {
	TranscriptPanel,
	turnsToMessages,
} from "#/features/voice/components/transcript/transcript-panel";
import { useSparringGuide } from "#/features/voice/hooks/use-sparring-guide";
import { m } from "#/paraglide/messages";
import "@pipecat-ai/voice-ui-kit/styles.scoped.css";

type BaseProps = ComponentProps<typeof PipecatAppBase>;

/** Where the room's transcript data comes from. */
export type RoomSource =
	| { kind: "live"; conversationId: string }
	| {
			kind: "fixture";
			turns: Turn[];
			correctionText: string;
			transcript: "live" | "saved";
			/** Preview-only scene override (fixtures carry no scenarioId). */
			scene?: { title: string; body: string };
			/** Preview-only duration (fixtures carry no durationSec). */
			durationSec?: number;
	  };

/** Everything session-ish the room needs, as plain data + callbacks. */
export type RoomSession = {
	ready: boolean;
	query: "pending" | "error" | "ok";
	ended: boolean;
	endPending: boolean;
	connectionLost: boolean;
	errorDetail: string | null;
	/** Countdown text ("09:59 left"); null when uncapped (free talk) or unknown. */
	timeLeft?: string | null;
	/** End server-side, then leave. */
	onExit: () => void;
	/** Leave without ending (already ended). */
	onBack: () => void;
	/** End server-side and stay for the review. */
	onEnd: () => void;
	onConnectionLost: () => void;
	onRuntimeError: (detail: string | null) => void;
	isUserDisconnectingRef: RefObject<boolean>;
};

/** How the room connects. Preview injects `{ key, connect: false }` and the
base becomes provider context only — never connects, never touches devices. */
export type RoomConnection = {
	key: string;
	startBotParams?: BaseProps["startBotParams"];
	transformStartResponse?: BaseProps["startBotResponseTransformer"];
	connect: boolean;
};

/** No-op analyze for fixture sources: buttons render, clicks go nowhere. */
function noopAnalyze(_position: number): void {}

/**
 * The voice room — the ONE room component. Production and the DEV preview
 * render this same component; they only inject different `session`,
 * `connection`, and `source`. Nothing here fetches, connects, or navigates:
 * the route builds those and passes them in.
 */
export function VoiceRoom({
	session,
	connection,
	source,
}: {
	session: RoomSession;
	connection: RoomConnection;
	source: RoomSource;
}) {
	const guide = useSparringGuide();

	if (!session.ready || session.query === "pending") {
		return <VoiceRoomSkeleton />;
	}

	if (session.query === "error") {
		return (
			<div className="p-8" aria-hidden="true">
				<SessionModal
					open
					title={m["voice.connection.failed"]()}
					description={m["voice.room.sessionEnded"]()}
					actionLabel={m["voice.room.backToScenarios"]()}
					onAction={session.onExit}
				/>
			</div>
		);
	}

	if (session.ended) {
		return source.kind === "live" ? (
			<EndedReview
				conversationId={source.conversationId}
				onBack={session.onBack}
			/>
		) : (
			<EndedReview
				turns={source.turns}
				scene={source.scene}
				durationSec={source.durationSec}
				analyze={{ pending: false, failed: false, onAnalyze: noopAnalyze }}
				onBack={session.onBack}
			/>
		);
	}

	const objective = m["voice.room.freeTalk.objective"]();
	const transcript =
		source.kind === "live" ? (
			<TranscriptPanel
				className="w-full"
				conversationId={source.conversationId}
			/>
		) : source.transcript === "live" ? (
			<TranscriptPanel
				className="w-full"
				messages={turnsToMessages(source.turns)}
				persistedTurns={source.turns}
				allowReview
				analyze={{ pending: false, failed: false, onAnalyze: noopAnalyze }}
				fixtureText={source.correctionText}
			/>
		) : (
			<PersistedTranscript
				className="w-full"
				turns={source.turns}
				analyze={{ pending: false, failed: false, onAnalyze: noopAnalyze }}
			/>
		);

	return (
		<PipecatAppBase
			key={connection.key}
			transportType="smallwebrtc"
			startBotParams={connection.startBotParams}
			startBotResponseTransformer={connection.transformStartResponse}
			initDevicesOnMount={connection.connect}
			connectOnMount={connection.connect}
			noThemeProvider
		>
			{({ client, error, handleDisconnect: disconnect }) => {
				if (!client && !error) return <VoiceRoomSkeleton />;
				// Blocking modal — never a silent swap or redirect. The room
				// stays mounted behind it; leaving ends the conversation.
				return (
					<>
						{client && !session.connectionLost ? (
							<>
								<RoomErrorListener
									isUserDisconnectingRef={session.isUserDisconnectingRef}
									onConnectionLost={session.onConnectionLost}
									onRuntimeError={session.onRuntimeError}
								/>
								<VoiceRoomLayout
									objective={objective}
									voice={
										<VoicePanel
											title={m["voice.room.freeTalk.title"]()}
											objective={objective}
											ending={session.endPending}
											timeLeft={session.timeLeft}
											headerAction={
												<SparringGuideButton
													onClick={() => guide.setOpen(true)}
												/>
											}
											onEnd={() => {
												session.isUserDisconnectingRef.current = true;
												void (async () => {
													await disconnect?.();
													session.onEnd();
												})();
											}}
										/>
									}
									transcript={transcript}
								/>
							</>
						) : (
							<VoiceRoomSkeleton />
						)}
						<SessionModal
							open={error != null || session.connectionLost}
							title={m["voice.connection.failed"]()}
							description={
								error ?? session.errorDetail ?? m["voice.room.sessionEnded"]()
							}
							actionLabel={m["voice.room.backToScenarios"]()}
							onAction={session.onExit}
						/>
						<SparringGuideDialog open={guide.open} onClose={guide.close} />
					</>
				);
			}}
		</PipecatAppBase>
	);
}
