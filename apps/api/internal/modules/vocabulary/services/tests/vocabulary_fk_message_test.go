package tests

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

func TestVocabularyDeckItemService_Create_UnknownDeckNamesDeck(t *testing.T) {
	owner := "user_1"
	itemRepo := mocks.NewMockVocabularyDeckItemRepository(t)
	deckRepo := mocks.NewMockVocabularyDeckRepository(t)
	deckRepo.On("GetByID", mock.Anything, deckID).
		Return(&models.VocabularyDeck{ID: deckID, UserID: owner}, nil)
	itemRepo.On("Create", mock.Anything, mock.Anything).
		Return(&pgconn.PgError{Code: "23503"})
	svc := services.NewVocabularyDeckItemService(itemRepo, deckRepo)

	res, err := svc.Create(context.Background(), owner, false, requests.CreateVocabularyDeckItem{DeckID: deckID, Phrase: "hi", Meaning: "chao"})

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "referenced vocabulary deck not found")
}
