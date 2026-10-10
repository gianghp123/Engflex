/**
 * Prepare the database the Go API boots against, before it boots.
 *
 * Runs as the first half of the API's `webServer.command`
 * (`node ensure-e2e-db.mjs && go run ./cmd/server`) because Playwright
 * starts webServers *before* any setup hook — a globalSetup module or a setup
 * test project both run too late, and the API refuses to boot without its
 * database. Idempotent: `CREATE DATABASE` is skipped when it already exists
 * (42P04) and migrations are no-ops for an up-to-date database, so pointing
 * DATABASE_URL at the dev database is safe — the dev server's own `.env`
 * never wins because the config passes DB_* through the real environment.
 *
 * Kept as a script rather than a TS module so it can run standalone, with cwd
 * anywhere: `pg` resolves from tests/e2e/node_modules, and `make` runs in
 * apps/api.
 */
import { execFileSync } from "node:child_process";
import path from "node:path";
import pg from "pg";

const dsn = process.env.DATABASE_URL;
if (!dsn) {
	console.error("e2e: DATABASE_URL must be set (see README.md)");
	process.exit(1);
}

const url = new URL(dsn);
const dbName = url.pathname.replace(/^\//, "");
url.pathname = "/postgres";
const admin = new pg.Client({ connectionString: url.toString() });
await admin.connect();
try {
	await admin.query(`CREATE DATABASE "${dbName.replace(/"/g, "")}"`);
	console.log(`e2e: created database ${dbName}`);
} catch (error) {
	// 42P04 = duplicate_database: already exists, which is fine.
	if (error.code !== "42P04") throw error;
} finally {
	await admin.end();
}

// Real env wins over apps/api/.env (godotenv never overwrites), so this
// migrates the database DATABASE_URL names.
execFileSync("make", ["migration-up"], {
	cwd: path.join(import.meta.dirname, "../../apps/api"),
	env: { ...process.env },
	stdio: "inherit",
});
