export type DictationLevel = "easy" | "normal" | "hard";

export const DICTATION_LEVELS: DictationLevel[] = ["easy", "normal", "hard"];

export function splitWords(sentence: string): string[] {
	return sentence.trim().split(/\s+/).filter(Boolean);
}

export function maskWord(word: string): string {
	return "*".repeat(word.length);
}

export function maskSentence(sentence: string): string {
	return splitWords(sentence).map(maskWord).join(" ");
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

export function cleanWordForComparison(word: string): string {
	return word.toLowerCase().replace(/^[^\p{L}\p{N}]+|[^\p{L}\p{N}]+$/gu, "");
}

export function getWordHint(word: string, level: DictationLevel): string {
	const clean = cleanWordForComparison(word);
	if (level === "easy") {
		if (clean.length <= 1) return clean;
		return `${clean[0]}${"*".repeat(clean.length - 1)}`;
	}
	if (level === "normal") {
		return "*".repeat(clean.length);
	}
	return "";
}
