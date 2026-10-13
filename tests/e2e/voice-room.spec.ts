import { expect, test } from "@playwright/test";
import {
	apiCreateConversation,
	apiEndConversation,
	clerkUserId,
	closeDb,
	dbDeleteConversation,
	dbSeedFeedback,
	dbSeedTurn,
} from "./helpers";

const created: string[] = [];

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
});

test.afterAll(async () => {
	await closeDb();
});

async function tracked(id: string): Promise<string> {
	created.push(id);
	return id;
}

test("service fail: dead engine shows the blocking error modal", async ({
	page,
}) => {
	await page.goto("/voice");
	const conv = await apiCreateConversation({ mode: "free_talk" });
	const id = await tracked(conv.id);

	// The engine is down by design (VOICE_SERVICE_URL points at a dead
	// port), so startBot fails and the kit surfaces it as `error`.
	await page.goto(`/voice/room/${id}`);
	const dialog = page.getByRole("alertdialog");
	await expect(dialog).toBeVisible({ timeout: 30_000 });
	await expect(dialog).toContainText("Connection failed. Please try again.");

	await dialog.getByRole("button", { name: "Back to scenarios" }).click();
	await expect(page).toHaveURL(/\/voice\/?$/);
});

test("ended session: persisted review renders seeded turns", async ({
	page,
}) => {
	await page.goto("/voice");
	const conv = await apiCreateConversation({ mode: "free_talk" });
	const id = await tracked(conv.id);
	await dbSeedTurn(id, {
		role: "user",
		text: "Hello from the e2e seed",
	});
	await apiEndConversation(id);

	await page.goto(`/voice/room/${id}`);
	await expect(page.getByText("Live transcript")).toBeVisible({
		timeout: 20_000,
	});
	await expect(page.getByText("Hello from the e2e seed")).toBeVisible();

	await page.getByRole("button", { name: "Back to scenarios" }).click();
	await expect(page).toHaveURL(/\/voice\/?$/);
});

test("analyze: engine-down error, then seeded feedback card", async ({
	page,
}) => {
	await page.goto("/voice");
	const conv = await apiCreateConversation({ mode: "free_talk" });
	const id = await tracked(conv.id);
	const turnId = await dbSeedTurn(id, {
		role: "user",
		text: "I go to the office yesterday",
	});
	await apiEndConversation(id);

	await page.goto(`/voice/room/${id}`);
	await page.getByRole("button", { name: "Analyze turn" }).first().click();
	// No engine running: the mutation fails and the per-turn error shows.
	await expect(page.getByText("Analysis failed. Try again.")).toBeVisible({
		timeout: 20_000,
	});

	// Seed the feedback row directly: the card renders from persisted data.
	const userId = await clerkUserId();
	await dbSeedFeedback(userId, turnId, {
		corrected: "I went to the office yesterday.",
		spans: [
			{
				text: "go",
				occurrence: 1,
				status: "incorrect",
				correction: "went",
				// Distinct from tip on purpose: identical strings render two
				// nodes and make this assertion ambiguous under strict mode.
				reason: "Simple past is required for a finished time.",
			},
		],
		relevance: { status: "relevant", reason: null },
		alternatives: { language: null, contextual: null },
		tip: "Use the past tense for yesterday.",
	});
	await page.reload();
	// The card renders the highlighted span, correction row, tip (the
	// corrected string itself is shown in the repair block) plus the
	// re-analyze affordance.
	await expect(page.getByText("Use the past tense for yesterday.")).toBeVisible(
		{ timeout: 20_000 },
	);
	await expect(
		page.getByRole("button", { name: "Analyze again" }),
	).toBeVisible();
});

test("roleplay start: seeded scenario creates a conversation and the room loads", async ({
	page,
}) => {
	await page.goto("/voice");
	const conv = await apiCreateConversation({
		mode: "roleplay",
		scenarioId: "01b9a032-e2cf-5bb0-8df5-7bf49704bee8",
	});
	const id = await tracked(conv.id);

	// Same dead-engine shape as the free-talk test above: startBot fails,
	// the kit surfaces it, and the exit path returns to the voice hub.
	await page.goto(`/voice/room/${id}`);
	const dialog = page.getByRole("alertdialog");
	await expect(dialog).toBeVisible({ timeout: 30_000 });
	await expect(dialog).toContainText("Connection failed. Please try again.");

	await dialog.getByRole("button", { name: "Back to scenarios" }).click();
	await expect(page).toHaveURL(/\/voice\/?$/);
});
