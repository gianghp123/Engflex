export function splitWords(sentence: string): string[] {
	return sentence.trim().split(/\s+/).filter(Boolean);
}

export function maskWord(word: string): string {
	return "*".repeat(word.length);
}

export function maskSentence(sentence: string): string {
	return splitWords(sentence).map(maskWord).join(" ");
}

function normalizeWord(word: string): string {
	return word.normalize("NFC").toLowerCase().replace(/\p{P}/gu, "");
}

export type DictationLevel = "easy" | "normal" | "hard";

export function dictationInputWords(draft: string) {
	return splitWords(draft).filter((word) => normalizeWord(word));
}

export function checkDictation(
	sentence: string,
	draft: string,
	level: DictationLevel = "hard",
	answers: string[] = [],
) {
	const expected = splitWords(sentence).filter((word) => normalizeWord(word));
	const typed = splitWords(draft).map(normalizeWord).filter(Boolean);
	const count = Math.ceil(
		expected.length / (level === "easy" ? 3 : level === "normal" ? 2 : 1),
	);
	const positions = Array.from({ length: count }, (_, index) =>
		Math.floor((index * expected.length) / count),
	);
	const blankIndices = new Map(
		positions.map((position, index) => [position, index]),
	);
	const slots = expected.map((word, index) => {
		const blankIndex = blankIndices.get(index);
		const parts = splitWords(answers[blankIndex ?? -1] ?? "")
			.map(normalizeWord)
			.filter(Boolean);
		const answer = level === "hard" ? typed[index] : parts[0];
		const status = !answer
			? ("empty" as const)
			: (level === "hard" || parts.length === 1) &&
					answer === normalizeWord(word)
				? ("correct" as const)
				: ("incorrect" as const);
		return { word, status, blankIndex, position: index };
	});
	return {
		slots,
		complete:
			count > 0 &&
			(level !== "hard" || slots.length === typed.length) &&
			slots.every(
				(slot) => slot.blankIndex === undefined || slot.status === "correct",
			),
	};
}

/** Zip a sentence with its space-separated IPA, one token per word. */
export function pairWords(
	content: string,
	phonetic: string,
): { word: string; ipa?: string }[] {
	const ipa = splitWords(phonetic);
	return splitWords(content).map((word, index) => ({
		word,
		ipa: ipa[index],
	}));
}

export function extractYouTubeId(url?: string): string | null {
	if (!url) return null;
	const match = url.match(
		/(?:youtu\.be\/|youtube\.com\/(?:embed\/|v\/|watch\?v=|watch\?.+&v=))([\w-]{11})/,
	);
	return match ? match[1] : null;
}
