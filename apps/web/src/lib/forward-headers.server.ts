import "@tanstack/react-start/server-only";

// Hop-by-hop and framing headers must never be forwarded: fetch recomputes
// framing (a stale Content-Length breaks FormData re-serialization with its
// new boundary) and Host/Connection are fetch-forbidden — Undici drops them,
// but dropping explicitly keeps the contract visible. Everything else —
// CF-Connecting-IP, X-Forwarded-For chains, client hints — passes through
// untouched so Go sees what the edge and proxies saw. Overrides
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
