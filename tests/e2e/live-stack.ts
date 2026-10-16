import path from "node:path";
import { defineConfig, devices } from "@playwright/test";
import { loadEnv } from "vite";
import { FAKE_MIC_AUDIO } from "./fake-mic";

/**
 * Shared live-stack plumbing for the two live Playwright configs.
 *
 * Both boot the real engine, Go API and web server on dedicated ports with
 * `reuseExistingServer: false` (a dev server on 8000 carries its own
 * VOICE_SERVICE_URL and would silently make a run prove nothing). The only
 * difference between them is the engine env, so it is a parameter: the
 * recovery suite needs the STT gate forced open, and the Analyze suite must
 * run with the gate off.
 *
 * Env loading is identical in both configs: Playwright runs in Node and never
 * reads dotenv files itself, and Vite only exposes VITE_* to the browser.
 * tests/e2e/.env wins, apps/web/.env.local is the fallback.
 * Shell env still wins.
 */
export function loadPlaywrightEnv(): void {
	const E2E_DIR = import.meta.dirname;
	const WEB_DIR = path.join(E2E_DIR, "../../apps/web");
	for (const dir of [E2E_DIR, WEB_DIR]) {
		for (const [key, value] of Object.entries(
			loadEnv("development", dir, ""),
		)) {
			process.env[key] ??= value;
		}
	}
	process.env.CLERK_PUBLISHABLE_KEY ??= process.env.VITE_CLERK_PUBLISHABLE_KEY;
}

const ENSURE_E2E_DB = path.join(import.meta.dirname, "ensure-e2e-db.mjs");

export const API_PORT = Number(process.env.E2E_LIVE_API_PORT ?? 8010);
export const WEB_PORT = Number(process.env.E2E_LIVE_WEB_PORT ?? 3010);
export const ENGINE_PORT = Number(process.env.E2E_LIVE_ENGINE_PORT ?? 7861);

// helpers.ts builds its API base URL from these (same convention as the
// default config). Workers inherit the main process env, so this is how the
// specs learn they are talking to the live stack.
process.env.E2E_API_PORT ??= String(API_PORT);
process.env.E2E_WEB_PORT ??= String(WEB_PORT);

// Go reads DB_* parts (not DATABASE_URL); the e2e database is created and
// migrated by ensure-e2e-db.mjs.
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

/** Chrome flags for the file-based fake mic.
 *
 * Load-bearing pairing: without --use-fake-device-for-media-stream Chrome
 * uses the REAL microphone and the file flag is silently ignored (this cost
 * several live runs their audio — the engine transcribed the room). The
 * file is looped on purpose (no %noloop): playback starts when the mic
 * track is acquired, well before the session's STT pipeline is listening.
 */
const FAKE_MIC_ARGS = [
	"--use-fake-ui-for-media-stream",
	"--use-fake-device-for-media-stream",
	`--use-file-for-fake-audio-capture=${FAKE_MIC_AUDIO}`,
];

/** Build a live config whose engine runs with `engineEnv`.
 *
 * INTERNAL_SECRET must match VOICE_INTERNAL_SECRET below (dev value pinned
 * so a drifted .env fails loudly instead of 401-ing every callback).
 */
export function defineLiveConfig({
	projectName,
	testMatch,
	engineEnv,
}: {
	projectName: string;
	testMatch: RegExp;
	engineEnv: Record<string, string>;
}) {
	return defineConfig({
		testDir: ".",
		fullyParallel: false,
		workers: 1,
		retries: 0,
		timeout: 180_000,
		reporter: [["list"], ["html", { open: "never" }]],
		use: {
			baseURL: `http://localhost:${WEB_PORT}`,
			trace: "retain-on-failure",
		},
		projects: [
			{ name: "global setup", testMatch: /global\.setup\.ts/ },
			{
				name: projectName,
				testMatch,
				dependencies: ["global setup"],
				use: {
					...devices["Desktop Chrome"],
					storageState: "playwright/.clerk/user.json",
					launchOptions: { args: FAKE_MIC_ARGS },
				},
			},
		],
		webServer: [
			{
				command: `uv run src/bot.py -t webrtc --port ${ENGINE_PORT}`,
				cwd: "../../apps/voice",
				url: `http://localhost:${ENGINE_PORT}/status`,
				reuseExistingServer: false,
				timeout: 120_000,
				// Turn batches must land in the e2e API/DB, not the dev one.
				env: {
					API_URL: `http://localhost:${API_PORT}`,
					INTERNAL_SECRET: "dev-internal-secret",
					SENTRY_DSN: "",
					...engineEnv,
				},
			},
			{
				// The API cannot boot without its database, and webServers
				// start before every setup hook — so create/migrate it here.
				command: `node ${ENSURE_E2E_DB} && go run ./cmd/server`,
				cwd: "../../apps/api",
				url: `http://localhost:${API_PORT}/health`,
				reuseExistingServer: false,
				timeout: 60_000,
				env: {
					PORT: String(API_PORT),
					...e2eDbParts(),
					// The router allows only configured origins: the specs
					// drive the API cross-origin from the live web server.
					CORS_ORIGINS: `http://localhost:${WEB_PORT}`,
					VOICE_SERVICE_URL: `http://localhost:${ENGINE_PORT}`,
					VOICE_INTERNAL_SECRET: "dev-internal-secret",
					VOICE_START_TIMEOUT_MS: "30000",
					VOICE_OFFER_TIMEOUT_MS: "30000",
					// One LLM pass on a proxied endpoint: generous, the tests
					// assert the outcome, not the latency.
					VOICE_ANALYZE_TIMEOUT_MS: "90000",
				},
			},
			{
				command: `pnpm dev --port ${WEB_PORT}`,
				cwd: "../../apps/web",
				url: `http://localhost:${WEB_PORT}`,
				reuseExistingServer: false,
				timeout: 120_000,
				env: {
					// Server-side API base (already carries the /api/v1 prefix).
					// The browser talks same-origin, so no VITE_ var is needed.
					API_INTERNAL_URL: `http://localhost:${API_PORT}/api/v1`,
				},
			},
		],
	});
}
