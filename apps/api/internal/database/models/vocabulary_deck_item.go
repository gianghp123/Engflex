package models

import "time"

// VocabularyDeckItem is one flashcard row inside a deck: a reference to a
// shared dictionary entry (vocabulary_items) plus per-row presentation
// (sequence, example sentence, note).
type VocabularyDeckItem struct {
	ID              string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DeckID          string          `gorm:"type:uuid;not null" json:"deckId"`
	ItemID          string          `gorm:"type:uuid;not null" json:"itemId"`
	Sequence        int             `gorm:"not null;default:0" json:"sequence"`
	ExampleSentence string          `json:"exampleSentence"`
	Note            string          `json:"note"`
	CreatedAt       time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
	Deck            *VocabularyDeck `gorm:"->;foreignKey:DeckID;references:ID" json:"deck,omitempty"`
	Item            *VocabularyItem `gorm:"->;foreignKey:ItemID;references:ID" json:"item,omitempty"`
}

func (VocabularyDeckItem) TableName() string { return "vocabulary_deck_items" }
