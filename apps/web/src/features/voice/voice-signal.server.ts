import "@tanstack/react-start/server-only";

import { auth } from "@clerk/tanstack-react-start/server";
import { getRequestHeaders } from "@tanstack/react-start/server";
import { API_INTERNAL_URL_FALLBACK, API_ROUTES } from "#/app/api-routes";
import { pickForwardedHeaders } from "#/lib/forward-headers.server";

const INTERNAL_API_URL =
	process.env.API_INTERNAL_URL ?? API_INTERNAL_URL_FALLBACK;

export type SignalAction = "start" | "offer";

/**
 * Forwards one voice-signaling request to the Go API byte-verbatim: method,
 * body bytes, status, and content-type all survive untouched (the offer answer
 * is raw SDP — any JSON parse/re-serialize in this path corrupts it). Auth
 * mirrors fetchApiServer: Clerk cookie → Bearer minted server-side, all
 * incoming headers forwarded so the pre-auth session limiter (CF key with
 * ClientIP fallback) keeps bucketing per client.
 */
export async function forwardVoiceSignal(args: {
	conversationId: string;
	action: SignalAction;
	method: string;
	body: ArrayBuffer | null;
	contentType: string | null;
}): Promise<Response> {
	const { conversationId, action, method, body, contentType } = args;
	if (!conversationId || conversationId.includes("/")) {
		return Response.json({ message: "bad conversation id" }, { status: 400 });
	}
	const { getToken } = await auth();
	const token = await getToken();
	if (!token) {
		return Response.json({ message: "unauthorized" }, { status: 401 });
	}
	// Forward all incoming headers (framing stripped by the helper): the
	// limiter keys on CF-Connecting-IP, and byte-verbatim body handling
	// stays untouched — headers never alter the SDP bytes. Explicit
	// Content-Type/Authorization win over forwarded copies.
	const path =
		action === "start"
			? API_ROUTES.CONVERSATIONS.START(conversationId)
			: API_ROUTES.CONVERSATIONS.OFFER(conversationId);
	const forwarded = pickForwardedHeaders(getRequestHeaders());
	if (contentType) {
		forwarded["Content-Type"] = contentType;
	}
	forwarded.Authorization = `Bearer ${token}`;
	let upstream: Response;
	try {
		upstream = await fetch(`${INTERNAL_API_URL}${path}`, {
			method,
			headers: forwarded,
			body: body ?? undefined,
			// Voice provisioning is slow; e2e budgets 30 s for start.
			// Abort surfaces as 503 below, which the Pipecat client
			// retries (5xx ≤ 3 attempts; 4xx fails immediately).
			signal: AbortSignal.timeout(30_000),
		});
	} catch {
		return Response.json(
			{ message: "voice service unreachable" },
			{ status: 503 },
		);
	}
	const bytes = await upstream.arrayBuffer();
	const upstreamContentType = upstream.headers.get("content-type");
	return new Response(bytes, {
		status: upstream.status,
		headers: upstreamContentType
			? { "content-type": upstreamContentType }
			: undefined,
	});
}
