import type { VocabularyPartOfSpeech } from "@engflex/contracts";
import { m } from "#/paraglide/messages";

/** Display label for a wire POS token; the token itself is never shown. */
export const POS_LABELS: Record<VocabularyPartOfSpeech, () => string> = {
	noun: () => m["vocabulary.pos.noun"](),
	verb: () => m["vocabulary.pos.verb"](),
	adjective: () => m["vocabulary.pos.adjective"](),
	adverb: () => m["vocabulary.pos.adverb"](),
	pronoun: () => m["vocabulary.pos.pronoun"](),
	preposition: () => m["vocabulary.pos.preposition"](),
	conjunction: () => m["vocabulary.pos.conjunction"](),
	interjection: () => m["vocabulary.pos.interjection"](),
	determiner: () => m["vocabulary.pos.determiner"](),
	phrase: () => m["vocabulary.pos.phrase"](),
	idiom: () => m["vocabulary.pos.idiom"](),
	phrasal_verb: () => m["vocabulary.pos.phrasalVerb"](),
};

export function posLabel(value: VocabularyPartOfSpeech | undefined): string {
	return value ? POS_LABELS[value]() : "";
}
