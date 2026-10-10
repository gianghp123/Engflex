import { defineLiveConfig, loadPlaywrightEnv } from "./live-stack";

/**
 * LIVE e2e for the Analyze flow: a real voice engine, a real session, a real
 * Analyze round trip — fake mic → Deepgram → collector → live turn post → Go
 * poll → Analyze button → engine `/analyze` → diagnostics card.
 *
 * Transcript recovery has its own config (playwright.recovery.config.ts)
 * because it drives the learner-facing correction flow. Nothing is forced in
 * either config any more: there is no confidence gate to open.
 *
 * Requires the engine's own prerequisites (apps/voice/.env with real STT,
 * LLM and TTS keys + the Piper voice in models/ for the fake-mic fixture) —
 * see README.md.
 */
loadPlaywrightEnv();

export default defineLiveConfig({
	projectName: "voice-live",
	testMatch: /voice-live\.spec\.ts/,
	engineEnv: {},
});
