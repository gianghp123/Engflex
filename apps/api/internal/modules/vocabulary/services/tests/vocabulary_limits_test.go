package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

// The product caps (common.MaxUserDecks / MaxDeckWords / MaxUserWords) are
// enforced at every creation mouth and only for genuinely new rows: re-adding
// a saved word, or a membership that already exists, stays allowed at the cap.

func TestVocabularyItemService_AddWord_WordLimit(t *testing.T) {
	itemID := "22222222-2222-2222-2222-222222222222"

	tests := []struct {
		name    string
		owned   int64
		saved   bool
		wantErr int
	}{
		{name: "new word under the cap enrols", owned: common.MaxUserWords - 1},
		{name: "new word at the cap is a 422", owned: common.MaxUserWords, wantErr: http.StatusUnprocessableEntity},
		{name: "a re-added saved word stays allowed at the cap", owned: common.MaxUserWords, saved: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := mocks.NewMockVocabularyItemRepository(t)
			user := mocks.NewMockUserVocabularyRepository(t)

			items.On("GetPrivateByTermKey", mock.Anything, "user_1", "serendipity").
				Return(&models.VocabularyItem{ID: itemID, Term: "serendipity"}, nil)
			if tt.wantErr == 0 {
				items.On("GetByID", mock.Anything, "user_1", itemID).
					Return(&models.VocabularyItem{ID: itemID, Term: "serendipity"}, nil)
			}
			if tt.saved {
				user.On("Get", mock.Anything, "user_1", itemID).
					Return(&models.UserVocabulary{UserID: "user_1", ItemID: itemID}, nil)
			} else {
				user.On("Get", mock.Anything, "user_1", itemID).Return(nil, gorm.ErrRecordNotFound)
				user.On("CountByUser", mock.Anything, "user_1").Return(tt.owned, nil)
				if tt.wantErr == 0 {
					user.On("Upsert", mock.Anything, mock.Anything, false).Return(nil)
				}
			}

			svc := services.NewVocabularyItemService(items, user,
				mocks.NewMockVocabularyDeckRepository(t), mocks.NewMockVocabularyDeckItemRepository(t))
			got, err := svc.AddWord(context.Background(), "user_1", requests.AddUserVocabularyWord{Term: "serendipity", CEFR: "B2"})
			if tt.wantErr != 0 {
				requireAppError(t, err, tt.wantErr)
				require.Nil(t, got)
				user.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			if !tt.saved {
				user.AssertCalled(t, "Upsert", mock.Anything, mock.Anything, false)
			}
		})
	}
}

func TestVocabularyDeckService_Create_DeckLimit(t *testing.T) {
	tests := []struct {
		name    string
		owned   int64
		wantErr int
	}{
		{name: "under the cap creates", owned: common.MaxUserDecks - 1},
		{name: "at the cap is a 422", owned: common.MaxUserDecks, wantErr: http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo := setupDeck(t)
			repo.On("CountByUser", mock.Anything, "user_1").Return(tt.owned, nil)
			if tt.wantErr == 0 {
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			}

			deck, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyDeck{Name: "Focus words"})
			if tt.wantErr != 0 {
				requireAppError(t, err, tt.wantErr)
				require.Nil(t, deck)
				repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
				return
			}
			require.NoError(t, err)
			require.Equal(t, "Focus words", deck.Name)
		})
	}
}

func TestVocabularyDeckItemService_Create_DeckWordLimit(t *testing.T) {
	owner := "user_1"
	deck := &models.VocabularyDeck{ID: deckID, UserID: &owner, Name: "Business"}
	entry := &models.VocabularyItem{ID: wordID, Term: "bank"}

	tests := []struct {
		name     string
		members  int64
		existing *models.VocabularyDeckItem
		wantErr  int
	}{
		{name: "under the cap adds", members: common.MaxDeckWords - 1},
		{name: "at the cap is a 422", members: common.MaxDeckWords, wantErr: http.StatusUnprocessableEntity},
		{name: "an existing membership is reused at the cap", members: common.MaxDeckWords, existing: &models.VocabularyDeckItem{ID: deckItem, DeckID: deckID, ItemID: wordID}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, itemRepo, deckRepo, entryRepo := setupDeckItem(t)
			deckRepo.On("GetByID", mock.Anything, deckID).Return(deck, nil)
			entryRepo.On("GetByID", mock.Anything, "user_1", wordID).Return(entry, nil)
			if tt.existing != nil {
				itemRepo.On("GetByDeckAndItem", mock.Anything, deckID, wordID).Return(tt.existing, nil)
			} else {
				itemRepo.On("GetByDeckAndItem", mock.Anything, deckID, wordID).Return(nil, gorm.ErrRecordNotFound)
				itemRepo.On("CountByDeck", mock.Anything, deckID).Return(tt.members, nil)
				if tt.wantErr == 0 {
					itemRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
				}
			}

			got, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyDeckItem{DeckID: deckID, ItemID: wordID})
			if tt.wantErr != 0 {
				requireAppError(t, err, tt.wantErr)
				require.Nil(t, got)
				itemRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
