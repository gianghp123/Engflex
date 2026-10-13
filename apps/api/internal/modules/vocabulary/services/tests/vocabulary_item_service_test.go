package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

const itemID = "33333333-3333-3333-3333-333333333333"

func setupItem(t *testing.T) (services.VocabularyItemService, *mocks.MockVocabularyItemRepository) {
	t.Helper()
	mockRepo := mocks.NewMockVocabularyItemRepository(t)
	svc := services.NewVocabularyItemService(mockRepo)
	return svc, mockRepo
}

func TestVocabularyItemService_List_FiltersDraftForNonAdmin(t *testing.T) {
	tests := []struct {
		name    string
		isAdmin bool
	}{
		{name: "non-admin: repo called with onlyPublished=true", isAdmin: false},
		{name: "admin: repo called with onlyPublished=false", isAdmin: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupItem(t)
			mockRepo.On("List", mock.Anything, 20, 0, !tt.isAdmin).Return([]*models.VocabularyItem{}, nil)
			mockRepo.On("Count", mock.Anything, !tt.isAdmin).Return(int64(0), nil)

			_, _, err := svc.List(context.Background(), 20, 0, tt.isAdmin)
			require.NoError(t, err)
		})
	}
}

func TestVocabularyItemService_GetByID(t *testing.T) {
	draft := &models.VocabularyItem{ID: itemID, Term: "ephemeral", Status: enums.PublishStatusDraft}
	published := &models.VocabularyItem{ID: itemID, Term: "robust", Status: enums.PublishStatusPublished}

	tests := []struct {
		name       string
		item       *models.VocabularyItem
		isAdmin    bool
		wantErr    bool
		wantStatus int
	}{
		{name: "anyone sees a published entry", item: published},
		{name: "non-admin blocked from a draft entry: 404", item: draft, wantErr: true, wantStatus: http.StatusNotFound},
		{name: "admin sees a draft entry", item: draft, isAdmin: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupItem(t)
			mockRepo.On("GetByID", mock.Anything, itemID).Return(tt.item, nil)

			res, err := svc.GetByID(context.Background(), itemID, tt.isAdmin)

			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
			}
		})
	}
}

func TestVocabularyItemService_Create_CustomWordAlwaysPublished(t *testing.T) {
	svc, mockRepo := setupItem(t)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(i *models.VocabularyItem) bool {
		return i.CreatedByUserID != nil && *i.CreatedByUserID == "user_1" && i.Status == enums.PublishStatusPublished
	})).Return(nil)

	res, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyItem{Term: "serendipity", CEFR: "C1"})

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, enums.PublishStatusPublished, res.Status)
}

func TestVocabularyItemService_CreateSystemEntry_StartsDraft(t *testing.T) {
	svc, mockRepo := setupItem(t)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(i *models.VocabularyItem) bool {
		return i.CreatedByUserID == nil && i.Status == enums.PublishStatusDraft
	})).Return(nil)

	res, err := svc.CreateSystemEntry(context.Background(), requests.CreateVocabularyItem{Term: "ubiquitous", CEFR: "C2"})

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Nil(t, res.CreatedByUserID)
	assert.Equal(t, enums.PublishStatusDraft, res.Status)
}

func TestVocabularyItemService_Update(t *testing.T) {
	customUser := "user_1"
	systemEntry := &models.VocabularyItem{ID: itemID, Term: "old", Status: enums.PublishStatusDraft}
	customWord := &models.VocabularyItem{ID: itemID, Term: "old", CreatedByUserID: &customUser}

	tests := []struct {
		name       string
		item       *models.VocabularyItem
		wantErr    bool
		wantStatus int
	}{
		{name: "admin updates a system entry", item: systemEntry},
		{name: "blocked from a user's custom word even via admin route: 403", item: customWord, wantErr: true, wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupItem(t)
			mockRepo.On("GetByID", mock.Anything, itemID).Return(tt.item, nil)
			if !tt.wantErr {
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			}

			res, err := svc.Update(context.Background(), itemID, requests.UpdateVocabularyItem{Term: strPtr("new")})

			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, "new", res.Term)
			}
		})
	}
}

func TestVocabularyItemService_Delete(t *testing.T) {
	customUser := "user_1"
	systemEntry := &models.VocabularyItem{ID: itemID}
	customWord := &models.VocabularyItem{ID: itemID, CreatedByUserID: &customUser}

	tests := []struct {
		name       string
		item       *models.VocabularyItem
		wantErr    bool
		wantStatus int
	}{
		{name: "admin deletes a system entry", item: systemEntry},
		{name: "blocked from a user's custom word: 403", item: customWord, wantErr: true, wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupItem(t)
			mockRepo.On("GetByID", mock.Anything, itemID).Return(tt.item, nil)
			if !tt.wantErr {
				mockRepo.On("Delete", mock.Anything, itemID).Return(nil)
			}

			err := svc.Delete(context.Background(), itemID)

			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
