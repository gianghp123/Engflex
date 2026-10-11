import { expect, test } from "@playwright/test";

/**
 * Same-origin voice signaling contract (2b). The browser POSTs START/OFFER
 * to the web origin; Server Routes forward to the Go API byte-verbatim.
 * These request-level specs pin the route surface without needing a voice
 * engine: unauthenticated callers must get a JSON error, never a 404
 * (missing route) or an HTML error page (wrong layer).
 */
test("POST /api/voice/:id/start without session is a JSON 401", async ({
	request,
}) => {
	const res = await request.post("/api/voice/no-such-conv/start", {
		data: {},
	});
	expect(res.status()).toBe(401);
	const body = (await res.json()) as { message?: unknown };
	expect(typeof body.message).toBe("string");
});

test("POST /api/voice/:id/offer without session is a JSON 401", async ({
	request,
}) => {
	const res = await request.post("/api/voice/no-such-conv/offer", {
		data: { sdp: "v=0", type: "offer" },
	});
	expect(res.status()).toBe(401);
	const body = (await res.json()) as { message?: unknown };
	expect(typeof body.message).toBe("string");
});

test("PATCH /api/voice/:id/offer without session is a JSON 401", async ({
	request,
}) => {
	const res = await request.patch("/api/voice/no-such-conv/offer", {
		data: { candidates: [] },
	});
	expect(res.status()).toBe(401);
	const body = (await res.json()) as { message?: unknown };
	expect(typeof body.message).toBe("string");
});
