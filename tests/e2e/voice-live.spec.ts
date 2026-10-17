import { expect, test } from "@playwright/test";
import { ensureFakeMicAudio } from "./fake-mic";
import {
	apiCreateConversation,
	closeDb,
	dbDeleteConversation,
	dbLearnerTurns,
	dbListTurns,
} from "./helpers";

/**
 * Live flow against a real engine (see playwright.live.config.ts): the fake
 * mic really streams through WebRTC -> Deepgram -> collector -> live turn post
 * -> Go poll -> Analyze -> engine /analyze -> diagnostics card.
 *
 * Everything asserted here is something only a live run can prove. The
 * position invariant under test: the Analyze button on the user's live row
 * resolves to the persisted turn at the *same* ordinal. A mismatch surfaces as
 * a 400 ("only learner turns can be analyzed") or as feedback on the wrong
 * turn — the DB assertions below catch both.
 */

/** Anchor word from the fake-mic sentence ("Well, I go to the office
 *  yesterday"). Matched loosely: streaming STT returns lowercase unpunctuated
 *  text, and a mangled fragment can precede the clean transcription — so both
 *  the UI and the DB look for the first turn *containing* this word, not the
 *  first turn. */
const SPOKEN_KEYWORD = /yesterday/i;

// Worst case is phase-driven, not latency-driven: the 29 s file loop can
// start up to ~20 s before the session exists, so the first in-session
// instance may be a full loop period away — and the proxied LLM endpoint's
// latency varies wildly run to run (15 s vs 60 s+ for the same greeting).
// Generous waits + a per-test timeout to match (the config default of 180 s
// would kill a worst-case run).
test.setTimeout(480_000);

const created: string[] = [];

test.beforeAll(() => {
	ensureFakeMicAudio();
});

test.beforeEach(async ({ context }) => {
	// A browser's first LLM action opens the explanation-language pick, so
	// start every test already asked and let Analyze/Check run straight through.
	await context.addInitScript(() => {
		for (const surface of ["vocabulary", "writing", "analyze"]) {
			window.localStorage.setItem(
				`engflex:explanation-language:asked:${surface}`,
				"1",
			);
		}
	});
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
	// Keep the conversation when the test fails: the boundary report
	// (live-boundaries.mjs) reads its turns/feedback from the DB, and
	// afterEach deletion would turn its db/analyze rows misleadingly red.
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

test("live session: fake-mic turn analyzes into a diagnostics card", async ({
	page,
}) => {
	const conv = await apiCreateConversation({ mode: "free_talk" });
	created.push(conv.id);
	// Marker for live-boundaries.mjs: attributes engine log lines (which
	// carry no conversation id) to this session by time.
	console.log(`e2e: live conversation ${conv.id}`);

	await page.goto(`/voice/room/${conv.id}`);

	// The partner speaks first: session, LLM, TTS and the assistant
	// aggregator all work end to end. Generous: the proxied LLM endpoint
	// took 60 s+ for the greeting on one run vs ~15 s on others.
	const partnerRow = page
		.locator("li")
		.filter({ has: page.getByText("Partner", { exact: true }) })
		.first();
	await expect(partnerRow).toBeVisible({ timeout: 120_000 });

	// The fake mic is heard and transcribed: STT, the collector merge, the
	// fire-and-forget turn post and the 2s poll all work. Waits for a row
	// containing the anchor word (up to several loop periods: playback phase
	// vs. session start is uncontrolled) rather than the first user row,
	// which may hold a mangled streaming fragment.
	const userRow = page
		.locator("li")
		.filter({ hasText: SPOKEN_KEYWORD })
		.first();
	await expect(userRow).toBeVisible({ timeout: 150_000 });

	// The row's ordinal among its siblings is the group ordinal, which the
	// panel uses as the turn position for the Analyze call. Recording it
	// lets the DB assertions verify the position invariant directly instead
	// of assuming the greeting is position 1 (playback phase decides that).
	// Counted in-page: chaining an xpath locator here trips Playwright's
	// selector parser ("Unexpected token ':' ... parsed as css").
	const groupOrdinal = await userRow.evaluate((el) => {
		let ordinal = 1;
		let sibling = el.previousElementSibling;
		while (sibling) {
			ordinal += 1;
			sibling = sibling.previousElementSibling;
		}
		return ordinal;
	});

	// The button appears only once the position exists server-side, so its
	// presence also proves the turn post landed.
	const analyze = userRow.getByRole("button", { name: "Analyze turn" });
	await expect(analyze).toBeVisible({ timeout: 30_000 });
	await analyze.click();

	// A real analysis pass (one LLM round trip through the engine). The result
	// opens as a dialog, so assertions below are page-scoped, not row-scoped.
	await expect(page.getByText("Analyzed turn")).toBeVisible({
		timeout: 90_000,
	});
	await expect(
		page.getByRole("button", { name: "Analyze again" }),
	).toBeVisible();
	await expect(page.getByText("Analysis failed. Try again.")).toHaveCount(0);

	// Ground truth for the position invariant: the clicked row was the
	// Nth group in the panel, so its turn must sit at DB position N, and the
	// feedback must be keyed by that turn's UUID. No assumption about the
	// greeting being position 1 — playback phase decides the interleaving.
	const turns = await dbListTurns(conv.id);
	expect(
		turns.some((turn) => turn.role === "ai"),
		`no greeting persisted: ${JSON.stringify(turns)}`,
	).toBe(true);

	const learners = await dbLearnerTurns(conv.id);
	const analyzed = learners.find((learner) =>
		SPOKEN_KEYWORD.test(learner.turn.text),
	);
	expect(
		analyzed,
		`no clean learner turn: ${JSON.stringify(turns)}`,
	).toBeTruthy();
	expect(
		analyzed?.turn.position,
		`group ordinal ${groupOrdinal} != persisted position`,
	).toBe(groupOrdinal);
	expect(
		analyzed?.feedback,
		"analyze never reached the clean turn",
	).not.toBeNull();
	expect(analyzed?.feedback?.corrected.length ?? 0).toBeGreaterThan(0);
	expect(analyzed?.feedback?.tip.length ?? 0).toBeGreaterThan(0);
});
