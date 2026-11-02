package tests

import (
	"context"
	"engflex-api/internal/common/enums"
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
	dictionary := mocks.NewMockVocabularyItemRepository(t)
	dictionary.On("GetByID", mock.Anything, mock.Anything).Return(&models.VocabularyItem{ID: "11111111-1111-1111-1111-111111111111", Status: enums.PublishStatusPublished}, nil).Maybe()
	itemRepo.On("GetByID", mock.Anything, mock.Anything).Return(&models.VocabularyDeckItem{DeckID: deckID, ItemID: "11111111-1111-1111-1111-111111111111", Item: &models.VocabularyItem{Status: enums.PublishStatusPublished}}, nil).Maybe()
	svc := services.NewVocabularyDeckItemService(itemRepo, deckRepo, dictionary, fakePremiumChecker{})

	res, err := svc.Create(context.Background(), owner, false, requests.CreateVocabularyDeckItem{DeckID: deckID, ItemID: "11111111-1111-1111-1111-111111111111"})

	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "referenced vocabulary deck not found")
}
