package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
)

// The per-field 404 contract lives in the pre-write checks: the deck lookup
// names a missing deck, the entry lookup names a missing entry. The insert's
// own FK fallback stays generic and names no side.
func TestVocabularyDeckItemService_Create_MissingSideNamed(t *testing.T) {
	deckID := "11111111-1111-1111-1111-111111111111"
	itemID := "22222222-2222-2222-2222-222222222222"
	owner := "user_1"

	tests := []struct {
		name      string
		setupMock func(decks *mocks.MockVocabularyDeckRepository, entries *mocks.MockVocabularyItemRepository)
		message   string
	}{
		{
			name: "unknown deck names the deck",
			setupMock: func(decks *mocks.MockVocabularyDeckRepository, entries *mocks.MockVocabularyItemRepository) {
				decks.On("GetByID", mock.Anything, deckID).
					Return(nil, &pgconn.PgError{Code: "23503"})
			},
			message: "referenced vocabulary deck not found",
		},
		{
			name: "unknown item names the item",
			setupMock: func(decks *mocks.MockVocabularyDeckRepository, entries *mocks.MockVocabularyItemRepository) {
				decks.On("GetByID", mock.Anything, deckID).
					Return(&models.VocabularyDeck{ID: deckID, UserID: &owner}, nil)
				entries.On("GetByID", mock.Anything, owner, itemID).
					Return(nil, &pgconn.PgError{Code: "23503"})
			},
			message: "referenced vocabulary item not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, itemRepo, deckRepo, entryRepo := setupDeckItem(t)
			tt.setupMock(deckRepo, entryRepo)

			res, err := svc.Create(context.Background(), owner, requests.CreateVocabularyDeckItem{DeckID: deckID, ItemID: itemID})

			appErr := requireAppError(t, err, http.StatusNotFound)
			assert.Equal(t, tt.message, appErr.Message)
			assert.Nil(t, res)
			itemRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		})
	}
}
