import {
	ConversationStatusEnded,
	ConversationStatusFailed,
} from "@engflex/contracts";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useCallback, useRef } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { useBreadcrumbs } from "#/app/breadcrumbs";
import { VoiceRoom } from "#/features/voice/components/room/voice-room";
import { useSessionCountdown } from "#/features/voice/hooks/use-session-countdown";
import { useVoiceSession } from "#/features/voice/hooks/use-voice-session";
import { useScenario } from "#/features/voice/queries";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/_app/voice/room/$conversationId")({
	component: VoiceRoomPage,
});

function formatLeft(totalSec: number): string {
	const mm = Math.floor(totalSec / 60);
	const ss = totalSec % 60;
	return `${String(mm).padStart(2, "0")}:${String(ss).padStart(2, "0")}`;
}

/** Production injector: live session state in, same VoiceRoom out. */
function VoiceRoomPage() {
	useBreadcrumbs([
		{ label: m["nav.item.voice"](), to: APP_ROUTES.VOICE.LIST },
		{ label: m["voice.crumb.room"]() },
	]);
	const { conversationId } = Route.useParams();
	const navigate = useNavigate();
	const exit = () => navigate({ to: APP_ROUTES.VOICE.LIST });
	const session = useVoiceSession(conversationId, exit);
	const status = session.conversation.data?.status;
	const scenarioId = session.conversation.data?.scenarioId;
	const scenario = useScenario(scenarioId);
	// StrictMode-safe: the ref makes double-effect expiry a no-op.
	const expired = useRef(false);
	const expire = useCallback(() => {
		if (expired.current) return;
		expired.current = true;
		void session.handleEnd();
	}, [session]);
	const left = useSessionCountdown(scenario.data?.maxDuration, expire);

	return (
		<VoiceRoom
			session={{
				ready: session.ready,
				query: session.conversation.isPending
					? "pending"
					: session.conversation.isError
						? "error"
						: "ok",
				ended:
					status === ConversationStatusEnded ||
					status === ConversationStatusFailed,
				endPending: session.endPending,
				connectionLost: session.connectionLost,
				errorDetail: session.errorDetail,
				timeLeft:
					scenario.data?.maxDuration != null
						? m["voice.room.controls.timeLeft"]({ left: formatLeft(left) })
						: null,
				// The call header's scenario card: title + objective while the
				// session runs; free talk renders its own pair when absent.
				scene: scenario.data
					? { title: scenario.data.title, body: scenario.data.objective }
					: undefined,
				onExit: () => void session.handleExit(),
				onBack: () => void exit(),
				onEnd: () => void session.handleEnd(),
				onConnectionLost: () => session.setConnectionLost(true),
				onRuntimeError: (text) => session.setErrorDetail(text),
				isUserDisconnectingRef: session.userDisconnecting,
			}}
			connection={{
				key: conversationId,
				startBotParams: session.startBotParams,
				transformStartResponse: session.transformStartResponse,
				connect: true,
			}}
			source={{ kind: "live", conversationId }}
		/>
	);
}
