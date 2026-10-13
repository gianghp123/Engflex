package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/services"
)

func TestVocabularyDeckItemService_ListByDeck_ForeignDeck(t *testing.T) {
	owner := "user_1"
	other := "user_9"
	deckRepo := mocks.NewMockVocabularyDeckRepository(t)
	deckRepo.On("GetByID", mock.Anything, deckID).
		Return(&models.VocabularyDeck{ID: deckID, UserID: owner, Name: "Core"}, nil)
	itemRepo := mocks.NewMockVocabularyDeckItemRepository(t)
	svc := services.NewVocabularyDeckItemService(itemRepo, deckRepo)

	res, err := svc.ListByDeck(context.Background(), other, deckID)

	requireAppError(t, err, http.StatusForbidden)
	assert.Nil(t, res)
	itemRepo.AssertNotCalled(t, "ListByDeck", mock.Anything, mock.Anything)
}
