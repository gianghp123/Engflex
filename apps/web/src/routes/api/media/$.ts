import { createFileRoute } from "@tanstack/react-router";
import { isToeicMediaKey } from "#/features/exams/media-key";

const MINIO_INTERNAL =
	process.env.MEDIA_INTERNAL_ENDPOINT?.trim() || "http://localhost:9000";
const MEDIA_BUCKET = process.env.MEDIA_BUCKET?.trim() || "lesson-audio";

/**
 * Same-origin media forward for every lesson-media object (audio today,
 * any content-type tomorrow — nothing here is audio-specific). Lesson media
 * uses MEDIA_BUCKET. Story objects use the stories bucket. The fixed toeic bucket exposes only its image/audio
 * folders; full exam JSON stays private. Range requests
 * are forwarded and 206s passed through verbatim (`<audio>`/`<video>`
 * seeking breaks otherwise). No auth: posture identical to today's
 * anonymous media reads; pages stay gated.
 */
export const Route = createFileRoute("/api/media/$")({
	server: {
		handlers: {
			GET: async ({ request, params }) => {
				const key = (params._splat ?? "").replace(/^\/+/, "");
				if (!key || key.includes("..")) {
					return Response.json({ message: "bad media key" }, { status: 404 });
				}
				const toeic = key.startsWith("toeic/");
				if (toeic && !isToeicMediaKey(key)) {
					return Response.json({ message: "bad media key" }, { status: 404 });
				}
				const stories = key.startsWith("stories/");
				const bucket = toeic ? "toeic" : stories ? "stories" : MEDIA_BUCKET;
				const objectKey = toeic || stories ? key.slice(bucket.length + 1) : key;
				const object = objectKey.split("/").map(encodeURIComponent).join("/");
				const range = request.headers.get("range");
				let upstream: Response;
				try {
					upstream = await fetch(`${MINIO_INTERNAL}/${bucket}/${object}`, {
						headers: range ? { range } : {},
					});
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
				return new Response(upstream.body, {
					status: upstream.status,
					headers,
				});
			},
		},
	},
});
