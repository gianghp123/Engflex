import type { CEFR, VocabularyPartOfSpeech } from "@engflex/contracts";

/** The add-word dialog's form payload (mapped to the API in service.ts). */
export interface AddWordInput {
	term: string;
	cefr: CEFR;
	partOfSpeech: VocabularyPartOfSpeech;
	definition: string;
	example: string;
}
