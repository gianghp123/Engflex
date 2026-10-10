/**
 * Boundary report for a live voice-e2e run. Reads the COMBINED Playwright run
 * log (engine + Go + vite all arrive as [WebServer] lines) and the e2e
 * database, then prints one verdict per pipeline boundary for a conversation:
 *
 *   session -> webrtc -> pcm -> vad -> stt-text -> db -> analyze
 *
 * Purpose: when the live test fails, this names the first failing link
 * instead of leaving a 150 s timeout as the only evidence. The pipeline
 * itself is 100% real (no mocks anywhere); this script only READS logs/DB.
 *
 * Engine loguru kwargs (conversation_id, turn counts) are NOT rendered by the
 * log format, so engine lines are attributed by TIME: the single
 * "session starting" line after the test's conversation marker owns every
 * engine line until the next session start (serial worker: unambiguous).
 * Go slog lines carry conversationID=/conversationID= keyvals and match
 * directly. Turn texts come from the DB (ground truth), not log parsing.
 *
 * Usage:
 *   node live-boundaries.mjs <run-log-path> <conversation-id>
 *   (DATABASE_URL must be set, same as the suite)
 */
import fs from "node:fs";
import pg from "pg";

const [logPath, convId] = process.argv.slice(2);
if (!logPath || !convId) {
	console.error("usage: node live-boundaries.mjs <run-log> <conversation-id>");
	process.exit(2);
}

const lines = fs.readFileSync(logPath, "utf8").split("\n");
const ws = lines.filter((l) => l.includes("[WebServer]"));
const isGo = (l) => /level=(INFO|ERROR|WARN|DEBUG)|msg="/.test(l);
const go = ws.filter(isGo);
const engine = ws.filter((l) => !isGo(l));

// Test marker printed by the spec (console.log lands in the run output).
const markerIdx = lines.findIndex((l) =>
	l.includes(`e2e: live conversation ${convId}`),
);

// Engine session attribution by time: first "session starting" at/after the
// marker (fall back to the first one in the log).
let sessionIdx = engine.findIndex(
	(l) =>
		l.includes("session starting") &&
		(markerIdx === -1 || lines.indexOf(l) >= markerIdx),
);
if (sessionIdx === -1)
	sessionIdx = engine.findIndex((l) => l.includes("session starting"));
const session = sessionIdx === -1 ? [] : engine.slice(sessionIdx);

const rows = [];
const row = (link, status, evidence) => rows.push({ link, status, evidence });
const find = (arr, re) => arr.find((l) => re.test(l));

if (sessionIdx === -1) {
	row("session", "FAIL", "no engine 'session starting' in log");
} else {
	row("session", "PASS", find(session, /session starting/)?.slice(-90));
}

const saw = (re) => (session.length ? find(session, re) : undefined);
const ice = saw(/ICE connection state is completed/);
row(
	"webrtc",
	ice ? "PASS" : session.length ? "FAIL" : "NOT-REACHED",
	ice?.slice(-110),
);

const timeouts = session.filter((l) =>
	l.includes("No audio frame received"),
).length;
const ttfb = saw(
	/DeepgramSTTService#\d+ usage audio seconds|DeepgramSTTService#\d+ TTFB/,
);
const dgConn = saw(/DeepgramSTTService#\d+: Websocket connection initialized/);
row(
	"pcm",
	!session.length
		? "NOT-REACHED"
		: ttfb
			? "PASS"
			: timeouts > 5
				? "FAIL"
				: "DEGRADED",
	ttfb
		? ttfb.slice(-110)
		: `audio-frame timeouts=${timeouts} (sustained spam means starved RTP); deepgram ws=${dgConn ? "up" : "down"}`,
);

const vadStart = saw(/User started speaking/);
row(
	"vad",
	vadStart ? "PASS" : session.length ? "FAIL" : "NOT-REACHED",
	vadStart?.slice(-110),
);

// STT text: user turns beyond the seeded "Hello!" opener, from engine LLM
// context lines (the only place full turn text is logged).
const userTexts = new Set();
for (const l of session) {
	const m = l.match(/'role': 'user', 'content': '((?:[^'\\]|\\.)*)'/g);
	if (m) for (const t of m) userTexts.add(t);
}
const realTexts = [...userTexts].filter((t) => !t.includes("Hello!"));
row(
	"stt-text",
	realTexts.length ? "PASS" : session.length ? "FAIL" : "NOT-REACHED",
	realTexts.length
		? realTexts.map((t) => t.slice(-80)).join(" | ")
		: "no transcribed user turn in engine log",
);

// DB ground truth (roles, positions, texts, feedback) — no log parsing.
let dbTurns = [];
let dbError = null;
try {
	const pool = new pg.Pool({ connectionString: process.env.DATABASE_URL });
	const { rows } = await pool.query(
		`SELECT t.position, t.role, LEFT(t.text, 60) AS text,
		        (f.subject_id IS NOT NULL) AS has_feedback
		 FROM conversation_turns t
		 LEFT JOIN feedbacks f
		   ON f.subject_type = 'conversation_turn' AND f.subject_id = t.id::text
		 WHERE t.conversation_id = $1 ORDER BY t.position`,
		[convId],
	);
	dbTurns = rows;
	await pool.end();
} catch (e) {
	dbError = String(e);
}
if (dbError) {
	row("db", "UNKNOWN", `db query failed: ${dbError}`);
} else if (!dbTurns.length) {
	row("db", "FAIL", "no turns persisted for the conversation");
} else {
	const summary = dbTurns
		.map(
			(t) =>
				`#${t.position}:${t.role}${t.has_feedback ? "+fb" : ""} ${JSON.stringify(t.text)}`,
		)
		.join(" / ");
	row("db", "PASS", summary);
}

const goConv = go.filter((l) => l.includes(convId));
const analyzed = goConv.find((l) => l.includes("turn analyzed"));
const analyzeFail = goConv.find(
	(l) => l.includes("analyze turn") && l.includes("level=ERROR"),
);
row(
	"analyze",
	analyzed ? "PASS" : analyzeFail ? "FAIL" : "NOT-REACHED",
	analyzed?.slice(-140) ??
		analyzeFail?.slice(-200) ??
		"no analyze attempt logged for the conversation",
);

const width = Math.max(...rows.map((r) => r.link.length));
for (const r of rows) {
	console.log(
		`${r.link.padEnd(width)}  ${r.status.padEnd(11)} ${r.evidence ?? ""}`,
	);
}
const firstBad = rows.find(
	(r) => r.status === "FAIL" || r.status === "DEGRADED",
);
console.log(
	firstBad
		? `\nfirst failing boundary: ${firstBad.link}`
		: "\nall boundaries PASS",
);
