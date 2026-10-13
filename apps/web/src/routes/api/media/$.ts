import { createFileRoute } from "@tanstack/react-router";

const MINIO_INTERNAL =
	process.env.MEDIA_INTERNAL_ENDPOINT ?? "http://localhost:9000";
const MEDIA_BUCKET = process.env.MEDIA_BUCKET ?? "lesson-audio";

/**
 * Same-origin media forward for every lesson-media object (audio today,
 * any content-type tomorrow — nothing here is audio-specific). The bucket
 * is public-read and fixed server-side: only the object key travels in the
 * URL, so callers can never steer this at another bucket. Range requests
 * are forwarded and 206s passed through verbatim (`<audio>`/`<video>`
 * seeking breaks otherwise). No auth: posture identical to today's
 * anonymous bucket reads, keys are unguessable UUIDs, pages stay gated.
 */
export const Route = createFileRoute("/api/media/$")({
	server: {
		handlers: {
			GET: async ({ request, params }) => {
				const key = (params._splat ?? "").replace(/^\/+/, "");
				if (!key || key.includes("..")) {
					return Response.json({ message: "bad media key" }, { status: 404 });
				}
				const object = key.split("/").map(encodeURIComponent).join("/");
				const range = request.headers.get("range");
				let upstream: Response;
				try {
					upstream = await fetch(
						`${MINIO_INTERNAL}/${MEDIA_BUCKET}/${object}`,
						{
							headers: range ? { range } : {},
						},
					);
				} catch {
					return Response.json(
						{ message: "media unavailable" },
						{ status: 502 },
					);
				}
				if (upstream.status === 404) {
					return Response.json({ message: "media not found" }, { status: 404 });
				}
				if (!upstream.ok) {
					return Response.json(
						{ message: "media unavailable" },
						{ status: 502 },
					);
				}
				const bytes = await upstream.arrayBuffer();
				const headers: Record<string, string> = {};
				for (const h of [
					"content-type",
					"content-length",
					"content-range",
					"accept-ranges",
					"etag",
					"last-modified",
				]) {
					const v = upstream.headers.get(h);
					if (v) headers[h] = v;
				}
				return new Response(bytes, { status: upstream.status, headers });
			},
		},
	},
});
