import path from "node:path";
import { clerk, clerkSetup } from "@clerk/testing/playwright";
import { test as setup } from "@playwright/test";

// Must run serially: clerkSetup() exports env (CLERK_FAPI,
// CLERK_TESTING_TOKEN) into this worker's process only.
setup.describe.configure({ mode: "serial" });

// The e2e database is created and migrated by ensure-e2e-db.mjs, which
// runs as part of the API's webServer command — webServers start before any
// setup hook, so it cannot live here.
setup("global setup", async () => {
	for (const name of ["DATABASE_URL", "CLERK_SECRET_KEY"] as const) {
		if (!process.env[name]) {
			throw new Error(`e2e: ${name} must be set (see README.md)`);
		}
	}
	await clerkSetup();
});

const authFile = path.join(
	import.meta.dirname,
	"./playwright/.clerk/user.json",
);

setup("authenticate and save state to storage", async ({ page }) => {
	const email = process.env.E2E_CLERK_USER_EMAIL;
	if (!email) {
		throw new Error("e2e: E2E_CLERK_USER_EMAIL must be set (see README.md)");
	}
	// Server-side token: bypasses verification/MFA regardless of instance
	// settings. Requires a pre-created user; the +clerk_test suffix
	// suppresses all Clerk email delivery.
	await page.goto("/");
	await clerk.signIn({ page, emailAddress: email });
	// Prove the session works against a real authed route, not just /.
	await page.goto("/voice");
	await page.waitForURL("**/voice");
	await page.context().storageState({ path: authFile });
});
