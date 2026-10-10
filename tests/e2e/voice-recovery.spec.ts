import fs from "node:fs";
import path from "node:path";
import { expect, test } from "@playwright/test";
import { ensureFakeMicAudio } from "./fake-mic";
import {
	apiCreateConversation,
	closeDb,
	dbDeleteConversation,
	dbListTurns,
} from "./helpers";

/**
 * On-demand transcript correction against a real engine, with nothing forced:
 * the learner opens the modal and every assertion follows from that.
 *
 * What only this run can prove, end to end:
 *  phase 1 (nothing intercepts): the fake-mic turn commits on its own and NO
 *    dialog appears. No signal interrupts the learner mid-sentence, and that is
 *    only observable end to end.
 *  phase 2 (edit): Review transcript opens a modal holding the ENGINE's copy of
 *    the turn; an edit rewrites that turn *in place* at the same position (DB
 *    ground truth) and the tutor regenerates a reply. The panel redraws to
 *    match, without anything being written to pipecat's own list.
 *  phase 3 (record): Record captures a re-speak off-pipeline and its
 *    transcription becomes the field; the button survives for a SECOND
 *    attempt, and sending corrects the reviewed turn IN PLACE — same
 *    position, no appended turn — and gets a reply. The live mic is muted
 *    while the modal is open, so nothing short of this run proves the
 *    pipeline hears silence behind it.
 */

const EDITED_MARKER = "e2e correction edited sentence";
const RAW_ANCHOR = /yesterday/i;
const REVIEW = "Review transcript";
// Frozen clip under fixtures/, verified against the real endpoint: this is
// what Deepgram hears in it, byte for byte.
const FIXTURE_SENTENCE = "Well, I went to the market yesterday.";

test.setTimeout(240_000);

const created: string[] = [];

test.beforeAll(() => {
	ensureFakeMicAudio();
});

test.beforeEach(async ({ context }) => {
	await context.addCookies([
		{
			name: "PARAGLIDE_LOCALE",
			value: "en",
			domain: "localhost",
			path: "/",
		},
	]);
});

test.afterEach(async ({ browser }, testInfo) => {
	// `browser` is destructured because Playwright requires the object
	// pattern here; the hook only needs testInfo.
	void browser;
	if (testInfo.status !== testInfo.expectedStatus) {
		console.log(`e2e: keeping failed conversation(s) ${created.join(",")}`);
		created.splice(0);
		return;
	}
	await Promise.all(created.splice(0).map((id) => dbDeleteConversation(id)));
});

test.afterAll(async () => {
	await closeDb();
});

/** Every transcript bubble, in order, and nothing else on the page.
 *
 * `page.locator("li")` also matches the sidebar's navigation items, so scoping
 * to the two speaker roles is what makes a bubble count comparable with a turn
 * count.
 */
async function bubbleTexts(page: import("@playwright/test").Page) {
	const [learner, partner] = await Promise.all([
		learnerBubbles(page).allInnerTexts(),
		partnerBubbles(page).allInnerTexts(),
	]);
	// Panel order interleaves the two roles, so sort by position in the document
	// rather than concatenating.
	const rows = await page
		.locator("li")
		.filter({ hasText: /^(P|Y)\s(Partner|You)\b/ })
		.allInnerTexts();
	return rows.length ? rows : [...partner, ...learner];
}

/** Learner bubbles in the live transcript panel. */
function learnerBubbles(page: import("@playwright/test").Page) {
	return page
		.locator("li")
		.filter({ has: page.getByText("You", { exact: true }) });
}

/** Tutor bubbles in the live transcript panel. */
function partnerBubbles(page: import("@playwright/test").Page) {
	return page
		.locator("li")
		.filter({ has: page.getByText("Partner", { exact: true }) });
}

async function userTurns(conversationId: string) {
	return (await dbListTurns(conversationId)).filter((t) => t.role === "user");
}

async function aiTurnCount(conversationId: string) {
	return (await dbListTurns(conversationId)).filter((t) => t.role === "ai")
		.length;
}

/**
 * Wait until the conversation stops growing.
 *
 * The tutor can still be replying to a turn when the learner opens the modal,
 * and that reply lands legitimately — persisted, but arriving while the modal
 * is open. Without this, every "nothing changed while the modal was open"
 * assertion races against it and reports a disturbance that is not one.
 */
