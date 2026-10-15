package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

func TestVocabularyDeckService_Update_Forbidden(t *testing.T) {
	owner := "user_1"
	other := "user_2"
	deckID := "11111111-1111-1111-1111-111111111111"
	mockRepo := mocks.NewMockVocabularyDeckRepository(t)
	mockRepo.On("GetByID", mock.Anything, deckID).Return(&models.VocabularyDeck{ID: deckID, UserID: &owner, Name: "Core"}, nil)
	svc := services.NewVocabularyDeckService(mockRepo, mocks.NewMockVocabularyDeckItemRepository(t))

	res, err := svc.Update(context.Background(), other, deckID, requests.UpdateVocabularyDeck{})

	requireAppError(t, err, http.StatusForbidden)
	assert.Nil(t, res)
}

func TestVocabularyDeckItemService_List_EmptyReturns200(t *testing.T) {
	deckID := "11111111-1111-1111-1111-111111111111"
	owner := "user_1"
	svc, itemRepo, deckRepo, _ := setupDeckItem(t)
	deckRepo.On("GetByID", mock.Anything, deckID).
		Return(&models.VocabularyDeck{ID: deckID, UserID: &owner}, nil)
	itemRepo.On("ListByDeck", mock.Anything, owner, deckID).Return(make([]*models.VocabularyDeckItem, 0), nil)

	res, err := svc.List(context.Background(), owner, deckID)

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Empty(t, res)
}

func TestVocabularyDeckItemService_Create_BadDeckID(t *testing.T) {
	svc, _, deckRepo, _ := setupDeckItem(t)
	deckRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, &common.InvalidIDError{Name: "deckId"})

	res, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyDeckItem{
		DeckID: "11111111-1111-1111-1111-111111111111",
		ItemID: "22222222-2222-2222-2222-222222222222",
	})

	requireAppError(t, err, http.StatusBadRequest)
	assert.Nil(t, res)
}

func TestUserVocabularyService_Save_Idempotent(t *testing.T) {
	userID := "user_1"
	itemID := "11111111-1111-1111-1111-111111111111"
	row := &models.UserVocabulary{UserID: userID, ItemID: itemID}
	repo := mocks.NewMockUserVocabularyRepository(t)
	// Re-saving hits the Upsert conflict path; mastered is untouched when omitted.
	repo.On("Upsert", mock.Anything, mock.Anything, false).Return(nil)
	repo.On("Get", mock.Anything, userID, itemID).Return(row, nil)
	entries := mocks.NewMockVocabularyItemRepository(t)
	entries.On("GetByID", mock.Anything, userID, itemID).Return(&models.VocabularyItem{ID: itemID, Term: "bank"}, nil)
	svc := services.NewUserVocabularyService(repo, entries)

	for i := 0; i < 2; i++ {
		got, err := svc.Save(context.Background(), userID, requests.SaveUserVocabulary{ItemID: itemID, SourceType: "manual"})
		require.NoError(t, err)
		assert.Equal(t, itemID, got.ItemID)
	}
}

func TestUserVocabularyService_Save_BadItemID(t *testing.T) {
	repo := mocks.NewMockUserVocabularyRepository(t)
	repo.On("Upsert", mock.Anything, mock.Anything, false).Return(&common.InvalidIDError{Name: "itemId"})
	entries := mocks.NewMockVocabularyItemRepository(t)
	entries.On("GetByID", mock.Anything, "user_1", "11111111-1111-1111-1111-111111111111").
		Return(&models.VocabularyItem{ID: "11111111-1111-1111-1111-111111111111", Term: "bank"}, nil)
	svc := services.NewUserVocabularyService(repo, entries)

	res, err := svc.Save(context.Background(), "user_1", requests.SaveUserVocabulary{ItemID: "11111111-1111-1111-1111-111111111111", SourceType: "manual"})

	requireAppError(t, err, http.StatusBadRequest)
	assert.Nil(t, res)
}

func TestVocabularyItemService_Create_Conflict(t *testing.T) {
	tests := []struct {
		name      string
		createErr error
		status    int
	}{
		{"duplicate term conflicts", &pgconn.PgError{Code: "23505"}, http.StatusConflict},
		{"first insert succeeds", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockVocabularyItemRepository(t)
			repo.On("Create", mock.Anything, mock.Anything).Return(tt.createErr)
			svc := services.NewVocabularyItemService(repo, mocks.NewMockUserVocabularyRepository(t), mocks.NewMockVocabularyDeckRepository(t), mocks.NewMockVocabularyDeckItemRepository(t))

			got, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyItem{Term: "bank", CEFR: "B1"})

			if tt.status == 0 {
				require.NoError(t, err)
				require.NotNil(t, got)
			} else {
				requireAppError(t, err, tt.status)
				assert.Nil(t, got)
			}
		})
	}
}
