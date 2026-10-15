package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

func TestNextIntervalDays(t *testing.T) {
	tests := []struct {
		name        string
		current     int
		reviewCount int
		grade       int
		want        int
	}{
		{"again resets to 0", 16, 3, 0, 0},
		{"good from 0 -> 1", 0, 0, 2, 1},
		{"hard repeats current", 8, 2, 1, 8},
		{"good doubles", 4, 1, 2, 8},
		{"easy quadruples", 4, 1, 3, 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, services.NextIntervalDays(tt.current, tt.reviewCount, tt.grade))
		})
	}
}

func TestUserVocabularyService_Save_FirstInsertSetsSchedule(t *testing.T) {
	repo := mocks.NewMockUserVocabularyRepository(t)
	itemID := "11111111-1111-1111-1111-111111111111"
	var inserted *models.UserVocabulary
	repo.On("Upsert", mock.Anything, mock.Anything, false).Run(func(a mock.Arguments) {
		inserted = a.Get(1).(*models.UserVocabulary)
	}).Return(nil)
	repo.On("Get", mock.Anything, "user_1", itemID).Return(&models.UserVocabulary{UserID: "user_1", ItemID: itemID}, nil)
	entries := mocks.NewMockVocabularyItemRepository(t)
	entries.On("GetByID", mock.Anything, "user_1", itemID).Return(&models.VocabularyItem{ID: itemID, Term: "bank"}, nil)
	svc := services.NewUserVocabularyService(repo, entries)

	_, err := svc.Save(context.Background(), "user_1", requests.SaveUserVocabulary{ItemID: itemID, SourceType: "video"})
	require.NoError(t, err)
	require.NotNil(t, inserted.SRSDueAt, "first save must enrol the word with a due date")
	assert.Equal(t, "video", string(inserted.SourceType))
}

func TestUserVocabularyService_Save_OmittingMasteredDoesNotUpdateIt(t *testing.T) {
	repo := mocks.NewMockUserVocabularyRepository(t)
	itemID := "11111111-1111-1111-1111-111111111111"
	// updateMastered must be false when the body omits mastered.
	repo.On("Upsert", mock.Anything, mock.Anything, false).Return(nil)
	repo.On("Get", mock.Anything, "user_1", itemID).Return(&models.UserVocabulary{UserID: "user_1", ItemID: itemID, Mastered: true}, nil)
	entries := mocks.NewMockVocabularyItemRepository(t)
	entries.On("GetByID", mock.Anything, "user_1", itemID).Return(&models.VocabularyItem{ID: itemID, Term: "bank"}, nil)
	svc := services.NewUserVocabularyService(repo, entries)

	_, err := svc.Save(context.Background(), "user_1", requests.SaveUserVocabulary{ItemID: itemID, SourceType: "manual"})
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUserVocabularyService_Save_InvisibleItem(t *testing.T) {
	itemID := "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name      string
		lookupErr error
		status    int
	}{
		{"missing item returns 404", gormErrNotFound(), http.StatusNotFound},
		{"malformed id returns 400", &common.InvalidIDError{Name: "itemId"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockUserVocabularyRepository(t)
			entries := mocks.NewMockVocabularyItemRepository(t)
			entries.On("GetByID", mock.Anything, "user_1", itemID).Return(nil, tt.lookupErr)
			svc := services.NewUserVocabularyService(repo, entries)

			res, err := svc.Save(context.Background(), "user_1", requests.SaveUserVocabulary{ItemID: itemID, SourceType: "manual"})

			requireAppError(t, err, tt.status)
			assert.Nil(t, res)
			repo.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestUserVocabularyService_Patch(t *testing.T) {
	itemID := "11111111-1111-1111-1111-111111111111"
	note := "say it out loud"
	on := true

	tests := []struct {
		name       string
		req        requests.PatchUserVocabulary
		setupMock  func(m *mocks.MockUserVocabularyRepository)
		wantErr    bool
		wantStatus int
	}{
		{
			name: "mastered only",
			req:  requests.PatchUserVocabulary{Mastered: &on},
			setupMock: func(m *mocks.MockUserVocabularyRepository) {
				m.On("Patch", mock.Anything, "user_1", itemID, &on, (*string)(nil)).Return(nil)
				m.On("Get", mock.Anything, "user_1", itemID).Return(&models.UserVocabulary{UserID: "user_1", ItemID: itemID, Mastered: true}, nil)
			},
		},
		{
			name: "note only",
			req:  requests.PatchUserVocabulary{Note: &note},
			setupMock: func(m *mocks.MockUserVocabularyRepository) {
				m.On("Patch", mock.Anything, "user_1", itemID, (*bool)(nil), &note).Return(nil)
				m.On("Get", mock.Anything, "user_1", itemID).Return(&models.UserVocabulary{UserID: "user_1", ItemID: itemID, Note: &note}, nil)
			},
		},
		{
			name: "nothing to update is 400",
			req:  requests.PatchUserVocabulary{},
			setupMock: func(m *mocks.MockUserVocabularyRepository) {
				// no repo calls expected
			},
			wantErr:    true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "unknown word is 404",
			req:  requests.PatchUserVocabulary{Mastered: &on},
			setupMock: func(m *mocks.MockUserVocabularyRepository) {
				m.On("Patch", mock.Anything, "user_1", itemID, &on, (*string)(nil)).Return(gorm.ErrRecordNotFound)
			},
			wantErr:    true,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockUserVocabularyRepository(t)
			tt.setupMock(repo)
			svc := services.NewUserVocabularyService(repo, mocks.NewMockVocabularyItemRepository(t))

			got, err := svc.Patch(context.Background(), "user_1", itemID, tt.req)
			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			repo.AssertExpectations(t)
		})
	}
}
