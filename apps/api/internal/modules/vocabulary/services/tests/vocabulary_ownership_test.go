package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"engflex-api/internal/database/models"
)

func TestVocabularyDeckItemService_List_ForeignDeck(t *testing.T) {
	owner := "user_1"
	tests := []struct {
		name    string
		deck    *models.VocabularyDeck
		message string
	}{
		{
			name:    "curated deck is read-only",
			deck:    &models.VocabularyDeck{ID: "11111111-1111-1111-1111-111111111111", UserID: nil, Name: "TOEIC 600"},
			message: "curated deck is read-only",
		},
		{
			name:    "foreign private deck is forbidden",
			deck:    &models.VocabularyDeck{ID: "11111111-1111-1111-1111-111111111111", UserID: &owner, Name: "Core"},
			message: "not your deck",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deckID := "11111111-1111-1111-1111-111111111111"
			svc, itemRepo, deckRepo, _ := setupDeckItem(t)
			deckRepo.On("GetByID", mock.Anything, deckID).Return(tt.deck, nil)

			res, err := svc.List(context.Background(), "user_9", deckID)

			appErr := requireAppError(t, err, http.StatusForbidden)
			assert.Equal(t, tt.message, appErr.Message)
			assert.Nil(t, res)
			itemRepo.AssertNotCalled(t, "ListByDeck", mock.Anything, mock.Anything)
		})
	}
}
