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
