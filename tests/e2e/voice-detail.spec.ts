import { expect, test } from "@playwright/test";
import { closeDb } from "./helpers";

const SEEDED_ID = "3a8b3191-c8a3-5855-8e05-0a268cedac3e";

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

test.afterAll(async () => {
	await closeDb();
});

test("detail: seeded scenario renders hero, content, and start CTA", async ({
	page,
}) => {
	await page.goto(`/voice/scenarios/${SEEDED_ID}`);

	// Hero: seeded title, topic badge, duration badge, partner tile.
	await expect(
		page.getByRole("heading", {
			name: "Sprint retrospective & timeline pushback",
		}),
	).toBeVisible({ timeout: 20_000 });
	await expect(page.getByText("Architecture reviews")).toBeVisible();
	await expect(page.getByText("Up to 8 min")).toBeVisible();
	await expect(page.getByText("Tom · Product manager")).toBeVisible();

	// Body: context, opening prompt, vocabulary.
	await expect(page.getByText("The situation & context")).toBeVisible();
	await expect(page.getByText("Tom (Opening statement):")).toBeVisible();
	await expect(page.getByText("Target vocabulary")).toBeVisible();
	await expect(page.getByText("backpressure", { exact: true })).toBeVisible();

	// Runs shell is empty in v1.
	await expect(page.getByText("No runs yet")).toBeVisible();

	await expect(
		page.getByRole("button", { name: "Start sparring with Mo" }),
	).toBeVisible();
});

test("detail: unknown scenario returns to the browser", async ({ page }) => {
	await page.goto("/voice/scenarios/00000000-0000-0000-0000-000000000000");
	await expect(page).toHaveURL(/\/voice\/scenarios\/?$/, { timeout: 20_000 });
	await expect(
		page.getByRole("heading", { name: "Select roleplay scenario" }),
	).toBeVisible();
});
