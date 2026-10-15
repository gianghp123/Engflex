package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

func TestDeckItemService_CuratedDeckIsReadOnly(t *testing.T) {
	deckID := "11111111-1111-1111-1111-111111111111"
	decks := mocks.NewMockVocabularyDeckRepository(t)
	items := mocks.NewMockVocabularyDeckItemRepository(t)
	// Curated deck: user_id IS NULL.
	decks.On("GetByID", mock.Anything, deckID).Return(&models.VocabularyDeck{ID: deckID, UserID: nil, Name: "TOEIC 600"}, nil)
	svc := services.NewVocabularyDeckItemService(items, decks, mocks.NewMockVocabularyItemRepository(t))

	_, err := svc.List(context.Background(), "user_1", deckID)
	requireAppError(t, err, http.StatusForbidden)
}

func TestDeckItemService_Create_RejectsInvisibleItem(t *testing.T) {
	deckID := "11111111-1111-1111-1111-111111111111"
	itemID := "22222222-2222-2222-2222-222222222222"
	owner := "user_1"
	decks := mocks.NewMockVocabularyDeckRepository(t)
	items := mocks.NewMockVocabularyDeckItemRepository(t)
	entries := mocks.NewMockVocabularyItemRepository(t)
	decks.On("GetByID", mock.Anything, deckID).Return(&models.VocabularyDeck{ID: deckID, UserID: &owner}, nil)
	entries.On("GetByID", mock.Anything, owner, itemID).Return(nil, gormErrNotFound())
	svc := services.NewVocabularyDeckItemService(items, decks, entries)

	_, err := svc.Create(context.Background(), owner, requests.CreateVocabularyDeckItem{DeckID: deckID, ItemID: itemID})
	requireAppError(t, err, http.StatusNotFound)
	assert.Equal(t, 0, len(items.Calls))
}
