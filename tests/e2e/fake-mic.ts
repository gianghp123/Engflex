import { execFileSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

/**
 * Chromium's `--use-fake-device-for-media-stream` plays a 440 Hz tone, which
 * STT rejects — no transcript, no turn, nothing to analyze. The live suite
 * feeds a real speech WAV instead, synthesized offline with a local Piper
 * voice (see apps/voice/scripts/make_fake_mic_audio.py) — independent of
 * whatever TTS the engine itself is configured to use.
 *
 * Both the config (Chrome flag) and the spec (generation) resolve the path
 * here so they cannot drift.
 */
export const FAKE_MIC_AUDIO = path.join(
	import.meta.dirname,
	"./playwright/.fake-mic/speech.wav",
);

/** Generate the fixture if missing. Idempotent, so reruns cost nothing. */
export function ensureFakeMicAudio(): string {
	if (fs.existsSync(FAKE_MIC_AUDIO)) return FAKE_MIC_AUDIO;
	fs.mkdirSync(path.dirname(FAKE_MIC_AUDIO), { recursive: true });
	execFileSync(
		"uv",
		["run", "python", "scripts/make_fake_mic_audio.py", FAKE_MIC_AUDIO],
		{
			cwd: path.join(import.meta.dirname, "../../apps/voice"),
			stdio: "inherit",
		},
	);
	return FAKE_MIC_AUDIO;
}
