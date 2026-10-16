import { expect, test } from "@playwright/test";
import {
	apiCreateConversation,
	apiEndConversation,
	clerkUserId,
	closeDb,
	dbConversationExists,
	dbDeleteConversation,
	dbDeleteScenario,
	dbFailConversation,
	dbInsertScenario,
	dbSeedFeedback,
	dbSeedTurn,
} from "./helpers";

/** Seeded catalog rows (20261005193724_seed_voice_browser.sql). */
const SCENARIO_ID = "7c000001-0000-4000-8000-000000000001"; // "Ordering coffee to go" (Maya)
const TOPIC_ID = "7a000001-0000-4000-8000-000000000001";

const created: string[] = [];
const createdScenarios: string[] = [];

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

test.afterEach(async () => {
	await Promise.all(created.splice(0).map((id) => dbDeleteConversation(id)));
	await Promise.all(
		createdScenarios.splice(0).map((id) => dbDeleteScenario(id)),
	);
});

test.afterAll(async () => {
	await closeDb();
});

async function endedRoleplay(scenarioId: string): Promise<string> {
	const conv = await apiCreateConversation({ mode: "roleplay", scenarioId });
	await apiEndConversation(conv.id);
	created.push(conv.id);
	return conv.id;
}

async function endedFreeTalk(): Promise<string> {
	const conv = await apiCreateConversation({ mode: "free_talk" });
	await apiEndConversation(conv.id);
	created.push(conv.id);
	return conv.id;
}

test("list rows open the persisted review, without the survey", async ({
	page,
}) => {
	const roleplayId = await endedRoleplay(SCENARIO_ID);
	const turnId = await dbSeedTurn(roleplayId, {
		role: "user",
		text: "One oat latte to go, please",
	});
	const freeTalkId = await endedFreeTalk();
	await dbSeedTurn(freeTalkId, {
		role: "user",
		text: "Small talk while waiting for coffee",
	});
	await dbSeedFeedback(await clerkUserId(), turnId, {
		corrected: "One oat latte to go, please.",
		spans: [],
		relevance: { status: "relevant", reason: null },
		alternatives: { language: null, contextual: null },
		tip: "Nice ordering line — keep it short and friendly.",
	});

	await page.goto("/voice/history");
	const row = page.locator(`a[href="/voice/history/${roleplayId}"]`);
	await expect(row).toBeVisible({ timeout: 20_000 });
	await expect(row).toContainText("Ordering coffee to go");
	await expect(row).toContainText("Maya");
	// No engine finalize ran: durationSec is null on this row.
	await expect(row).not.toContainText(" min");
	// The free-talk session lists too, with its fallback title.
	const freeTalkRow = page.locator(`a[href="/voice/history/${freeTalkId}"]`);
	await expect(freeTalkRow).toContainText("Free talk");
	await expect(page.getByText("Today")).toBeVisible();

	await row.click();
	await expect(page).toHaveURL(new RegExp(`/voice/history/${roleplayId}$`));
	// Exact: the feedback card also renders the corrected sentence WITH a
	// trailing period, and a substring match would be ambiguous. The raw
	// utterance renders twice by design — transcript bubble plus the
	// feedback card's utterance line, same as the room review — so pin the
	// count instead of assuming uniqueness.
	const utterance = page.getByText("One oat latte to go, please", {
		exact: true,
	});
	await expect(utterance).toHaveCount(2);
	await expect(utterance.first()).toBeVisible();
	await expect(
		page.getByText("Nice ordering line — keep it short and friendly."),
	).toBeVisible();
	// Analyze stays available on a past session: stored feedback shows the
	// re-analyze affordance.
	await expect(
		page.getByRole("button", { name: "Analyze again" }),
	).toBeVisible();
	// History variant: the post-session survey is gone, "Just practiced" is not the label.
	await expect(page.getByText("Help Mo improve your experience")).toHaveCount(0);
	await expect(page.getByText("Just practiced")).toHaveCount(0);

	// Seeded feedback auto-opens the analysis dialog on mount (same as the
	// room review on reload); dismiss it before leaving the page.
	await page.keyboard.press("Escape");
	await page.getByRole("button", { name: "Back to history" }).click();
	await expect(page).toHaveURL(/\/voice\/history$/);
});

test("failed sessions list with an interrupted badge", async ({ page }) => {
	const conv = await apiCreateConversation({ mode: "free_talk" });
	created.push(conv.id);
	await dbFailConversation(conv.id);
	await dbSeedTurn(conv.id, {
		role: "user",
		text: "Hello, can you hear me",
	});

	await page.goto("/voice/history");
	const row = page.locator(`a[href="/voice/history/${conv.id}"]`);
	await expect(row).toBeVisible({ timeout: 20_000 });
	await expect(row).toContainText("Interrupted");
});

