import path from "node:path";
import { defineConfig, devices } from "@playwright/test";
import { loadEnv } from "vite";

// Playwright runs in Node and never loads dotenv files on its own (and Vite
// only exposes VITE_*-prefixed vars to the browser anyway), so pull
// tests/e2e/.env (+ apps/web/.env.local as fallback) in here. Shell env
// wins: existing entries are never overwritten, and the e2e dir wins over
// the web fallback.
const E2E_DIR = import.meta.dirname;
const WEB_DIR = path.join(E2E_DIR, "../../apps/web");
for (const dir of [E2E_DIR, WEB_DIR]) {
	for (const [key, value] of Object.entries(loadEnv("development", dir, ""))) {
		process.env[key] ??= value;
	}
}
// @clerk/testing wants the unprefixed name; the app uses the VITE_ one.
process.env.CLERK_PUBLISHABLE_KEY ??= process.env.VITE_CLERK_PUBLISHABLE_KEY;

/**
 * E2E against a live backend (Go API + Postgres). The voice engine is
 * intentionally NOT started: VOICE_SERVICE_URL points at a dead port so
 * setup-failure paths are deterministic.
 *
 * Ports are dedicated (8001/3001) and both servers boot with
 * `reuseExistingServer: false`, exactly like live-stack.ts: a dev server on
 * 8000 carries its own VOICE_SERVICE_URL and its own database, so reusing it
 * silently makes a run prove nothing (the observed failure modes are a missing
 * error modal and seed rows landing in a database the API is not using).
 *
 * Required env (see README.md):
 *   DATABASE_URL              Postgres DSN — also the database the API is
 *                             pointed at, and what test seeding writes to
 *   CLERK_SECRET_KEY          same Clerk instance as the apps
 *   CLERK_PUBLISHABLE_KEY     or VITE_CLERK_PUBLISHABLE_KEY
 *   E2E_CLERK_USER_EMAIL      pre-created test user (…+clerk_test@… suffix)
 */
const ENSURE_E2E_DB = path.join(import.meta.dirname, "ensure-e2e-db.mjs");
const API_PORT = Number(process.env.E2E_API_PORT ?? 8001);
const WEB_PORT = Number(process.env.E2E_WEB_PORT ?? 3001);

// helpers.ts builds its API base URL from these (same convention as
// live-stack.ts), so the specs learn the port even when it was defaulted
// here rather than exported by the shell.
process.env.E2E_API_PORT ??= String(API_PORT);
process.env.E2E_WEB_PORT ??= String(WEB_PORT);

// Go reads DB_* parts (not DATABASE_URL), so derive them from the single
// DATABASE_URL — the same database the tests seed directly.
function e2eDbParts(): Record<string, string> {
	const raw = process.env.DATABASE_URL;
	if (!raw) return {};
	const url = new URL(raw);
	return {
		DB_HOST: url.hostname,
		DB_PORT: url.port || "5432",
		DB_USER: decodeURIComponent(url.username),
		DB_PASSWORD: decodeURIComponent(url.password),
		DB_NAME: url.pathname.replace(/^\//, ""),
		DB_SSLMODE: url.searchParams.get("sslmode") ?? "disable",
	};
}

export default defineConfig({
	testDir: ".",
	fullyParallel: false,
	workers: 1,
	retries: 0,
	reporter: [["list"], ["html", { open: "never" }]],
	use: {
		baseURL: `http://localhost:${WEB_PORT}`,
		trace: "retain-on-failure",
	},
	projects: [
		{ name: "global setup", testMatch: /global\.setup\.ts/ },
		{
			name: "voice",
			testMatch: /voice-(room|detail)\.spec\.ts/,
			dependencies: ["global setup"],
			use: {
				...devices["Desktop Chrome"],
				storageState: "playwright/.clerk/user.json",
				launchOptions: {
					args: [
						"--use-fake-device-for-media-stream",
						"--use-fake-ui-for-media-stream",
					],
				},
			},
		},
	],
	webServer: [
		{
			// The API cannot boot without its database, and webServers start
			// before every setup hook — so create/migrate it right here.
			command: `node ${ENSURE_E2E_DB} && go run ./cmd/server`,
			cwd: "../../apps/api",
			url: `http://localhost:${API_PORT}/health`,
			reuseExistingServer: false,
			timeout: 60_000,
			env: {
				PORT: String(API_PORT),
				...e2eDbParts(),
				VOICE_SERVICE_URL: "http://127.0.0.1:9",
				VOICE_START_TIMEOUT_MS: "3000",
				VOICE_OFFER_TIMEOUT_MS: "3000",
				VOICE_ANALYZE_TIMEOUT_MS: "10000",
			},
		},
		{
			command: `pnpm dev --port ${WEB_PORT}`,
			cwd: "../../apps/web",
			url: `http://localhost:${WEB_PORT}`,
			reuseExistingServer: false,
			timeout: 120_000,
			env: {
				// API_URL is the api client base and already carries the
				// /api/v1 prefix (see .env.local); routes are appended to it.
				// Set here so the dev server cannot fall back to the 8000 in
				// .env.local while this stack runs on 8001.
				VITE_API_URL: `http://localhost:${API_PORT}/api/v1`,
			},
		},
	],
});