async function waitForQuiet(
	page: import("@playwright/test").Page,
	conversationId: string,
	{ stableReads = 2, gapMs = 2000, maxMs = 30_000 } = {},
) {
	const deadline = Date.now() + maxMs;
	let previousDb = "";
	let previousPanel = "";
	let stable = 0;
	while (Date.now() < deadline) {
		const dbNow = (await dbListTurns(conversationId))
			.map((t) => `${t.position}:${t.role}:${t.text}`)
			.join("\n");
		const panelNow = (await bubbleTexts(page))
			.map((t) => t.replace(/\s+/g, " ").trim())
			.join("\n");
		// Full text, not counts. A streaming tutor reply grows a bubble's TEXT
		// without adding bubbles or turns, and snapshotting mid-stream makes
		// the rest of the reply look like a disturbance. Quiescence means the
		// words stopped changing on both sides.
		const quiet = dbNow === previousDb && panelNow === previousPanel;
		stable = quiet ? stable + 1 : 0;
		previousDb = dbNow;
		previousPanel = panelNow;
		if (stable >= stableReads) return;
		await new Promise((r) => setTimeout(r, gapMs));
	}
	throw new Error(
		`e2e: conversation never settled within ${maxMs}ms ` +
			`(db ${(await dbListTurns(conversationId)).length} turns, panel ${(await bubbleTexts(page)).length} bubbles)`,
	);
}

async function waitFor(
	read: () => Promise<boolean>,
	what: string,
	timeoutMs: number,
): Promise<void> {
	const deadline = Date.now() + timeoutMs;
	for (;;) {
		if (await read()) return;
		if (Date.now() > deadline) {
			throw new Error(`e2e: timed out waiting for ${what}`);
		}
		await new Promise((r) => setTimeout(r, 2000));
	}
}

