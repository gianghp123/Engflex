package responses

import "time"

// VocabularyDeckDetail is the GET /vocabulary-decks/:id shape: deck header
// plus its category and rows.
type VocabularyDeckDetail struct {
	ID           string                       `json:"id"`
	UserID       *string                      `json:"userId"`
	CategoryID   *string                      `json:"categoryId"`
	Name         string                       `json:"name"`
	Description  string                       `json:"description"`
	ThumbnailURL string                       `json:"thumbnailUrl"`
	Level        string                       `json:"level"`
	Category     *VocabularyCategoryResponse  `json:"category"`
	Items        []VocabularyDeckItemResponse `json:"items,omitempty"`
	CreatedAt    time.Time                    `json:"createdAt"`
	UpdatedAt    time.Time                    `json:"updatedAt"`
}

// VocabularyDeckItemResponse is one flashcard row with its entry attached.
type VocabularyDeckItemResponse struct {
	ID              string          `json:"id"`
	DeckID          string          `json:"deckId"`
	ItemID          string          `json:"itemId"`
	Sequence        int             `json:"sequence"`
	ExampleSentence string          `json:"exampleSentence"`
	Note            string          `json:"note"`
	Item            *VocabularyItem `json:"item,omitempty"`
}
