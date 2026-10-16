import { createFileRoute } from "@tanstack/react-router";
import { forwardVoiceSignal } from "#/features/voice/voice-signal.server";

export const Route = createFileRoute("/api/voice/$conversationId/start")({
	server: {
		handlers: {
			POST: async ({ request, params }) => {
				return forwardVoiceSignal({
					conversationId: params.conversationId,
					action: "start",
					method: "POST",
					body: await request.arrayBuffer(),
					contentType: request.headers.get("content-type"),
				});
			},
		},
	},
});
