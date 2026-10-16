import type {
	CEFR,
	VocabularyPartOfSpeech,
	VocabularySort,
	VocabularySource,
} from "@engflex/contracts";

/** The add-word dialog's form payload (mapped to the API in service.ts). */
export interface AddWordInput {
	term: string;
	cefr: CEFR;
	partOfSpeech: VocabularyPartOfSpeech;
	definition: string;
	example: string;
}

/** The hub's filter set: what the URL carries and the API narrows by. */
export interface VocabFilters {
	search: string;
	sort: VocabularySort;
	source: "all" | VocabularySource;
	cefr: "all" | CEFR;
}
