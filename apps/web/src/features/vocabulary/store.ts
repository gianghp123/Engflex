import type { VocabularyItem } from "@engflex/contracts";
import { Store } from "@tanstack/store";
import { VOCABULARY_ITEMS } from "./fixtures";
import type { AddWordInput } from "./types";

type VocabularyState = {
	mastered: Record<string, boolean>;
	notes: Record<string, string | undefined>;
	added: VocabularyItem[];
};

export const vocabularyStore = new Store<VocabularyState>({
	mastered: {},
	notes: {},
	added: [],
});

/** Fixture + user-added items with store overrides applied. */
export function getVocabularyItems(): VocabularyItem[] {
	const { mastered, notes, added } = vocabularyStore.state;
	return [...VOCABULARY_ITEMS, ...added].map((item) => ({
		...item,
		userState: item.userState
			? {
					...item.userState,
					mastered: mastered[item.id] ?? item.userState.mastered,
					note: item.id in notes ? notes[item.id] : item.userState.note,
				}
			: item.userState,
	}));
}

export function getVocabularyItem(id: string): VocabularyItem | undefined {
	return getVocabularyItems().find((item) => item.id === id);
}

export function toggleMastered(id: string): void {
	vocabularyStore.setState((state) => {
		const current =
			state.mastered[id] ??
			VOCABULARY_ITEMS.find((item) => item.id === id)?.userState?.mastered ??
			false;
		return {
			...state,
			mastered: { ...state.mastered, [id]: !current },
		};
	});
}

export function saveNote(id: string, note: string): void {
	vocabularyStore.setState((state) => ({
		...state,
		notes: { ...state.notes, [id]: note },
	}));
}

export function addCustomWord(input: AddWordInput): void {
	const id = `custom-${input.term.toLowerCase().replace(/[^a-z0-9]+/g, "-")}`;
	vocabularyStore.setState((state) => ({
		...state,
		added: [
			...state.added,
			{
				id,
				term: input.term,
				ipa: "",
				partOfSpeech: input.partOfSpeech,
				cefr: input.cefr,
				domain: "backend_db",
				status: "published",
				definition: input.definition,
				userState: {
					itemId: id,
					sourceType: "manual",
					sourceId: input.example || undefined,
					mastered: false,
					createdAt: "2026-09-23T09:00:00Z",
				},
			},
		],
	}));
}
