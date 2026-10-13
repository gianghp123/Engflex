import "@tanstack/react-start/server-only";

// Hop-by-hop and framing headers must never be forwarded: fetch recomputes
// framing (a stale Content-Length breaks FormData re-serialization with its
// new boundary) and Host/Connection are fetch-forbidden — Undici drops them,
// but dropping explicitly keeps the contract visible. Everything else —
// CF-Connecting-IP, X-Forwarded-For chains, client hints — passes through
// untouched so Go sees what the edge and proxies saw. Callers layer their
// own Content-Type/Authorization on top (JSON default, minted Bearer).
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
