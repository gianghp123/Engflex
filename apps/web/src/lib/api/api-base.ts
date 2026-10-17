// Isomorphic API primitives: base URL + query-string join + the wire shapes
// shared by the serverFn proxy (`api.ts`) and the server-only API transport
// (`api.server.ts`).
import type { ApiResponse } from "@engflex/contracts";

/** A request as it crosses the serverFn boundary: an API-relative path (the
 *  validators in `api.ts` reject anything else), the HTTP method, a JSON
 *  string or multipart body, and whether the caller wants the session
 *  credentials attached. */
export type ProxyRequest = {
	path: string;
	method: string;
	body?: string | FormData | null;
	withCredentials: boolean;
};

/** Result of one proxied call. `json` carries the parsed house envelope —
 *  the generated `ApiResponse` DTO, single source of truth in
 *  `@engflex/contracts`; `raw` carries the body when upstream did not answer
 *  JSON (callers must turn it into an ApiError regardless of status). */
export type ProxyResult =
	| { kind: "json"; status: number; envelope: ApiResponse }
	| { kind: "raw"; status: number; text: string };

/** Appends `page` / `pageSize` style params to a path. */
export function withQuery(
	path: string,
	params: Record<string, string | number>,
): string {
	const search = new URLSearchParams();
	for (const [key, value] of Object.entries(params)) {
		search.set(key, String(value));
	}
	const query = search.toString();
	return query ? `${path}?${query}` : path;
}
