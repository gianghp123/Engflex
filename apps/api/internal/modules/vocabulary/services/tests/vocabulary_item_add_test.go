package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

func TestVocabularyItemService_AddWord(t *testing.T) {
	itemID := "22222222-2222-2222-2222-222222222222"

	tests := []struct {
		name         string
		setupMock    func(items *mocks.MockVocabularyItemRepository, user *mocks.MockUserVocabularyRepository)
		wantCreated  bool
		wantEnrolled bool
	}{
		{
			name:         "new term: creates the private entry then enrols it",
			wantCreated:  true,
			wantEnrolled: true,
			setupMock: func(items *mocks.MockVocabularyItemRepository, user *mocks.MockUserVocabularyRepository) {
				items.On("GetPrivateByTermKey", mock.Anything, "user_1", "serendipity").Return(nil, gorm.ErrRecordNotFound)
				items.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
					a.Get(1).(*models.VocabularyItem).ID = itemID
				}).Return(nil)
				user.On("Get", mock.Anything, "user_1", itemID).Return(nil, gorm.ErrRecordNotFound)
				user.On("CountByUser", mock.Anything, "user_1").Return(int64(0), nil)
				user.On("Upsert", mock.Anything, mock.Anything, false).Return(nil)
				items.On("GetByID", mock.Anything, "user_1", itemID).Return(&models.VocabularyItem{ID: itemID, Term: "serendipity"}, nil)
			},
		},
		{
			name:         "existing entry not yet saved: enrols without creating a duplicate",
			wantEnrolled: true,
			setupMock: func(items *mocks.MockVocabularyItemRepository, user *mocks.MockUserVocabularyRepository) {
				items.On("GetPrivateByTermKey", mock.Anything, "user_1", "serendipity").Return(&models.VocabularyItem{ID: itemID, Term: "serendipity"}, nil)
				user.On("Get", mock.Anything, "user_1", itemID).Return(nil, gorm.ErrRecordNotFound)
				user.On("CountByUser", mock.Anything, "user_1").Return(int64(0), nil)
				user.On("Upsert", mock.Anything, mock.Anything, false).Return(nil)
				items.On("GetByID", mock.Anything, "user_1", itemID).Return(&models.VocabularyItem{ID: itemID, Term: "serendipity"}, nil)
			},
		},
		{
			name: "already enrolled: skips the write so note and source survive",
			setupMock: func(items *mocks.MockVocabularyItemRepository, user *mocks.MockUserVocabularyRepository) {
				note := "my note"
				items.On("GetPrivateByTermKey", mock.Anything, "user_1", "serendipity").Return(&models.VocabularyItem{ID: itemID, Term: "serendipity"}, nil)
				user.On("Get", mock.Anything, "user_1", itemID).Return(&models.UserVocabulary{UserID: "user_1", ItemID: itemID, Note: &note}, nil)
				items.On("GetByID", mock.Anything, "user_1", itemID).Return(&models.VocabularyItem{ID: itemID, Term: "serendipity"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := mocks.NewMockVocabularyItemRepository(t)
			user := mocks.NewMockUserVocabularyRepository(t)
			tt.setupMock(items, user)
			svc := services.NewVocabularyItemService(items, user, mocks.NewMockVocabularyDeckRepository(t), mocks.NewMockVocabularyDeckItemRepository(t))

			got, err := svc.AddWord(context.Background(), "user_1", requests.AddUserVocabularyWord{Term: "serendipity", CEFR: "B2"})
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, itemID, got.ID)
			if tt.wantCreated {
				items.AssertCalled(t, "Create", mock.Anything, mock.Anything)
			} else {
				items.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			}
			if tt.wantEnrolled {
				user.AssertCalled(t, "Upsert", mock.Anything, mock.Anything, false)
			} else {
				user.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}