test("a misheard turn is corrected on demand, and a re-take is answered", async ({
	page,
}) => {
	// The command rides HTTP and the capture rides the RTVI data channel, so a
	// silent client is the most likely failure. Forward the browser's own
	// account of what it received instead of guessing from the DOM.
	page.on("console", (msg) =>
		console.log(`[browser:${msg.type()}] ${msg.text()}`),
	);
	page.on("pageerror", (err) =>
		console.log(`[browser:pageerror] ${err.message}`),
	);

	const conv = await apiCreateConversation({ mode: "free_talk" });
	created.push(conv.id);
	console.log(`e2e: correction conversation ${conv.id}`);

	await page.goto(`/voice/room/${conv.id}`);

	// Session is live: the partner speaks first (copies the live suite's
	// generous greeting budget — the proxied LLM endpoint varies wildly).
	const partnerRows = partnerBubbles(page);
	await expect(partnerRows.first()).toBeVisible({ timeout: 60_000 });
	const greetingCount = await aiTurnCount(conv.id);

	// Phase 1 — nothing intercepts. The first fake-mic instance commits as an
	// ordinary turn and NO dialog appears: no signal can interrupt the learner
	// mid-sentence, and only this run can see it.
	await waitFor(
		async () => (await userTurns(conv.id)).length > 0,
		"the first learner turn to commit",
		180_000,
	);
	await expect(page.getByRole("dialog")).toHaveCount(0, { timeout: 15_000 });

	// Phase 2 — the learner opens the modal and corrects the turn. Stop talking
	// and let the tutor finish first: a reply still in flight would land inside
	// the modal and look like the modal disturbing the conversation.
	await waitForQuiet(page, conv.id);
	const review = page.getByRole("button", { name: REVIEW, exact: true });
	await expect(review).toBeEnabled();
	await review.click();
	const dialog = page.getByRole("dialog");
	await expect(dialog).toBeVisible({ timeout: 30_000 });

	// A strict snapshot of EVERY message the modal must not disturb: the full
	// text of every panel bubble and every persisted turn, not counts.
	//
	// Counts are not enough, and this suite has been fooled by them. A leaked
	// utterance can MERGE into the learner turn already on screen — the bubble
	// count stays put while its text grows to "sentence sentence sentence". The
	// tutor answering speech made inside the modal is a second thing counts miss
	// entirely, because it adds an assistant bubble rather than a learner one.
	// Comparing whole message state catches both, plus any silent replacement.
	const snapshot = async () => ({
		panel: (await bubbleTexts(page)).map((t) => t.replace(/\s+/g, " ").trim()),
		db: (await dbListTurns(conv.id)).map(
			(t) => `${t.position}:${t.role}:${t.text}`,
		),
	});
	const describeState = async () => JSON.stringify(await snapshot(), null, 1);
	const expectUndisturbed = async (
		before: Awaited<ReturnType<typeof snapshot>>,
		when: string,
	) => {
		const now = await snapshot();
		expect(
			now,
			`the modal disturbed the conversation ${when}: ${await describeState()}`,
		).toEqual(before);
	};

	// Snapshots are taken fresh where they are asserted: phase 3 snapshots after
	// its own modal opens (see beforeRecord), so a reply still landing from an
	// earlier phase is never mistaken for a disturbance.

	// The modal is modal for a reason: the client mutes the learner's mic track.
	// It must NOT also pause the tutor's playback — Radix's focus
	// trap can suspend media elements it contains, and if it did, the learner
	// would be correcting in silence while the bot is mid-sentence.
	const audioState = await page.evaluate(() =>
		[...document.querySelectorAll("audio")].map((a) => ({
			paused: a.paused,
			muted: a.muted,
		})),
	);
	expect(
		audioState.filter((a) => a.paused),
		`the focus trap paused the bot's audio: ${JSON.stringify(audioState)}`,
	).toEqual([]);
	// The modal holds the engine's own copy, which is what Send will rewrite.
	const textarea = dialog.getByRole("textbox");
	await expect(textarea).toBeVisible();
	const shown = await textarea.inputValue();
	expect(shown.length, "the modal must show the transcript").toBeGreaterThan(0);

	// The correction rewrites a turn IN PLACE, so remember which position the
	// learner turn occupies. It is the LATEST learner turn, because that is what
	// Review resolves to — and it is not 1: the greeting is recorded first, and
	// the looping fake mic commits more than one learner turn.
	const learnerTurns = await userTurns(conv.id);
	const targetPosition = Math.max(...learnerTurns.map((t) => t.position));
	expect(learnerTurns.length, "no learner turn to review").toBeGreaterThan(0);

	await textarea.fill(EDITED_MARKER);
	await dialog
		.getByRole("button", { name: "Send message", exact: true })
		.click();
	await expect(dialog).toBeHidden({ timeout: 30_000 });

	// Phase 2b — the panel is a derivation, not a mirror of pipecat. Nothing is
	// written back to pipecat's own list; the reviewed turn is redrawn with the
	// corrected words and the reply that answered the old ones is dropped. So
	// the corrected sentence must be on screen, and the learner turn count must
	// NOT have grown — a correction replaces, it never appends.
	const learnerRows = learnerBubbles(page);
	const learnerCountBefore = await learnerRows.count();
	await expect(page.getByText(EDITED_MARKER, { exact: true })).toBeVisible({
		timeout: 30_000,
	});
	expect(
		await learnerRows.count(),
		`the correction appended a learner turn instead of replacing one: ${await describeState()}`,
	).toBe(learnerCountBefore);
	// The newest learner bubble must BE the corrected sentence, not merely be
	// followed by one. ("The old words are gone" is not assertable here: the fake
	// mic loops one fixed phrase, so the same sentence recurs in later turns.)
	const newestAfterEdit = (await learnerRows.allInnerTexts()).at(-1) ?? "";
	expect(
		newestAfterEdit,
		`the newest learner bubble is not the correction: ${await describeState()}`,
	).toContain(EDITED_MARKER);

	// DB ground truth: that same position now carries the edited text. Later
	// positions are raw loop audio and are supposed to commit.
	await waitFor(
		async () => {
			const turn = (await dbListTurns(conv.id)).find(
				(t) => t.position === targetPosition,
			);
			return turn?.text === EDITED_MARKER;
		},
		`the corrected text at position ${targetPosition}`,
		60_000,
	);
	const corrected = (await dbListTurns(conv.id)).find(
		(t) => t.position === targetPosition,
	);
	expect(corrected?.text).toBe(EDITED_MARKER);
	expect(RAW_ANCHOR.test(corrected?.text ?? "")).toBe(false);

	// The correction regenerates a reply.
	await expect
		.poll(async () => aiTurnCount(conv.id), { timeout: 60_000 })
		.toBeGreaterThan(greetingCount);

	// Phase 3 — isolated retake. The re-speak is recorded off-pipeline and
	// transcribed into the field: each attempt replaces it, and Send corrects
	// the reviewed turn in place. An attempt is never a new turn, however many
	// the learner makes.
	await waitForQuiet(page, conv.id);
	await review.click();
	await expect(dialog).toBeVisible({ timeout: 30_000 });
	// Fresh snapshot: phase 2's correction regenerated a reply that may still
	// have been landing when phase 2 ended. The modal is open and the mic is
	// muted, so once quiet, nothing else can legitimately move.
	await waitForQuiet(page, conv.id);
	const beforeRecord = await snapshot();
	const field = dialog.getByRole("textbox");

	// The fixture clip stands in for the microphone: MediaRecorder would only
	// capture the fake device's silence in CI. The bytes travel the real path
	// — hook seam, upload, engine, Deepgram — and only the capture is faked.
	const fixture = fs.readFileSync(
		path.join(import.meta.dirname, "fixtures", "retake.wav"),
	);
	await page.evaluate(
		(bytes: number[]) => {
			(
				window as unknown as { __E2E_RETAKE_BYTES__?: ArrayBuffer }
			).__E2E_RETAKE_BYTES__ = new Uint8Array(bytes).buffer;
		},
		[...fixture],
	);
	await dialog.getByRole("button", { name: "Re-record", exact: true }).click();
	await expect(dialog.getByText("Recording")).toBeVisible({ timeout: 30_000 });
	await dialog.getByRole("button", { name: "Stop", exact: true }).click();

	// The utterance reached the field. It is the fixture sentence — a real
	// transcription, not the misheard turn — so typing and speaking demonstrably
	// fill one field with content that differs from what was reviewed. Polled,
	// because the upload round-trips through Go, the engine, and Deepgram.
	await expect
		.poll(async () => field.inputValue(), { timeout: 60_000 })
		.toBe(FIXTURE_SENTENCE);
	const spoken = await field.inputValue();

	// A second attempt. The button is still there: the number of tries is not
	// capped, and only a live run proves the endpoint accepts a second one.
	await dialog.getByRole("button", { name: "Re-record", exact: true }).click();
	await dialog.getByRole("button", { name: "Stop", exact: true }).click();
	await expect
		.poll(async () => field.inputValue(), { timeout: 60_000 })
		.toBe(FIXTURE_SENTENCE);

	// The utterance reached the field. Nothing else may have moved: no learner
	// bubble, no tutor reply, no persisted turn. A reply to the transcribed
	// re-speak would be the old leak wearing a new shape.
	await expectUndisturbed(beforeRecord, "after a recorded attempt");

	// A second attempt must also disturb nothing.
	await expectUndisturbed(beforeRecord, "after a second recorded attempt");

	// Read the transcript while the modal is open: nothing can commit
	// underneath us, so the position we are about to correct is stable.
	const beforeSend = await userTurns(conv.id);
	const target = Math.max(...beforeSend.map((t) => t.position));
	const aiBefore = await aiTurnCount(conv.id);

	await dialog
		.getByRole("button", { name: "Send message", exact: true })
		.click();
	await expect(dialog).toBeHidden({ timeout: 30_000 });

	// In place: the reviewed turn now carries what was spoken, and the number of
	// learner turns is unchanged.
	await waitFor(
		async () => {
			const turn = (await dbListTurns(conv.id)).find(
				(t) => t.position === target,
			);
			return turn?.text === spoken;
		},
		`the spoken correction at position ${target}`,
		90_000,
	);
	const afterSend = await userTurns(conv.id);
	expect(
		afterSend.length,
		`a spoken attempt appended a turn instead of correcting one: ${JSON.stringify(
			afterSend.map((t) => t.text),
		)}`,
	).toBe(beforeSend.length);
	expect(afterSend.find((t) => t.position === target)?.text).toBe(spoken);

	// The panel must show the replacement, not just the database. A green DB
	// assertion with a stale panel is exactly the failure this suite exists to
	// catch, so check the rendered bubble text directly.
	await expect(
		page.getByText(spoken, { exact: true }),
		`the spoken sentence never reached the panel: ${await describeState()}`,
	).toBeVisible({ timeout: 30_000 });
	// NOTE: the fixture sentence genuinely differs from the turn it corrects,
	// so this phase proves "the new words replaced the old ones" twice over —
	// once typed in phase 2, once spoken here. What is proven HERE is that the
	// panel shows the spoken attempt, and that it is a replacement rather than
	// an addition.
	const newestLearner =
		(await learnerBubbles(page).allInnerTexts()).at(-1) ?? "";
	expect(
		newestLearner,
		`the newest learner bubble is not the spoken sentence: ${await describeState()}`,
	).toContain(spoken);

	// The tutor MUST answer the corrected sentence — silence here would be just as
	// broken as answering too early. The modal is closed by now, so a reply is
	// the correct outcome.
	await expect
		.poll(async () => aiTurnCount(conv.id), { timeout: 60_000 })
		.toBeGreaterThan(aiBefore);
	// And it must reach the panel, not only the database.
	await expect
		.poll(
			async () => (await partnerBubbles(page).allInnerTexts()).join(" ").length,
			{ timeout: 60_000 },
		)
		.toBeGreaterThan(0);
});

