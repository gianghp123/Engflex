package requests

import (
	"engflex-api/internal/llm"
)

// VocabularyDetails is the wire copy of the entry's rich content (tygo cannot
// resolve cross-package types).
type VocabularyDetails struct {
	IPA            string              `json:"ipa"`
	AudioUS        string              `json:"audioUs"`
	AudioUK        string              `json:"audioUk"`
	Syllables      string              `json:"syllables"`
	StressTip      string              `json:"stressTip"`
	CommonSlip     string              `json:"commonSlip"`
	LongDefinition string              `json:"longDefinition"`
	Etymology      string              `json:"etymology"`
	Contexts       []VocabularyContext `json:"contexts"`
	Collocations   []Collocation       `json:"collocations"`
	WordForms      []WordForm          `json:"wordForms"`
}

// VocabularyContext is one labeled usage quote.
type VocabularyContext struct {
	Label string `json:"label"`
	Quote string `json:"quote"`
}

// Collocation is a fixed phrase with its pattern.
type Collocation struct {
	Phrase  string `json:"phrase"`
	Pattern string `json:"pattern"`
	Example string `json:"example"`
}

// WordForm is one morphology entry.
type WordForm struct {
	FormType string   `json:"formType"`
	Forms    []string `json:"forms"`
}

// CreateVocabularyItem adds a private (user-owned) dictionary entry.
type CreateVocabularyItem struct {
	Term         string             `json:"term" binding:"required"`
	Definition   string             `json:"definition"`
	CEFR         string             `json:"cefr" binding:"required,oneof=A1 A2 B1 B2 C1 C2"`
	PartOfSpeech *string            `json:"partOfSpeech" binding:"omitempty,oneof=noun verb adjective adverb pronoun preposition conjunction interjection determiner phrase idiom phrasal_verb"`
	Details      *VocabularyDetails `json:"details"`
}

// CreateVocabularyDeck creates a user-owned flashcard deck.
type CreateVocabularyDeck struct {
	CategoryID   *string `json:"categoryId" binding:"omitempty,uuid"`
	Name         string  `json:"name" binding:"required,max=200"`
	Description  string  `json:"description"`
	ThumbnailURL string  `json:"thumbnailUrl"`
	Level        string  `json:"level"`
}

// UpdateVocabularyDeck renames or re-categorises a deck.
type UpdateVocabularyDeck struct {
	CategoryID   *string `json:"categoryId" binding:"omitempty,uuid"`
	Name         *string `json:"name" binding:"omitempty,max=200"`
	Description  *string `json:"description"`
	ThumbnailURL *string `json:"thumbnailUrl"`
	Level        *string `json:"level"`
}

// CreateVocabularyDeckItem adds one flashcard row to a deck: a reference to
// a visible dictionary entry plus per-row presentation.
type CreateVocabularyDeckItem struct {
	DeckID          string `json:"deckId" binding:"required,uuid"`
	ItemID          string `json:"itemId" binding:"required,uuid"`
	Sequence        int    `json:"sequence"`
	ExampleSentence string `json:"exampleSentence"`
	Note            string `json:"note"`
}

// UpdateVocabularyDeckItem edits a flashcard row's presentation.
type UpdateVocabularyDeckItem struct {
	ExampleSentence *string `json:"exampleSentence"`
	Note            *string `json:"note"`
}

// SaveUserVocabulary saves per-user state over a shared item (idempotent).
type SaveUserVocabulary struct {
	ItemID     string  `json:"itemId" binding:"required,uuid"`
	SourceType string  `json:"sourceType" binding:"required,oneof=lesson conversation video manual"`
	SourceID   *string `json:"sourceId" binding:"omitempty,uuid"`
	Note       *string `json:"note"`
	Mastered   *bool   `json:"mastered"`
}

// AddUserVocabularyWord adds a word to the caller's library by content: the
// server find-or-creates a private dictionary entry (unique per user by
// lower(term)) and enrols it. Idempotent — re-adding an existing term
// re-enrols it rather than 409ing.
type AddUserVocabularyWord struct {
	Term         string  `json:"term" binding:"required"`
	Definition   string  `json:"definition"`
	CEFR         string  `json:"cefr" binding:"required,oneof=A1 A2 B1 B2 C1 C2"`
	PartOfSpeech *string `json:"partOfSpeech" binding:"omitempty,oneof=noun verb adjective adverb pronoun preposition conjunction interjection determiner phrase idiom phrasal_verb"`
	// DeckID optionally enrols the word and drops it straight into one of the
	// caller's private decks — the deck page's "add a word" path.
	DeckID  *string            `json:"deckId" binding:"omitempty,uuid"`
	Details *VocabularyDetails `json:"details"`
	// SourceType/SourceID record where the word came from when the save was
	// started from content (the lesson/video context dock). Omitted = manual.
	SourceType string  `json:"sourceType" binding:"omitempty,oneof=lesson conversation video manual"`
	SourceID   *string `json:"sourceId" binding:"omitempty,uuid"`
}

// PatchUserVocabulary updates one saved word's mutable state. Both fields are
// optional pointers: omitted means "leave unchanged", so a PATCH never clears
// a note or un-masters a word by accident. An empty body is rejected by the
// service.
type PatchUserVocabulary struct {
	Mastered *bool   `json:"mastered"`
	Note     *string `json:"note"`
}

// ReviewUserVocabulary grades one review (0 = again … 3 = easy). Grade is a
// pointer so that a missing grade fails `required`: with a plain int the zero
// value is indistinguishable from "again" and `{}` would bind successfully.
type ReviewUserVocabulary struct {
	Grade *int `json:"grade" binding:"required,oneof=0 1 2 3"`
}

// GenerateWordDraft asks for an LLM-authored word draft for the add-word
// dialog. Read-only: nothing is stored; the reviewed result rides the normal
// add-word save.
type GenerateWordDraft struct {
	llm.LanguageParam `tstype:",extends"`
	Term              string `json:"term" binding:"required,max=120"`
	Context           string `json:"context" binding:"max=400"`
}
