import { expect, test } from "@playwright/test";
import { closeDb } from "./helpers";

const SEEDED_ID = "7c000003-0000-4000-8000-000000000003";

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
			name: "Asking for street directions",
		}),
	).toBeVisible({ timeout: 20_000 });
	await expect(page.getByText("Getting around")).toBeVisible();
	await expect(page.getByText("Up to 5 min")).toBeVisible();
	await expect(page.getByText("Rosa · Taxi driver")).toBeVisible();

	// Body: context, opening prompt, vocabulary.
	await expect(page.getByText("The situation & context")).toBeVisible();
	await expect(page.getByText("Rosa:")).toBeVisible();
	await expect(page.getByText("Target vocabulary")).toBeVisible();
	await expect(page.getByText("corner", { exact: true })).toBeVisible();

	// No runs on this seeded scenario: the runs section shows the empty copy.
	// (20s: the runs list loads client-side after the scenario query.)
	await expect(page.getByText("No runs yet")).toBeVisible({ timeout: 20_000 });

	await expect(
		page.getByRole("button", { name: "Start sparring with Mo" }),
	).toBeVisible();
});

test("detail: unknown scenario shows the not-found page", async ({ page }) => {
	await page.goto("/voice/scenarios/00000000-0000-0000-0000-000000000000");
	await expect(page).toHaveURL(
		/\/voice\/scenarios\/00000000-0000-0000-0000-000000000000$/,
	);
	await expect(
		page.getByText("This page doesn't exist or the item was removed."),
	).toBeVisible({ timeout: 20_000 });
});
