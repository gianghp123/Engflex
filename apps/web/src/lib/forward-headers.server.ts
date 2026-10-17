import "@tanstack/react-start/server-only";

// Hop-by-hop and framing headers must never be forwarded: fetch recomputes
// framing (a stale Content-Length breaks FormData re-serialization with its
// new boundary) and Host/Connection are fetch-forbidden — Undici drops them,
// but dropping explicitly keeps the contract visible.
//
// `Origin` is dropped for a different reason: our only client is the browser
// talking to *us*, and the API is internal-only (the browser never reaches it
// directly — every call rides this proxy). Forwarding the browser's origin
// turns an authenticated server-to-server hop into something the API reads as
// a cross-origin browser request, and gin-contrib/cors rejects it with an
// empty-body 403 *before* RequireAuth runs — one warn log line with no
// user_id, no handler, and nothing else to go on. The API reads Origin
// nowhere else, so nothing downstream misses it.
//
// Everything else — CF-Connecting-IP, X-Forwarded-For chains, client hints —
// passes through untouched so Go sees what the edge and proxies saw. Overrides
// (Content-Type JSON default, minted Authorization Bearer) must go through
// `set()` on a Headers instance built from the result: the returned keys are
// lowercase, so a record-style capitalized assignment adds a SECOND entry
// that fetch merges into one invalid value instead of replacing the original.
const DROP = new Set([
	"connection",
	"keep-alive",
	"proxy-authenticate",
	"proxy-authorization",
	"te",
	"trailer",
	"transfer-encoding",
	"upgrade",
	"host",
	"content-length",
	// Browser-provenance header: see the note above — forwarded, it makes the
	// API's CORS middleware reject this trusted hop with an empty-body 403.
	"origin",
]);

export function pickForwardedHeaders(headers: Headers): Record<string, string> {
	const out: Record<string, string> = {};
	headers.forEach((value, key) => {
		if (!DROP.has(key.toLowerCase()) && value !== "") {
			out[key] = value;
		}
	});
	return out;
}