test("dismissing the modal hands the microphone back", async ({ page }) => {
	const conv = await apiCreateConversation({ mode: "free_talk" });
	created.push(conv.id);
	await page.goto(`/voice/room/${conv.id}`);

	await expect(
		page
			.locator("li")
			.filter({ has: page.getByText("Partner", { exact: true }) })
			.first(),
	).toBeVisible({ timeout: 60_000 });
	await waitFor(
		async () => (await userTurns(conv.id)).length > 0,
		"the first learner turn to commit",
		180_000,
	);

	// Snapshot the whole transcript, so "dismissing recorded nothing" is a real
	// invariant rather than a count comparison.
	const before = await dbListTurns(conv.id);
	await page.getByRole("button", { name: REVIEW, exact: true }).click();
	const dialog = page.getByRole("dialog");
	await expect(dialog).toBeVisible({ timeout: 30_000 });
	await dialog.getByRole("button", { name: "Cancel", exact: true }).click();
	await expect(dialog).toBeHidden({ timeout: 30_000 });

	// Deliberately NOT asserting that another turn commits: the fixture loops
	// with 60 s of lead silence and the tutor may still be speaking, so barge-in
	// can swallow the next utterance (the "Barge-in is normal" gotcha). What
	// must hold is that dismissing changed nothing that already existed, and
	// that the session is still usable.
	const after = await dbListTurns(conv.id);
	for (const turn of before) {
		const still = after.find((t) => t.position === turn.position);
		expect(still?.text, `dismissing changed position ${turn.position}`).toBe(
			turn.text,
		);
	}
	await expect(
		page.getByRole("button", { name: REVIEW, exact: true }),
	).toBeEnabled();

	// The mic came back: with the loopback fixture still talking, a muted mic
	// could never commit again, so growth here is the positive proof that
	// dismissing restored it. Settle first — audio already in flight when the
	// modal closed may still land once, and that is not proof of anything.
	await waitForQuiet(page, conv.id);
	const resettled = await userTurns(conv.id);
	await waitFor(
		async () => (await userTurns(conv.id)).length > resettled.length,
		"no turn committed after dismiss — the mic stayed muted",
		180_000,
	);
});

