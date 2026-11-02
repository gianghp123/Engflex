// Explicit provider-free acceptance against task-owned Docker containers.
import assert from "node:assert/strict";
import { execFileSync, spawnSync } from "node:child_process";
import { randomUUID } from "node:crypto";
import { fileURLToPath } from "node:url";
import { createClerkClient } from "@clerk/backend";
import pg from "pg";
import { loadEnv } from "vite";

const root = fileURLToPath(new URL("../../", import.meta.url));
const env = {
	...loadEnv("development", `${root}/apps/web`, ""),
	...process.env,
};
assert.ok(
	env.CLERK_SECRET_KEY?.startsWith("sk_test_"),
	"Development Clerk required",
);
const dsn = env.TEST_DATABASE_URL;
assert.ok(
	["localhost", "127.0.0.1"].includes(new URL(dsn).hostname),
	"Local test database required",
);
assert.ok(
	new URL(dsn).pathname.endsWith("_test"),
	"Dedicated test database required",
);
const origin = env.RUNTIME_API_ORIGIN ?? "http://127.0.0.1:55980";
const voiceOrigin = env.RUNTIME_VOICE_ORIGIN ?? "http://127.0.0.1:55960";
for (const endpoint of [origin, voiceOrigin]) {
	assert.ok(
		["localhost", "127.0.0.1"].includes(new URL(endpoint).hostname),
		"Local runtime endpoints required",
	);
}
const container = "codex-modular-voice";
const internal = "codex-internal-test";
const clerk = createClerkClient({ secretKey: env.CLERK_SECRET_KEY });
const database = new pg.Client({ connectionString: dsn });
const sessions = [];
const conversations = [];
const profiles = [];
const lessonId = randomUUID();
let stage = "setup";

async function request(path, expected, token, body, secret) {
	const response = await fetch(origin + path, {
		redirect: "error",
		method: body === undefined ? "GET" : "POST",
		headers: {
			"Content-Type": "application/json",
			...(token ? { Authorization: `Bearer ${token}` } : {}),
			...(secret ? { "X-Internal-Secret": secret } : {}),
		},
		body: body === undefined ? undefined : JSON.stringify(body),
		signal: AbortSignal.timeout(15000),
	});
	assert.equal(response.status, expected, stage);
	return response.json();
}

