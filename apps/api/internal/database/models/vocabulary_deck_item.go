package models

import "time"

// VocabularyDeckItem is one flashcard row inside a user-owned deck (the
// former deck-side "vocabulary_items" shape, renamed). The shared dictionary
// entries live in vocabulary_items (VocabularyItem).
type VocabularyDeckItem struct {
	ID                string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DeckID            string          `gorm:"type:uuid;not null" json:"deckId"`
	VideoExerciseID   *string         `gorm:"type:uuid" json:"videoExerciseId"`
	VideoTranscriptID *string         `gorm:"type:uuid" json:"videoTranscriptId"`
	Phrase            string          `gorm:"not null" json:"phrase"`
	NormalizedPhrase  string          `gorm:"not null" json:"normalizedPhrase"`
	Meaning           string          `gorm:"not null" json:"meaning"`
	ExampleSentence   string          `json:"exampleSentence"`
	Note              string          `json:"note"`
	CreatedAt         time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt         time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
	Deck              *VocabularyDeck `gorm:"->;foreignKey:DeckID;references:ID" json:"deck,omitempty"`
}

func (VocabularyDeckItem) TableName() string { return "vocabulary_deck_items" }
