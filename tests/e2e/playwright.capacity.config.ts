import { defineLiveConfig, loadPlaywrightEnv } from "./live-stack";

/**
 * LIVE e2e for admission control: a real dispatcher with MAX_SESSIONS=1.
 * Proves the cap rejects the (N+1)th start through the whole stack (engine
 * 429 -> Go 503 -> the room's blocking error modal) and that the slot frees
 * again once the live session's teardown reaches the dispatcher.
 *
 * Same prerequisites as the other live suites (real STT/LLM/TTS keys + the
 * Piper voice for the fake-mic fixture) — see README.md.
 */
loadPlaywrightEnv();

export default defineLiveConfig({
	projectName: "voice-capacity",
	testMatch: /voice-capacity\.spec\.ts/,
	engineEnv: {
		MAX_SESSIONS: "1",
	},
});
