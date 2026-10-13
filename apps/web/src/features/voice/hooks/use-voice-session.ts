import type { ConversationSession } from "@engflex/contracts";
import { useEffect, useRef, useState } from "react";
import { APP_ROUTES } from "#/app/app-route";
import { useConversation, useEndConversation } from "#/features/voice/queries";
import { authToken } from "#/lib/auth-token";

/**
 * The room's session wiring: conversation query, end mutation, auth headers,
 * connection health, and the three ways out (disconnect-and-leave,
 * end-and-stay, voice-panel end). Navigation is injected as `onExit` so the
 * room stays testable; the page passes the router navigate.
 */
export function useVoiceSession(
	conversationId: string,
	onExit: () => void | Promise<void>,
) {
	const conversation = useConversation(conversationId);
	const endConversation = useEndConversation(conversationId);
	const [mounted, setMounted] = useState(false);
	const [authHeaders, setAuthHeaders] = useState<Headers | null>(null);
	const [connectionLost, setConnectionLost] = useState(false);
	const [errorDetail, setErrorDetail] = useState<string | null>(null);
	const userDisconnecting = useRef(false);

	useEffect(() => setMounted(true), []);

	useEffect(() => {
		let cancelled = false;
		void authToken().then((token) => {
			if (cancelled) return;
			const headers = new Headers();
			if (token) headers.set("Authorization", `Bearer ${token}`);
			setAuthHeaders(headers);
		});
		return () => {
			cancelled = true;
		};
	}, []);

	useEffect(() => {
		function beforeUnload(event: BeforeUnloadEvent) {
			event.preventDefault();
			event.returnValue = "";
		}
		window.addEventListener("beforeunload", beforeUnload);
		return () => window.removeEventListener("beforeunload", beforeUnload);
	}, []);

	const startBotParams = {
		endpoint: APP_ROUTES.VOICE.SIGNAL_START.replace(
			"$conversationId",
			conversationId,
		),
		// Null until the token resolves; the room renders a skeleton until
		// `ready`, so this is never sent without headers.
		headers: authHeaders ?? undefined,
	};

	function transformStartResponse(response: unknown) {
		const data = (response as { data?: ConversationSession } | null | undefined)
			?.data;
		return {
			webrtcRequestParams: {
				endpoint: APP_ROUTES.VOICE.SIGNAL_OFFER.replace(
					"$conversationId",
					conversationId,
				),
				headers: authHeaders ?? undefined,
			},
			iceConfig: data?.iceConfig
				? {
						iceServers: data.iceConfig.iceServers.map((server) => ({
							urls: server.urls,
							...(server.username ? { username: server.username } : {}),
							...(server.credential ? { credential: server.credential } : {}),
						})),
					}
				: undefined,
		};
	}

	/** End server-side, then leave: end is idempotent, so navigate regardless. */
	async function handleExit() {
		userDisconnecting.current = true;
		try {
			await endConversation.mutateAsync();
		} catch {
			// end is idempotent server-side; navigate regardless
		}
		await onExit();
	}

	// End and stay: the query invalidation flips status to ended, which
	// renders the persisted review instead of navigating away.
	async function handleEnd() {
		userDisconnecting.current = true;
		try {
			await endConversation.mutateAsync();
		} catch {
			// end is idempotent server-side; the review renders regardless
		}
	}

	return {
		conversation,
		endPending: endConversation.isPending,
		ready: mounted && authHeaders !== null,
		startBotParams,
		transformStartResponse,
		connectionLost,
		setConnectionLost,
		errorDetail,
		setErrorDetail,
		userDisconnecting,
		handleExit,
		handleEnd,
	};
}
