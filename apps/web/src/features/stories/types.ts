import type { CEFR } from "@engflex/contracts";

export const STORIES_PAGE_SIZE = 12;

export interface StoriesSearch {
	tab?: "all" | "saved";
	cefr?: "all" | CEFR;
	topic?: string;
	search?: string;
}

export interface StoryFilters {
	cefr: "all" | CEFR;
	topic: string;
	search: string;
}
