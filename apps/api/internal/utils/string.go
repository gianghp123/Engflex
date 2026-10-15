package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// Slugify derives a URL-safe slug from free text: lowercase alphanumerics,
// everything else collapsed to dashes, with leading/trailing dashes trimmed.
func Slugify(text string) string {
	slug := strings.ToLower(text)
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, slug)
	return strings.Trim(slug, "-")
}

// NewSuffixedSlug derives a URL slug from free text with a random hex suffix
// so it satisfies UNIQUE constraints without a retry loop.
func NewSuffixedSlug(text string) string {
	slug := Slugify(text)
	var suffix [3]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return slug
	}
	return slug + "-" + hex.EncodeToString(suffix[:])
}

// TidyTerm collapses internal whitespace and trims — the form we store. Case is
// preserved: the term is display text ("iPhone" stays "iPhone").
func TidyTerm(term string) string {
	return strings.Join(strings.Fields(term), " ")
}

// TermKey is a term's identity for lookup and uniqueness: TidyTerm, case-folded.
// It must stay byte-identical to the generated vocabulary_items.term_key
// expression — TestTidyTermAndTermKey and the vocabulary migration's real-DB
// check pin the same table.
func TermKey(term string) string {
	return strings.ToLower(TidyTerm(term))
}
