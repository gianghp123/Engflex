package requests

import "engflex-api/internal/common/enums"

// CreateVocabularyDeck creates a flashcard deck (personal via
// /vocabulary-decks, or system-curated via the admin endpoint — the service
// method called decides which).
type CreateVocabularyDeck struct {
	CategoryID   *string `json:"categoryId" binding:"omitempty,uuid"`
	Name         string  `json:"name" binding:"required,max=200"`
	Description  string  `json:"description"`
	ThumbnailURL string  `json:"thumbnailUrl"`
	Level        string  `json:"level"`
}

// UpdateVocabularyDeck renames or re-categorises a deck. Status only takes
// effect for a system deck updated by an admin (service-enforced).
type UpdateVocabularyDeck struct {
	CategoryID   *string              `json:"categoryId" binding:"omitempty,uuid"`
	Name         *string              `json:"name" binding:"omitempty,max=200"`
	Description  *string              `json:"description"`
	ThumbnailURL *string              `json:"thumbnailUrl"`
	Level        *string              `json:"level"`
	Status       *enums.PublishStatus `json:"status" binding:"omitempty,oneof=draft published"`
}

// CreateVocabularyDeckItem adds one flashcard row to a deck.
type CreateVocabularyDeckItem struct {
	DeckID            string  `json:"deckId" binding:"required,uuid"`
	VideoExerciseID   *string `json:"videoExerciseId" binding:"omitempty,uuid"`
	VideoTranscriptID *string `json:"videoTranscriptId" binding:"omitempty,uuid"`
	Phrase            string  `json:"phrase" binding:"required"`
	Meaning           string  `json:"meaning" binding:"required"`
	ExampleSentence   string  `json:"exampleSentence"`
	Note              string  `json:"note"`
}

// UpdateVocabularyDeckItem edits a flashcard row.
type UpdateVocabularyDeckItem struct {
	Phrase          *string `json:"phrase"`
	Meaning         *string `json:"meaning"`
	ExampleSentence *string `json:"exampleSentence"`
	Note            *string `json:"note"`
}

// CreateVocabularyItem adds a shared dictionary entry or a custom word.
type CreateVocabularyItem struct {
	Term       string  `json:"term" binding:"required"`
	Definition string  `json:"definition"`
	CEFR       string  `json:"cefr" binding:"required,oneof=A1 A2 B1 B2 C1 C2"`
	Domain     *string `json:"domain"`
}

// UpdateVocabularyItem edits a shared dictionary entry (admin-only, and
// only entries with no CreatedByUserID — a user's custom word stays theirs
// to manage, this never touches it).
type UpdateVocabularyItem struct {
	Term       *string              `json:"term"`
	Definition *string              `json:"definition"`
	CEFR       *string              `json:"cefr" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Domain     *string              `json:"domain"`
	Status     *enums.PublishStatus `json:"status" binding:"omitempty,oneof=draft published"`
}

// SaveUserVocabulary saves per-user state over a shared item (idempotent).
type SaveUserVocabulary struct {
	ItemID     string  `json:"itemId" binding:"required,uuid"`
	SourceType string  `json:"sourceType" binding:"required,oneof=lesson conversation manual"`
	SourceID   *string `json:"sourceId" binding:"omitempty,uuid"`
	Note       *string `json:"note"`
	Mastered   *bool   `json:"mastered"`
}
