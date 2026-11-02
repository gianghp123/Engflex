import { clerk } from "@clerk/testing/playwright";
import { test as base } from "@playwright/test";

// Stored Clerk JWT cookies expire quickly. Establish a fresh browser session
// per context so later tests exercise the page rather than an expired snapshot.
export const test = base.extend({
	page: [
		async ({ page }, use) => {
			const email = process.env.E2E_CLERK_USER_EMAIL;
			if (!email) throw new Error("E2E_CLERK_USER_EMAIL is required");
			await page.context().clearCookies();
			// Default browser suites assert English chrome; establish the locale
			// before the first SSR render and Clerk's navigation, not afterward.
			await page.context().addCookies([
				{
					name: "PARAGLIDE_LOCALE",
					value: "en",
					domain: "localhost",
					path: "/",
				},
			]);
			await page.goto("/");
			await clerk.signIn({ page, emailAddress: email });
			await use(page);
		},
		{ scope: "test", timeout: 45_000 },
	],
});

export { expect } from "@playwright/test";