try {
	await database.connect();
	const users = await clerk.users.getUserList({
		emailAddress: [env.E2E_CLERK_USER_EMAIL, env.E2E_ADMIN_CLERK_USER_EMAIL],
	});
	const learner = users.data.find((u) =>
		u.emailAddresses.some((e) => e.emailAddress === env.E2E_CLERK_USER_EMAIL),
	);
	const admin = users.data.find((u) =>
		u.emailAddresses.some(
			(e) => e.emailAddress === env.E2E_ADMIN_CLERK_USER_EMAIL,
		),
	);
	assert.ok(learner && learner.publicMetadata.role !== "admin");
	assert.ok(admin?.publicMetadata.role === "admin");
	const profile = await database.query(
		"INSERT INTO user_profiles (clerk_user_id,onboarding_completed_at) VALUES ($1,now()) ON CONFLICT (clerk_user_id) DO NOTHING RETURNING clerk_user_id",
		[learner.id],
	);
	if (profile.rowCount) profiles.push(learner.id);
	async function token(user) {
		const session = await clerk.sessions.createSession({ userId: user.id });
		sessions.push(session.id);
		return (await clerk.sessions.getToken(session.id)).jwt;
	}
	const learnerToken = await token(learner);
	const adminToken = await token(admin);
	await database.query(
		"INSERT INTO lessons (id,slug,title,section_id,cefr_level,status,is_premium) SELECT $1,$2,'Runtime premium probe',id,'A1','published',true FROM lesson_sections LIMIT 1",
		[lessonId, `runtime-${lessonId}`],
	);
	const publicLesson = (
		await database.query(
			"SELECT id FROM lessons WHERE status='published' AND NOT is_premium LIMIT 1",
		)
	).rows[0].id;
	stage = "release smoke";
	const smoke = spawnSync(
		"python",
		[`${root}/infrastructure/scripts/smoke.py`],
		{
			cwd: root,
			env: {
				...process.env,
				SMOKE_API_URL: origin,
				SMOKE_LEARNER_TOKEN: learnerToken,
				SMOKE_PREMIUM_TOKEN: adminToken,
				SMOKE_PUBLIC_CONTENT_PATH: `/api/v1/lessons/${publicLesson}`,
				SMOKE_PREMIUM_CONTENT_PATH: `/api/v1/lessons/${lessonId}`,
				VOICE_INTERNAL_SECRET: internal,
				VNPAY_HASH_SECRET: "e2e-vnpay-secret",
			},
			encoding: "utf8",
		},
	);
	process.stdout.write(smoke.stdout ?? "");
	process.stderr.write(smoke.stderr ?? "");
	assert.equal(smoke.status, 0, "release smoke failed");
	console.log(
		"PASS release HTTP smoke (authorized-content probe uses admin preview; paid entitlement has separate tests)",
	);
	stage = "Go -> official Voice start";
	const created = await request(
		"/api/v1/conversations",
		201,
		await token(learner),
		{ mode: "free_talk" },
	);
	const id = created.data.id;
	conversations.push(id);
	const path = `/api/v1/conversations/${id}`;
	const started = await request(`${path}/start`, 200, await token(learner), {});
	assert.equal(started.data.sessionId, id);
	const sessionId = (
		await database.query(
			"SELECT speech_session_id FROM conversations WHERE id=$1",
			[id],
		)
	).rows[0].speech_session_id;
	assert.ok(sessionId);
	const replay = await request(`${path}/start`, 200, await token(learner), {});
	assert.equal(replay.data.sessionId, id);
	assert.equal(
		(
			await database.query(
				"SELECT speech_session_id FROM conversations WHERE id=$1",
				[id],
			)
		).rows[0].speech_session_id,
		sessionId,
	);
	console.log("PASS Go confirms official runner idempotent session");
	stage = "engine restart";
	const label = execFileSync(
		"docker",
		[
			"inspect",
			"--format",
			'{{index .Config.Labels "engflex.validation"}}',
			container,
		],
		{ encoding: "utf8" },
	).trim();
	assert.equal(label, "modular", "Refusing to restart a non-test container");
	execFileSync("docker", ["restart", "--time", "120", container], {
		stdio: "ignore",
	});
	let ready = false;
	for (let attempt = 0; attempt < 40; attempt++) {
		try {
			ready = (
				await fetch(`${voiceOrigin}/readyz`, {
					signal: AbortSignal.timeout(1000),
				})
			).ok;
		} catch {}
		if (ready) break;
		await new Promise((resolve) => setTimeout(resolve, 500));
	}
	assert.ok(ready, "Voice did not become ready after restart");
	const restarted = await request(
		`${path}/start`,
		200,
		await token(learner),
		{},
	);
	assert.equal(restarted.data.sessionId, id);
	const newSession = (
		await database.query(
			"SELECT speech_session_id FROM conversations WHERE id=$1",
			[id],
		)
	).rows[0].speech_session_id;
	assert.ok(newSession && newSession !== sessionId);
	console.log("PASS Go replaces stale engine session after engine restart");
	stage = "snapshots and atomic finalize";
	const internalPath = `/api/v1/internal/conversations/${id}`;
	const snapshot = {
		engineSessionId: newSession,
		revision: 1,
		turns: [
			{ position: 1, role: "user", text: "Validation original" },
			{ position: 2, role: "ai", text: "Validation stale tail" },
		],
	};
	await request(
		`${internalPath}/turns`,
		409,
		undefined,
		{ ...snapshot, engineSessionId: sessionId },
		internal,
	);
	await request(`${internalPath}/turns`, 200, undefined, snapshot, internal);
	const corrected = {
		engineSessionId: newSession,
		revision: 2,
		turns: [{ position: 1, role: "user", text: "Validation corrected" }],
	};
	await request(`${internalPath}/turns`, 200, undefined, corrected, internal);
	await request(`${internalPath}/turns`, 200, undefined, corrected, internal);
	await request(
		`${internalPath}/turns`,
		409,
		undefined,
		{
			...corrected,
			turns: [{ position: 1, role: "user", text: "Conflicting revision" }],
		},
		internal,
	);
	await request(`${internalPath}/turns`, 200, undefined, snapshot, internal);
	const final = { ...corrected, durationSec: 0 };
	await request(`${internalPath}/finalize`, 200, undefined, final, internal);
	await request(`${internalPath}/finalize`, 200, undefined, final, internal);
	const reviewed = (await request(path, 200, await token(learner))).data;
	assert.equal(reviewed.status, "ended");
	assert.equal(reviewed.durationSec, 0);
	assert.equal(reviewed.turns.length, 1);
	assert.equal(reviewed.turns[0].text, "Validation corrected");
	console.log(
		"PASS session identity, snapshot reorder/replay/conflict, correction truncation and zero-duration atomic finalize",
	);
} catch (error) {
	console.error(`FAIL ${stage}: ${error.name}`);
	if (error.code === "ERR_ASSERTION") console.error(error.message);
	process.exitCode = 1;
} finally {
	for (const id of conversations)
		await database.query("DELETE FROM conversations WHERE id=$1", [id]);
	await database.query("DELETE FROM lessons WHERE id=$1", [lessonId]);
	for (const id of profiles)
		await database.query("DELETE FROM user_profiles WHERE clerk_user_id=$1", [
			id,
		]);
	await database.end();
	for (const id of sessions) await clerk.sessions.revokeSession(id);
}
