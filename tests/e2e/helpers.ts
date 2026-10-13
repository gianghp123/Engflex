import { createClerkClient } from "@clerk/backend";
import type { Conversation, TurnFeedback } from "@engflex/contracts";
import { Pool } from "pg";

const API_BASE = `http://localhost:${process.env.E2E_API_PORT ?? 8001}`;

function mustEnv(name: string): string {
	const value = process.env[name];
	if (!value) throw new Error(`e2e: ${name} must be set (see README.md)`);
	return value;
}

const clerk = () =>
	createClerkClient({ secretKey: mustEnv("CLERK_SECRET_KEY") });

let cachedUserId: string | null = null;
let sessionId: string | null = null;

/** Test user id, resolved once per worker via the Backend API. */
export async function clerkUserId(): Promise<string> {
	if (cachedUserId) return cachedUserId;
	const email = mustEnv("E2E_CLERK_USER_EMAIL");
	const found = await clerk().users.getUserList({ emailAddress: [email] });
	const user = found.data[0];
	if (!user) {
		throw new Error(
			`e2e: no Clerk user with email ${email} (create it in the Clerk dashboard, see README.md)`,
		);
	}
	cachedUserId = user.id;
	return user.id;
}

/** Fresh session JWT, minted server-side per the Clerk testing docs
 *  (create session -> create session token -> Bearer header). Tokens live
 *  60s, so one is minted for every API call and used immediately. */
export async function apiToken(): Promise<string> {
	const userId = await clerkUserId();
	try {
		sessionId ??= (await clerk().sessions.createSession({ userId })).id;
		const { jwt } = await clerk().sessions.getToken(sessionId);
		if (!jwt) throw new Error("empty token");
		return jwt;
	} catch {
		// Session expired or revoked: recreate once, then give up loudly.
		sessionId = (await clerk().sessions.createSession({ userId })).id;
		const { jwt } = await clerk().sessions.getToken(sessionId);
		if (!jwt) throw new Error("e2e: Clerk returned an empty session token");
		return jwt;
	}
}

export async function revokeSession(): Promise<void> {
	if (!sessionId) return;
	try {
		await clerk().sessions.revokeSession(sessionId);
	} finally {
		sessionId = null;
	}
}

async function apiFetch(
	method: string,
	path: string,
	body?: unknown,
): Promise<unknown> {
	const token = await apiToken();
	const res = await fetch(`${API_BASE}${path}`, {
		method,
		headers: {
			"Content-Type": "application/json",
			Authorization: `Bearer ${token}`,
		},
		body: body ? JSON.stringify(body) : undefined,
	});
	const json = (await res.json()) as { message: string; data?: unknown };
	if (!res.ok) {
		throw new Error(`e2e: ${method} ${path} -> ${res.status} ${json.message}`);
	}
	return json.data;
}

export async function apiCreateConversation(input: {
	mode: "free_talk" | "roleplay";
	scenarioId?: string;
}): Promise<Conversation> {
	return (await apiFetch(
		"POST",
		"/api/v1/conversations",
		input,
	)) as Conversation;
}

export async function apiEndConversation(id: string): Promise<Conversation> {
	return (await apiFetch(
		"POST",
		`/api/v1/conversations/${id}/end`,
	)) as Conversation;
}

let pool: Pool | null = null;

export function db(): Pool {
	pool ??= new Pool({ connectionString: mustEnv("DATABASE_URL") });
	return pool;
}

export async function dbSeedTurn(
	conversationId: string,
	input: { role: "user" | "ai"; text: string; position?: number },
): Promise<string> {
	const position =
		input.position ??
		Number(
			(
				await db().query(
					"SELECT COALESCE(MAX(position), 0) + 1 AS next FROM conversation_turns WHERE conversation_id = $1",
					[conversationId],
				)
			).rows[0].next,
		);
	const { rows } = await db().query(
		`INSERT INTO conversation_turns (conversation_id, position, role, text)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		[conversationId, position, input.role, input.text],
	);
	return rows[0].id as string;
}

export async function dbSeedFeedback(
	userId: string,
	turnId: string,
	payload: unknown,
): Promise<void> {
	await db().query(
		`INSERT INTO feedbacks (user_id, subject_type, subject_id, payload)
		 VALUES ($1, 'conversation_turn', $2, $3)
		 ON CONFLICT (subject_type, subject_id)
		 DO UPDATE SET payload = EXCLUDED.payload, updated_at = now()`,
		[userId, turnId, JSON.stringify(payload)],
	);
}

export async function dbDeleteConversation(id: string): Promise<void> {
	// turns cascade; feedback rows are subject-keyed, delete explicitly
	await db().query(
		"DELETE FROM feedbacks WHERE subject_type = 'conversation_turn' AND subject_id IN (SELECT id FROM conversation_turns WHERE conversation_id = $1)",
		[id],
	);
	await db().query("DELETE FROM conversations WHERE id = $1", [id]);
}

export type DbTurn = {
	id: string;
	position: number;
	role: string;
	text: string;
};

/** Persisted turns in position order — the live panel's Analyze buttons are
 *  keyed by exactly this ordinal, so the DB is the ground truth for whether
 *  the UI attached feedback to the turn it meant. */
export async function dbListTurns(conversationId: string): Promise<DbTurn[]> {
	const { rows } = await db().query(
		`SELECT id, position, role, text FROM conversation_turns
		 WHERE conversation_id = $1 ORDER BY position`,
		[conversationId],
	);
	return rows as DbTurn[];
}

/** Every learner turn plus whatever feedback is stored for it, in position
 *  order. The live suite analyzes the first *cleanly transcribed* turn (not
 *  necessarily the first turn: a mangled streaming fragment can precede it),
 *  so it needs the whole list, not just rows[0]. */
export async function dbLearnerTurns(conversationId: string): Promise<
	{
		turn: DbTurn;
		feedback: TurnFeedback | null;
	}[]
> {
	const { rows } = await db().query(
		`SELECT t.id, t.position, t.role, t.text, f.payload
		 FROM conversation_turns t
		 LEFT JOIN feedbacks f
		   ON f.subject_type = 'conversation_turn' AND f.subject_id = t.id
		 WHERE t.conversation_id = $1 AND t.role = 'user'
		 ORDER BY t.position`,
		[conversationId],
	);
	return rows.map((row) => {
		const payload =
			row.payload === null
				? null
				: typeof row.payload === "string"
					? JSON.parse(row.payload)
					: row.payload;
		return {
			turn: row as DbTurn,
			feedback: payload as TurnFeedback | null,
		};
	});
}

export async function closeDb(): Promise<void> {
	await revokeSession();
	await pool?.end();
	pool = null;
}
