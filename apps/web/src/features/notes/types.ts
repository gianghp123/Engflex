export type NoteKind = "vocab" | "lesson" | "video" | "story";

export interface NoteItem {
	id: string;
	kind: NoteKind;
	title: string;
	subtitle?: string;
	summary?: string;
	personalNote?: string;
	tags: string[];
	createdAt: string;
	linkTo: string;
	thumbnail?: string;
	extraInfo?: {
		level?: string;
		duration?: string;
		badge?: string;
	};
}

export type NoteTabFilter = "all" | NoteKind;
