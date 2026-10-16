package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

func TestVocabularyDeckService_Update_Forbidden(t *testing.T) {
	owner := "user_1"
	other := "user_2"
	mockRepo := mocks.NewMockVocabularyDeckRepository(t)
	mockRepo.On("GetByID", mock.Anything, deckID).Return(&models.VocabularyDeck{ID: deckID, UserID: owner, Name: "Core"}, nil)
	svc := services.NewVocabularyDeckService(mockRepo)

	res, err := svc.Update(context.Background(), other, deckID, false, requests.UpdateVocabularyDeck{})

	requireAppError(t, err, http.StatusForbidden)
	assert.Nil(t, res)
}

func TestVocabularyDeckItemService_ListByDeck_EmptyReturns200(t *testing.T) {
	owner := "user_1"
	itemRepo := mocks.NewMockVocabularyDeckItemRepository(t)
	deckRepo := mocks.NewMockVocabularyDeckRepository(t)
	deckRepo.On("GetByID", mock.Anything, deckID).
		Return(&models.VocabularyDeck{ID: deckID, UserID: owner}, nil)
	itemRepo.On("ListByDeck", mock.Anything, deckID).Return(make([]*models.VocabularyDeckItem, 0), nil)
	svc := services.NewVocabularyDeckItemService(itemRepo, deckRepo)

	res, err := svc.ListByDeck(context.Background(), owner, deckID, false)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Empty(t, res)
}

func TestVocabularyDeckItemService_Create_BadDeckID(t *testing.T) {
	itemRepo := mocks.NewMockVocabularyDeckItemRepository(t)
	deckRepo := mocks.NewMockVocabularyDeckRepository(t)
	deckRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, &common.InvalidIDError{Name: "deckId"})
	svc := services.NewVocabularyDeckItemService(itemRepo, deckRepo)

	res, err := svc.Create(context.Background(), "user_1", false, requests.CreateVocabularyDeckItem{DeckID: "11111111-1111-1111-1111-111111111111", Phrase: "hello", Meaning: "xin chao"})

	requireAppError(t, err, http.StatusBadRequest)
	assert.Nil(t, res)
}

func TestUserVocabularyService_Save_Idempotent(t *testing.T) {
	userID := "user_1"
	itemID := "11111111-1111-1111-1111-111111111111"
	row := &models.UserVocabulary{UserID: userID, ItemID: itemID}
	repo := mocks.NewMockUserVocabularyRepository(t)
	itemRepo := mocks.NewMockVocabularyItemRepository(t)
	itemRepo.On("GetByID", mock.Anything, itemID).Return(&models.VocabularyItem{ID: itemID, Status: enums.PublishStatusPublished}, nil)
	repo.On("Upsert", mock.Anything, mock.Anything).Return(nil)
	repo.On("Get", mock.Anything, userID, itemID).Return(row, nil)
	svc := services.NewUserVocabularyService(repo, itemRepo)

	for i := 0; i < 2; i++ {
		got, err := svc.Save(context.Background(), userID, requests.SaveUserVocabulary{ItemID: itemID, SourceType: "manual"})
		require.NoError(t, err)
		assert.Equal(t, itemID, got.ItemID)
	}
}

func TestUserVocabularyService_Save_BadItemID(t *testing.T) {
	repo := mocks.NewMockUserVocabularyRepository(t)
	itemRepo := mocks.NewMockVocabularyItemRepository(t)
	itemRepo.On("GetByID", mock.Anything, "11111111-1111-1111-1111-111111111111").
		Return(nil, &common.InvalidIDError{Name: "itemId"})
	svc := services.NewUserVocabularyService(repo, itemRepo)

	res, err := svc.Save(context.Background(), "user_1", requests.SaveUserVocabulary{ItemID: "11111111-1111-1111-1111-111111111111", SourceType: "manual"})

	requireAppError(t, err, http.StatusBadRequest)
	assert.Nil(t, res)
}

func TestUserVocabularyService_Save_DraftItemRejected(t *testing.T) {
	userID := "user_1"
	itemID := "11111111-1111-1111-1111-111111111111"
	repo := mocks.NewMockUserVocabularyRepository(t)
	itemRepo := mocks.NewMockVocabularyItemRepository(t)
	itemRepo.On("GetByID", mock.Anything, itemID).Return(&models.VocabularyItem{ID: itemID, Status: enums.PublishStatusDraft}, nil)
	svc := services.NewUserVocabularyService(repo, itemRepo)

	res, err := svc.Save(context.Background(), userID, requests.SaveUserVocabulary{ItemID: itemID, SourceType: "manual"})

	requireAppError(t, err, http.StatusUnprocessableEntity)
	assert.Nil(t, res)
}
