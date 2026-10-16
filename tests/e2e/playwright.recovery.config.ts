import { defineLiveConfig, loadPlaywrightEnv } from "./live-stack";

/**
 * LIVE e2e for the on-demand transcript-recovery loop.
 *
 * No engine env override: nothing forces a correction window, because the
 * learner is the only trigger. The spec therefore drives it exactly as a
 * learner would — review a turn, correct it, re-speak, send, dismiss. The
 * Analyze suite runs the same engine with the same defaults and must be
 * unaffected.
 */
loadPlaywrightEnv();

export default defineLiveConfig({
	projectName: "voice-recovery",
	testMatch: /voice-recovery\.spec\.ts/,
	engineEnv: {},
});
