import { createFileRoute } from "@tanstack/react-router";
import { forwardVoiceSignal } from "#/features/voice/voice-signal.server";

export const Route = createFileRoute("/api/voice/$conversationId/offer")({
	server: {
		handlers: {
			POST: async ({ request, params }) => {
				return forwardVoiceSignal({
					conversationId: params.conversationId,
					action: "offer",
					method: request.method,
					body: await request.arrayBuffer(),
					contentType: request.headers.get("content-type"),
				});
			},
			PATCH: async ({ request, params }) => {
				return forwardVoiceSignal({
					conversationId: params.conversationId,
					action: "offer",
					method: request.method,
					body: await request.arrayBuffer(),
					contentType: request.headers.get("content-type"),
				});
			},
		},
	},
});
