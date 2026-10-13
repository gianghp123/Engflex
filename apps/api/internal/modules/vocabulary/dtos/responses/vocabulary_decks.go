package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

// VocabularyDeckDetail is the GET /vocabulary-decks/:id shape: deck header
// plus its category. Items page separately via /vocabulary-deck-items.
type VocabularyDeckDetail struct {
	ID           string                      `json:"id"`
	CategoryID   *string                     `json:"categoryId"`
	Name         string                      `json:"name"`
	Description  string                      `json:"description"`
	ThumbnailURL string                      `json:"thumbnailUrl"`
	Level        string                      `json:"level"`
	IsDefault    bool                        `json:"isDefault"`
	Status       enums.PublishStatus         `json:"status"`
	Category     *VocabularyCategoryResponse `json:"category"`
	CreatedAt    time.Time                   `json:"createdAt"`
	UpdatedAt    time.Time                   `json:"updatedAt"`
}

// VocabularyDeckItemResponse is one flashcard row.
type VocabularyDeckItemResponse struct {
	ID                string  `json:"id"`
	DeckID            string  `json:"deckId"`
	VideoExerciseID   *string `json:"videoExerciseId"`
	VideoTranscriptID *string `json:"videoTranscriptId"`
	Phrase            string  `json:"phrase"`
	NormalizedPhrase  string  `json:"normalizedPhrase"`
	Meaning           string  `json:"meaning"`
	ExampleSentence   string  `json:"exampleSentence"`
	Note              string  `json:"note"`
}
