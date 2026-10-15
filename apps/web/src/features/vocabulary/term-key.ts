/**
 * A term's identity for lookup, mirroring `apps/api/internal/utils/string.go`
 * (`TermKey`): trim, collapse inner whitespace, case-fold — the same rule as the
 * generated `vocabulary_items.term_key` column.
 *
 * Advisory only. The API owns term identity and dedupes authoritatively when a
 * word is saved (`GetPrivateByTermKey` reuses the reader's own entry), so a
 * divergence here costs a missing or stale hint — never a duplicated word.
 *
 * Kept in the web rather than sent to the server because the only caller is a
 * local comparison against the library it already holds.
 */
export function termKey(term: string): string {
	return term.split(/\s+/).filter(Boolean).join(" ").toLowerCase();
}