test("empty sessions are hidden from history, their review still renders", async ({
	page,
}) => {
	const conv = await apiCreateConversation({ mode: "free_talk" });
	created.push(conv.id);
	await apiEndConversation(conv.id);

	await page.goto("/voice/history");
	// Hidden from the list: the fetch landed (empty state renders) and the
	// history link never renders.
	await expect(page.getByText("No sessions yet")).toBeVisible({
		timeout: 20_000,
	});
	await expect(
		page.locator(`a[href="/voice/history/${conv.id}"]`),
	).toHaveCount(0);

	// …but the detail path is unchanged and still shows the honest review.
	await page.goto(`/voice/history/${conv.id}`);
	await expect(page.getByText("We didn't catch that.")).toBeVisible({
		timeout: 20_000,
	});
	await expect(
		page.getByRole("button", { name: "Back to history" }),
	).toBeVisible();
});

test("a deleted scenario falls back to the generic roleplay title", async ({
	page,
}) => {
	const scenarioId = "7c0ffff1-0000-4000-8000-000000000001";
	await dbInsertScenario(scenarioId, TOPIC_ID, "Temporary e2e scenario");
	createdScenarios.push(scenarioId);
	const convId = await endedRoleplay(scenarioId);
	await dbSeedTurn(convId, {
		role: "user",
		text: "A temporary run that survives its scenario",
	});
	// Delete it now: conversations.scenario_id goes NULL (ON DELETE SET NULL).
	// The afterEach cleanup delete is a harmless no-op if this line ran.
	await dbDeleteScenario(scenarioId);

	await page.goto("/voice/history");
	const row = page.locator(`a[href="/voice/history/${convId}"]`);
	await expect(row).toBeVisible({ timeout: 20_000 });
	await expect(row).toContainText("Roleplay session");
});

test("scenario runs scope the history page and clear again", async ({
	page,
}) => {
	const roleplayIds: string[] = [];
	for (let i = 0; i < 4; i += 1) {
		roleplayIds.push(await endedRoleplay(SCENARIO_ID));
	}
	for (const [i, id] of roleplayIds.entries()) {
		await dbSeedTurn(id, {
			role: "user",
			text: `Scoped run number ${i + 1}`,
		});
	}
	const freeTalkId = await endedFreeTalk();
	await dbSeedTurn(freeTalkId, {
		role: "user",
		text: "Free talk outside the scenario scope",
	});

	await page.goto(`/voice/scenarios/${SCENARIO_ID}`);
	await expect(page.getByText("Previous sparring runs")).toBeVisible({
		timeout: 20_000,
	});
	await page.getByRole("link", { name: "View all" }).click();
	await expect(page).toHaveURL(
		new RegExp(`/voice/history\\?scenarioId=${SCENARIO_ID}$`),
	);

	// Scoped: the free-talk row is gone, the chip names the scenario.
	await expect(page.getByText("Scenario: Ordering coffee to go")).toBeVisible();
	await expect(
		page.locator(`a[href="/voice/history/${freeTalkId}"]`),
	).toHaveCount(0);
	await expect(
		page.getByText("Ordering coffee to go").first(),
	).toBeVisible();

	// Cleared: the unfiltered list is back.
	await page.getByRole("button", { name: "Clear filter" }).click();
	await expect(page).toHaveURL(/\/voice\/history\??$/);
	await expect(
		page.locator(`a[href="/voice/history/${freeTalkId}"]`),
	).toBeVisible();
});

test("pending sessions never appear in history", async ({ page }) => {
	// Order matters: CreateClosingActive closes the caller's non-terminal
	// conversations on the NEXT create, so the ended session is created
	// first and the pending one last — and is never ended.
	const endedId = await endedFreeTalk();
	await dbSeedTurn(endedId, {
		role: "user",
		text: "A finished session that stays listed",
	});
	const pending = await apiCreateConversation({ mode: "free_talk" });
	created.push(pending.id);

	await page.goto("/voice/history");
	// Positive anchor: the list actually rendered.
	await expect(
		page.locator(`a[href="/voice/history/${endedId}"]`),
	).toBeVisible({ timeout: 20_000 });
	// The pending session must not leak into history.
	await expect(
		page.locator(`a[href="/voice/history/${pending.id}"]`),
	).toHaveCount(0);
});

test("abandoned never-started sessions are deleted, not saved", async () => {
	// Creating B closes A inside the same transaction: A never reached the
	// engine and holds no turns, so it is deleted rather than ended.
	const a = await apiCreateConversation({ mode: "free_talk" });
	created.push(a.id);
	const b = await apiCreateConversation({ mode: "free_talk" });
	created.push(b.id);

	expect(await dbConversationExists(a.id)).toBe(false);
	expect(await dbConversationExists(b.id)).toBe(true);
});
