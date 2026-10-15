package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

// TestVocabularyDeckItemService_Create_Idempotent pins that re-adding a word to
// the same deck returns the existing row instead of surfacing the unique-index
// 409 (uq_vocabulary_deck_items_deck_item).
func TestVocabularyDeckItemService_Create_Idempotent(t *testing.T) {
	items := mocks.NewMockVocabularyDeckItemRepository(t)
	decks := mocks.NewMockVocabularyDeckRepository(t)
	entries := mocks.NewMockVocabularyItemRepository(t)

	uid := "user_1"
	decks.On("GetByID", mock.Anything, deckID).Return(&models.VocabularyDeck{ID: deckID, UserID: &uid}, nil)
	entries.On("GetByID", mock.Anything, "user_1", wordID).Return(&models.VocabularyItem{ID: wordID}, nil)
	existing := &models.VocabularyDeckItem{ID: deckItem, DeckID: deckID, ItemID: wordID}
	items.On("GetByDeckAndItem", mock.Anything, deckID, wordID).Return(existing, nil)

	svc := services.NewVocabularyDeckItemService(items, decks, entries)
	got, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyDeckItem{DeckID: deckID, ItemID: wordID})
	require.NoError(t, err)
	require.Equal(t, deckItem, got.ID)
	items.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}
