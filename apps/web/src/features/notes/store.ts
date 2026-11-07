import { useStore } from "@tanstack/react-store";
import { Store } from "@tanstack/store";
import { MOCK_NOTES } from "./fixtures";
import type { NoteItem, NoteTabFilter } from "./types";

export interface NotesStoreState {
	items: NoteItem[];
	activeTab: NoteTabFilter;
	searchQuery: string;
}

export const notesStore = new Store<NotesStoreState>({
	items: MOCK_NOTES,
	activeTab: "all",
	searchQuery: "",
});

export const notesActions = {
	setActiveTab: (tab: NoteTabFilter) => {
		notesStore.setState((state) => ({ ...state, activeTab: tab }));
	},
	setSearchQuery: (query: string) => {
		notesStore.setState((state) => ({ ...state, searchQuery: query }));
	},
	updatePersonalNote: (id: string, personalNote: string) => {
		notesStore.setState((state) => ({
			...state,
			items: state.items.map((item) =>
				item.id === id ? { ...item, personalNote } : item,
			),
		}));
	},
	removeItem: (id: string) => {
		notesStore.setState((state) => ({
			...state,
			items: state.items.filter((item) => item.id !== id),
		}));
	},
	addItem: (newItem: Omit<NoteItem, "id" | "createdAt">) => {
		const created: NoteItem = {
			...newItem,
			id: `note-${Date.now()}`,
			createdAt: new Date().toISOString().slice(0, 10),
		};
		notesStore.setState((state) => ({
			...state,
			items: [created, ...state.items],
		}));
	},
	resetFilters: () => {
		notesStore.setState((state) => ({
			...state,
			activeTab: "all",
			searchQuery: "",
		}));
	},
};

export function useNotes() {
	return useStore(notesStore, (state) => state.items);
}

export function useNotesFilter() {
	const activeTab = useStore(notesStore, (state) => state.activeTab);
	const searchQuery = useStore(notesStore, (state) => state.searchQuery);
	return { activeTab, searchQuery };
}

export function useFilteredNotes() {
	const items = useStore(notesStore, (state) => state.items);
	const activeTab = useStore(notesStore, (state) => state.activeTab);
	const searchQuery = useStore(notesStore, (state) => state.searchQuery);

	const trimmedQuery = searchQuery.trim().toLowerCase();

	return items.filter((item) => {
		if (activeTab !== "all" && item.kind !== activeTab) {
			return false;
		}

		if (!trimmedQuery) {
			return true;
		}

		const matchTitle = item.title.toLowerCase().includes(trimmedQuery);
		const matchNote = item.personalNote?.toLowerCase().includes(trimmedQuery);
		const matchSummary = item.summary?.toLowerCase().includes(trimmedQuery);
		const matchTags = item.tags.some((tag) =>
			tag.toLowerCase().includes(trimmedQuery),
		);

		return matchTitle || matchNote || matchSummary || matchTags;
	});
}

export function useNotesStats() {
	const items = useStore(notesStore, (state) => state.items);
	const total = items.length;
	const withNotes = items.filter(
		(item) => !!item.personalNote && item.personalNote.trim() !== "",
	).length;
	const vocabCount = items.filter((item) => item.kind === "vocab").length;
	const lessonCount = items.filter((item) => item.kind === "lesson").length;
	const videoCount = items.filter((item) => item.kind === "video").length;
	const storyCount = items.filter((item) => item.kind === "story").length;

	return {
		total,
		withNotes,
		vocabCount,
		lessonCount,
		videoCount,
		storyCount,
	};
}
