export function playDictationFeedback(hinted: boolean) {
	try {
		const context = new AudioContext();
		const tone = context.createOscillator();
		const volume = context.createGain();
		tone.frequency.value = hinted ? 440 : 660;
		volume.gain.setValueAtTime(0.06, context.currentTime);
		volume.gain.exponentialRampToValueAtTime(0.001, context.currentTime + 0.15);
		tone.connect(volume).connect(context.destination);
		tone.onended = () => void context.close().catch(() => {});
		if (context.state !== "running") {
			void context.close().catch(() => {});
			return;
		}
		tone.start();
		tone.stop(context.currentTime + 0.15);
	} catch {
		// Feedback is optional when the browser cannot play it.
	}
}