test("every learner turn keeps analyze, and the latest keeps edit across re-edits", async ({
	page,
}) => {
	const conv = await apiCreateConversation({ mode: "free_talk" });
	created.push(conv.id);
	await page.goto(`/voice/room/${conv.id}`);

	await expect(
		page
			.locator("li")
			.filter({ has: page.getByText("Partner", { exact: true }) })
			.first(),
	).toBeVisible({ timeout: 60_000 });
	// Two learner turns, so "all" and "previous" mean something.
	await waitFor(
		async () => (await userTurns(conv.id)).length >= 2,
		"two learner turns to commit",
		180_000,
	);

	const ANALYZE = "Analyze turn";
	const expectButtons = async (when: string) => {
		await waitForQuiet(page, conv.id);
		const bubbles = learnerBubbles(page);
		const count = await bubbles.count();
		expect(count, `fewer than two learner bubbles ${when}`).toBeGreaterThan(1);
		for (let i = 0; i < count; i++) {
			const bubble = bubbles.nth(i);
			await expect(
				bubble.getByRole("button", { name: ANALYZE, exact: true }),
				`learner bubble ${i} has no analyze button ${when}`,
			).toBeVisible({ timeout: 15_000 });
		}
		const last = bubbles.nth(count - 1);
		await expect(
			last.getByRole("button", { name: REVIEW, exact: true }),
			`latest learner bubble has no edit button ${when}`,
		).toBeVisible({ timeout: 15_000 });
	};

	await expectButtons("before any edit");

	// Edit the latest turn twice: re-editing must stay possible, and the
	// buttons must survive both the correction and its reply.
	const dialog = page.getByRole("dialog");
	for (const marker of ["e2e buttons first edit", "e2e buttons second edit"]) {
		await page.getByRole("button", { name: REVIEW, exact: true }).click();
		await expect(dialog).toBeVisible({ timeout: 30_000 });
		const field = dialog.getByRole("textbox");
		await field.fill(marker);
		await dialog
			.getByRole("button", { name: "Send message", exact: true })
			.click();
		await expect(dialog).toBeHidden({ timeout: 30_000 });
		await expectButtons(`after sending "${marker}"`);
	}

	// Each send replaced instead of appending: the second marker exists at
	// exactly one user position (loopback chatter never carries it).
	const turns = await userTurns(conv.id);
	expect(
		turns.filter((t) => t.text === "e2e buttons second edit"),
		"the second edit did not land in place exactly once",
	).toHaveLength(1);
});
