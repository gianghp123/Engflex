package llm

import (
	"slices"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// LanguageParam is the reusable explanation-language request fragment embedded
// by every request DTO that triggers an LLM call (word draft, writing score,
// turn analysis, session start). One field, one wire key, one binding rule —
// the alternative is the same tag copy-pasted into each request, which drifts.
//
// The value is a BCP-47 code ("en", "vi", …) naming the language the
// explanation fields are written in; empty means the default brief. DTOs embed
// it with `tstype:",extends"` so tygo emits `interface X extends LanguageParam`.
// Validation is owned by ExplanationLanguage below, which rejects codes no
// model can write in — the binding tag only bounds the wire.
type LanguageParam struct {
	Language string `json:"language,omitempty" binding:"omitempty,max=16"`
}

// ExplanationLanguage resolves a request's BCP-47 code to the English name of
// the language the explanation fields are written in ("" = English, the
// default brief). Every LLM-triggering request resolves its code here and
// passes the *name* into its prompt, so no prompt needs its own tag table and
// the model never sees a raw tag.
//
// Any assigned language is accepted — the web picker loads its options from
// the app's configured locales, so adding a locale needs no change here.
// Unassigned tags ("zz"), malformed tags ("not a tag") and the non-language
// codes are refused: better a 400 than a brief the model cannot interpret.
func ExplanationLanguage(code string) (string, bool) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", true
	}
	tag, err := language.Parse(code)
	if err != nil || slices.Contains(nonLanguageTags, tag) {
		return "", false
	}
	name := display.English.Languages().Name(tag)
	if name == "" || strings.EqualFold(name, code) {
		return "", false
	}
	return name, true
}

// nonLanguageTags are well-formed codes that name no single language; the
// model cannot write explanations in them.
var nonLanguageTags = []language.Tag{
	language.MustParse("und"), // Unknown language
	language.MustParse("mul"), // Multiple languages
	language.MustParse("zxx"), // No linguistic content
}
