import { expect, test } from "@playwright/test";
import { ensureFakeMicAudio } from "./fake-mic";
import { apiCreateConversation, closeDb, dbDeleteConversation } from "./helpers";

/**
 * Admission control end to end, with a dispatcher capped at one session.
 *
 * Only a live run can prove the whole chain: a second conversation's start
 * is refused by the ENGINE (429), surfaced by Go (503), rendered by the room
 * (the same blocking modal a dead engine produces) — and closing the live
 * session's page really does hand the slot back, because aiortc noticed the
 * dropped peer and the dispatcher released the lease.
 */

const ENGINE_PORT = Number(process.env.E2E_LIVE_ENGINE_PORT ?? 7861);
const created: string[] = [];

test.setTimeout(240_000);

test.beforeAll(() => {
	ensureFakeMicAudio();
});

test.beforeEach(async ({ context }) => {
	await context.addCookies([
		{ name: "PARAGLIDE_LOCALE", value: "en", domain: "localhost", path: "/" },
	]);
});

test.afterEach(async () => {
	await Promise.all(created.splice(0).map((id) => dbDeleteConversation(id)));
});

test.afterAll(async () => {
	await closeDb();
});

function partnerRow(page: import("@playwright/test").Page) {
	return page
		.locator("li")
		.filter({ has: page.getByText("Partner", { exact: true }) })
		.first();
}

/** Engine-side counters: the dispatcher's own view of its capacity. */
async function engineStatus(): Promise<{ active: number; pending: number }> {
	const res = await fetch(`http://localhost:${ENGINE_PORT}/status`);
	return (await res.json()) as { active: number; pending: number };
}

test("capacity: the (max+1)th start is refused until a session ends", async ({
	context,
}) => {
	// Two pages, one context: the live session must stay connected while the
	// second one proves the refusal — navigating one page would close the
	// other's peer connection and free the slot mid-assertion.
	const live = await apiCreateConversation({ mode: "free_talk" });
	created.push(live.id);
	const pageA = await context.newPage();
	await pageA.goto(`/voice/room/${live.id}`);
	await expect(partnerRow(pageA)).toBeVisible({ timeout: 120_000 });
	expect((await engineStatus()).active).toBe(1);

	// The single slot is taken. A second conversation cannot start.
	const blocked = await apiCreateConversation({ mode: "free_talk" });
	created.push(blocked.id);
	const pageB = await context.newPage();
	await pageB.goto(`/voice/room/${blocked.id}`);
	const dialog = pageB.getByRole("alertdialog");
	await expect(dialog).toBeVisible({ timeout: 30_000 });
	await expect(dialog).toContainText("Connection failed. Please try again.");

	// Teardown frees the slot: closing the live page drops the peer
	// connection, the pipeline ends, and the dispatcher releases the lease.
	await pageA.close();
	const deadline = Date.now() + 60_000;
	for (;;) {
		const status = await engineStatus();
		if (status.active === 0) break;
		if (Date.now() > deadline) {
			throw new Error(`e2e: the live slot never freed (active=${status.active})`);
		}
		await new Promise((r) => setTimeout(r, 2000));
	}

	// The blocked conversation starts cleanly once capacity returned.
	let started = false;
	for (let attempt = 0; attempt < 8 && !started; attempt++) {
		await pageB.goto(`/voice/room/${blocked.id}`);
		try {
			await expect(partnerRow(pageB)).toBeVisible({ timeout: 30_000 });
			started = true;
		} catch {
			await pageB.waitForTimeout(5_000);
		}
	}
	expect(started, "the freed slot never admitted the blocked conversation").toBe(true);
});
