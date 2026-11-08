import { CONTENT_TYPE_VALUES } from "@engflex/contracts";

export const NOTE_KINDS = ["vocab", ...CONTENT_TYPE_VALUES] as const;
export type NoteKind = (typeof NOTE_KINDS)[number];
export type NoteTabFilter = "all" | NoteKind;
