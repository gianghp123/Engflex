/**
 * Procedural Web Audio API sound engine for the Lusion × Engflex spatial landing page.
 * Synthesizes subtle, futuristic UI harmonics, CEFR orbital chimes, and acoustic scan
 * pulses without requiring external audio assets.
 */

let sharedCtx: AudioContext | null = null;

function getAudioContext(): AudioContext | null {
	if (typeof window === "undefined") return null;
	if (!sharedCtx) {
		const AudioCtx =
			window.AudioContext ||
			(window as unknown as { webkitAudioContext?: typeof AudioContext })
				.webkitAudioContext;
		if (!AudioCtx) return null;
		sharedCtx = new AudioCtx();
	}
	if (sharedCtx.state === "suspended") {
		void sharedCtx.resume();
	}
	return sharedCtx;
}

/** Plays a warm harmonic chord when the user enables interactive sound. */
export function playSoundToggleEffect(enabled: boolean): void {
	if (!enabled) return;
	const ctx = getAudioContext();
	if (!ctx) return;

	const now = ctx.currentTime;
	const freqs = [261.63, 329.63, 392.0, 523.25]; // C4 major 7th/octave shimmer
	for (let i = 0; i < freqs.length; i++) {
		const osc = ctx.createOscillator();
		const gain = ctx.createGain();
		osc.type = i % 2 === 0 ? "sine" : "triangle";
		osc.frequency.setValueAtTime(freqs[i], now + i * 0.035);

		gain.gain.setValueAtTime(0.0001, now + i * 0.035);
		gain.gain.exponentialRampToValueAtTime(0.035, now + i * 0.035 + 0.04);
		gain.gain.exponentialRampToValueAtTime(0.0001, now + i * 0.035 + 0.42);

		osc.connect(gain);
		gain.connect(ctx.destination);
		osc.start(now + i * 0.035);
		osc.stop(now + i * 0.035 + 0.45);
	}
}

/** Plays a pitched crystalline chime tuned to the selected CEFR tier (0 = A1 .. 5 = C2). */
export function playTierChime(enabled: boolean, tierIndex: number): void {
	if (!enabled) return;
	const ctx = getAudioContext();
	if (!ctx) return;

	const pentatonic = [293.66, 329.63, 392.0, 440.0, 523.25, 659.25];
	const baseFreq = pentatonic[tierIndex % pentatonic.length] ?? 392.0;
	const now = ctx.currentTime;

	const osc = ctx.createOscillator();
	const overtone = ctx.createOscillator();
	const gain = ctx.createGain();

	osc.type = "sine";
	overtone.type = "triangle";
	osc.frequency.setValueAtTime(baseFreq, now);
	osc.frequency.exponentialRampToValueAtTime(baseFreq * 1.02, now + 0.28);

	overtone.frequency.setValueAtTime(baseFreq * 1.5, now);

	gain.gain.setValueAtTime(0.0001, now);
	gain.gain.exponentialRampToValueAtTime(0.045, now + 0.02);
	gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.34);

	osc.connect(gain);
	overtone.connect(gain);
	gain.connect(ctx.destination);

	osc.start(now);
	overtone.start(now);
	osc.stop(now + 0.36);
	overtone.stop(now + 0.36);
}

/** Plays a resonant acoustic wave sweep when testing a voice scenario or phoneme scan. */
export function playVoicePulseSound(enabled: boolean): void {
	if (!enabled) return;
	const ctx = getAudioContext();
	if (!ctx) return;

	const now = ctx.currentTime;
	const osc1 = ctx.createOscillator();
	const osc2 = ctx.createOscillator();
	const filter = ctx.createBiquadFilter();
	const gain = ctx.createGain();

	osc1.type = "sawtooth";
	osc2.type = "sine";
	osc1.frequency.setValueAtTime(146.83, now); // D3
	osc1.frequency.exponentialRampToValueAtTime(293.66, now + 0.25);
	osc1.frequency.exponentialRampToValueAtTime(220.0, now + 0.55);

	osc2.frequency.setValueAtTime(440.0, now);
	osc2.frequency.exponentialRampToValueAtTime(587.33, now + 0.35);

	filter.type = "bandpass";
	filter.frequency.setValueAtTime(600, now);
	filter.frequency.exponentialRampToValueAtTime(1800, now + 0.25);
	filter.frequency.exponentialRampToValueAtTime(800, now + 0.55);
	filter.Q.setValueAtTime(3.5, now);

	gain.gain.setValueAtTime(0.0001, now);
	gain.gain.exponentialRampToValueAtTime(0.05, now + 0.05);
	gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.58);

	osc1.connect(filter);
	osc2.connect(filter);
	filter.connect(gain);
	gain.connect(ctx.destination);

	osc1.start(now);
	osc2.start(now);
	osc1.stop(now + 0.6);
	osc2.stop(now + 0.6);
}

/** Plays a subtle tactile micro-tick when clicking word chips or scenario tabs. */
export function playMicroTick(enabled: boolean, pitchMultiplier = 1): void {
	if (!enabled) return;
	const ctx = getAudioContext();
	if (!ctx) return;

	const now = ctx.currentTime;
	const osc = ctx.createOscillator();
	const gain = ctx.createGain();

	osc.type = "sine";
	osc.frequency.setValueAtTime(680 * pitchMultiplier, now);
	osc.frequency.exponentialRampToValueAtTime(
		320 * pitchMultiplier,
		now + 0.055,
	);

	gain.gain.setValueAtTime(0.0001, now);
	gain.gain.exponentialRampToValueAtTime(0.028, now + 0.008);
	gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.06);

	osc.connect(gain);
	gain.connect(ctx.destination);

	osc.start(now);
	osc.stop(now + 0.065);
}
