// Typed client for the API. Success or failure comes from the HTTP status;
// the envelope is {message, data[, pagination]}. Framework-agnostic: no
// router/query/auth imports here.
//
// Every request executes on the server: during SSR the proxy server
// functions below run in-process, and in the browser they are POST RPCs to
// our own server (/_serverFn/*). The handler performs the only API fetch
// there is (api.server.ts, stripped from the client bundle), so the browser
// never holds API tokens and never talks to the API directly.
import type { Pagination } from "@engflex/contracts";
import { createServerFn } from "@tanstack/react-start";
import { fetchApiServer } from "./api.server.ts";
import { type ProxyRequest, type ProxyResult, withQuery } from "./api-base.ts";

export { withQuery };

export class ApiError extends Error {
	status: number;

	constructor(status: number, message: string) {
		super(message);
		this.name = "ApiError";
		this.status = status;
	}
}

/** JSON proxy payload. Validated to a relative path so the RPC can never be
 *  steered at another host; the body rides as a string because the API wire
 *  format is JSON either way. */
function validateProxyInput(raw: unknown): ProxyRequest {
	const value = (raw ?? {}) as Record<string, unknown>;
	const path = value.path;
	if (
		typeof path !== "string" ||
		!path.startsWith("/") ||
		path.startsWith("//")
	) {
		throw new Error("api: path must be a relative API path");
	}
	const method =
		typeof value.method === "string" && value.method !== ""
			? value.method
			: "GET";
	const body = typeof value.body === "string" ? value.body : null;
	const withCredentials =
		typeof value.withCredentials === "boolean" ? value.withCredentials : true;
	return { path, method, body, withCredentials };
}

const callApiServer = createServerFn({ method: "POST" })
	.validator(validateProxyInput)
	.handler(({ data }) => fetchApiServer(data));

/** Multipart proxy. FormData cannot ride the JSON payload — the serializer
 *  supports it only as the top-level value — so uploads get their own
 *  function; the target path and credential flag travel in `__path` /
 *  `__credentials` control fields, deleted before anything leaves for the
 *  API. */
function validateUploadInput(raw: unknown): FormData {
	if (!(raw instanceof FormData)) {
		throw new Error("api: upload must be FormData");
	}
	return raw;
}

const callApiUpload = createServerFn({ method: "POST" })
	.validator(validateUploadInput)
	.handler(({ data }) => {
		const path = data.get("__path");
		if (
			typeof path !== "string" ||
			!path.startsWith("/") ||
			path.startsWith("//")
		) {
			throw new Error("api: upload is missing its __path");
		}
		const withCredentials = data.get("__credentials") !== "0";
		data.delete("__path");
		data.delete("__credentials");
		return fetchApiServer({
			path,
			method: "POST",
			body: data,
			withCredentials,
		});
	});

/** Envelope handling shared by both proxies. JSON bodies arrive already
 *  parsed into the generated `ApiResponse` (fetchApiServer); the raw variant
 *  (proxy error page, plain-text limiter) must never escape as a
 *  SyntaxError: callers only catch ApiError, and anything else crashes the
 *  boundary. */
function parseEnvelope<T>(result: ProxyResult): {
	data: T;
	pagination?: Pagination;
} {
	if (result.kind === "raw") {
		throw new ApiError(
			result.status,
			result.text || `Request failed (${result.status})`,
		);
	}
	if (result.status < 200 || result.status >= 300) {
		throw new ApiError(
			result.status,
			result.envelope.message || `Request failed (${result.status})`,
		);
	}
	return {
		data: result.envelope.data as T,
		pagination: result.envelope.pagination,
	};
}

async function request<T>(
	path: string,
	init?: RequestInit,
	opts: ApiOptions = {},
): Promise<{ data: T; pagination?: Pagination }> {
	const { query, withCredentials = true } = opts;
	const url = query ? withQuery(path, query) : path;
	// Multipart uploads take the top-level-FormData proxy; the serializer
	// cannot carry a FormData nested inside the JSON payload.
	const body = init?.body;
	if (body instanceof FormData) {
		body.set("__path", url);
		body.set("__credentials", withCredentials ? "1" : "0");
		return parseEnvelope<T>(await callApiUpload({ data: body }));
	}
	return parseEnvelope<T>(
		await callApiServer({
			data: {
				path: url,
				method: init?.method ?? "GET",
				body: typeof body === "string" ? body : null,
				withCredentials,
			},
		}),
	);
}

/** Per-request options: `query` appends URL params; `withCredentials`
 *  (default true) attaches the session Bearer — pass `false` for public
 *  endpoints that must not carry the user's credentials. */
export interface ApiOptions {
	query?: Record<string, string | number>;
	withCredentials?: boolean;
}

/** Non-paginated request (single object, create, update, delete). */
export async function api<T>(
	path: string,
	init?: RequestInit,
	opts?: ApiOptions,
): Promise<T> {
	const { data } = await request<T>(path, init, opts);
	return data;
}

/** Paginated list request. Throws if the response has no pagination meta. */
export async function apiPage<T>(
	path: string,
	init?: RequestInit,
	opts?: ApiOptions,
): Promise<{ items: T[]; pagination: Pagination }> {
	const { data, pagination } = await request<T[]>(path, init, opts);
	if (!pagination) {
		throw new ApiError(500, "Expected a paginated response");
	}
	return { items: data, pagination };
}
