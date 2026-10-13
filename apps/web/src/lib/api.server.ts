// The only module in the web app that talks to the API — and it runs
// exclusively on the server. Two guards keep it that way: the `*.server.*`
// filename is denied to the client graph by TanStack Start's
// import-protection plugin, and the marker below makes the restriction
// independent of the filename. The serverFn boundary in `api.ts` is the
// only legal importer; anywhere else, the client build fails.
import "@tanstack/react-start/server-only";

import { auth } from "@clerk/tanstack-react-start/server";
import type { ApiResponse } from "@engflex/contracts";
import { getRequestHeaders } from "@tanstack/react-start/server";
import { API_INTERNAL_URL_FALLBACK } from "#/app/api-routes";
import type { ProxyRequest, ProxyResult } from "./api-base.ts";
import { pickForwardedHeaders } from "./forward-headers.server.ts";

// Server-side API base: internal plain-HTTP service address. The browser
// never uses this (voice signaling goes same-origin, everything else rides
// serverFn).
const SERVER_API_URL =
	process.env.API_INTERNAL_URL ?? API_INTERNAL_URL_FALLBACK;

/**
 * Direct call to the API from the server. When the request asks for
 * credentials (the default), the session cookie is exchanged for a Bearer
 * via Clerk, so browser-executed RPCs and SSR renders travel the same
 * authenticated path; `withCredentials: false` skips token minting
 * entirely for public endpoints. The caller's IP is forwarded so the API's
 * rate limiter keeps bucketing per client.
 *
 * Not a privilege escalation: the base URL is fixed to our API, paths are
 * validated to stay relative upstream, and only the caller's own session
 * token is used — it can do what the caller's browser could. Returns a
 * `ProxyResult`: the parsed house envelope (the generated `ApiResponse`
 * DTO), or raw text for non-JSON bodies. `ApiError` construction stays in
 * `api.ts` so every caller sees one contract.
 */
export async function fetchApiServer(req: ProxyRequest): Promise<ProxyResult> {
	let token: string | null = null;
	if (req.withCredentials) {
		const { getToken } = await auth();
		token = await getToken();
	}
	// Forward every incoming header except framing (fetch recomputes it; a
	// stale Content-Length breaks FormData with its new boundary). The API
	// keys its limiter on CF-Connecting-IP with ClientIP fallback, so edge
	// and proxy headers must arrive intact — filtering here would silently
	// re-bucket clients. Content-Type keeps its existing rule (fetch sets
	// the multipart boundary for FormData); the minted Bearer below wins
	// over any forwarded Authorization.
	const isForm = req.body instanceof FormData;
	const headers: Record<string, string> = {
		...pickForwardedHeaders(getRequestHeaders()),
	};
	if (isForm) {
		delete headers["content-type"];
	} else {
		headers["Content-Type"] = "application/json";
	}
	if (token) {
		headers.Authorization = `Bearer ${token}`;
	}
	const res = await fetch(`${SERVER_API_URL}${req.path}`, {
		method: req.method,
		headers,
		body: req.body ?? undefined,
	});
	const text = await res.text();
	try {
		return {
			kind: "json",
			status: res.status,
			envelope: JSON.parse(text) as ApiResponse,
		};
	} catch {
		return { kind: "raw", status: res.status, text };
	}
}
