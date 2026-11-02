package requests

import "engflex-api/internal/common/enums"

// CreateVocabularyDeck creates a flashcard deck (personal via
// /vocabulary-decks, or system-curated via the admin endpoint — the service
// method called decides which). IsPremium is only ever honored on the
// system-deck path (service-enforced); a personal deck ignores it.
type CreateVocabularyDeck struct {
	CategoryID   *string `json:"categoryId" binding:"omitempty,uuid"`
	Name         string  `json:"name" binding:"required,max=200"`
	Description  string  `json:"description"`
	ThumbnailURL string  `json:"thumbnailUrl"`
	Level        string  `json:"level"`
	IsPremium    bool    `json:"isPremium"`
}

// UpdateVocabularyDeck renames or re-categorises a deck. Status and
// IsPremium only take effect for a system deck updated by an admin
// (service-enforced).
type UpdateVocabularyDeck struct {
	CategoryID   *string              `json:"categoryId" binding:"omitempty,uuid"`
	Name         *string              `json:"name" binding:"omitempty,max=200"`
	Description  *string              `json:"description"`
	ThumbnailURL *string              `json:"thumbnailUrl"`
	Level        *string              `json:"level"`
	Status       *enums.PublishStatus `json:"status" binding:"omitempty,oneof=draft published"`
	IsPremium    *bool                `json:"isPremium"`
}

// CreateVocabularyDeckItem adds one flashcard row to a deck.
type CreateVocabularyDeckItem struct {
	DeckID          string `json:"deckId" binding:"required,uuid"`
	ItemID          string `json:"itemId" binding:"required,uuid"`
	Sequence        int    `json:"sequence" binding:"min=0"`
	ExampleSentence string `json:"exampleSentence"`
	Note            string `json:"note"`
}

// UpdateVocabularyDeckItem edits a flashcard row.
type UpdateVocabularyDeckItem struct {
	ItemID          *string `json:"itemId" binding:"omitempty,uuid"`
	Sequence        *int    `json:"sequence" binding:"omitempty,min=0"`
	ExampleSentence *string `json:"exampleSentence"`
	Note            *string `json:"note"`
}

// CreateVocabularyItem adds a shared dictionary entry or a custom word.
// IsPremium is only ever honored on the admin system-entry path
// (service-enforced); a personal custom word ignores it.
type CreateVocabularyItem struct {
	Details      *VocabularyDetails            `json:"details"`
	Term         string                        `json:"term" binding:"required"`
	Definition   string                        `json:"definition"`
	CEFR         enums.CEFR                    `json:"cefr" binding:"required,oneof=A1 A2 B1 B2 C1 C2"`
	Domain       *enums.VocabularyDomain       `json:"domain" tstype:"VocabularyDomain | null" binding:"omitempty,oneof=backend_db distributed_systems devops_cloud frontend_ui ai_ml"`
	PartOfSpeech *enums.VocabularyPartOfSpeech `json:"partOfSpeech" tstype:"VocabularyPartOfSpeech | null" binding:"omitempty,oneof=noun verb adjective adverb pronoun preposition conjunction interjection determiner phrase idiom phrasal_verb"`
	IsPremium    bool                          `json:"isPremium"`
}

// UpdateVocabularyItem edits a shared dictionary entry (admin-only, and
// only entries with no CreatedByUserID — a user's custom word stays theirs
// to manage, this never touches it).
type UpdateVocabularyItem struct {
	Details               *VocabularyDetails            `json:"details"`
	DomainSpecified       bool                          `json:"-"`
	PartOfSpeechSpecified bool                          `json:"-"`
	Term                  *string                       `json:"term"`
	Definition            *string                       `json:"definition"`
	CEFR                  *enums.CEFR                   `json:"cefr" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Domain                *enums.VocabularyDomain       `json:"domain" tstype:"VocabularyDomain | null" binding:"omitempty,oneof=backend_db distributed_systems devops_cloud frontend_ui ai_ml"`
	PartOfSpeech          *enums.VocabularyPartOfSpeech `json:"partOfSpeech" tstype:"VocabularyPartOfSpeech | null" binding:"omitempty,oneof=noun verb adjective adverb pronoun preposition conjunction interjection determiner phrase idiom phrasal_verb"`
	Status                *enums.PublishStatus          `json:"status" binding:"omitempty,oneof=draft published"`
	IsPremium             *bool                         `json:"isPremium"`
}

// SaveUserVocabulary saves per-user state over a shared item (idempotent).
type SaveUserVocabulary struct {
	ItemID     string                 `json:"itemId" binding:"required,uuid"`
	SourceType enums.VocabularySource `json:"sourceType" binding:"required,oneof=lesson conversation video manual story"`
	SourceID   *string                `json:"sourceId" binding:"omitempty,uuid"`
	Note       *string                `json:"note"`
	Mastered   *bool                  `json:"mastered"`
}
