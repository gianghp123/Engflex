package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
)

const deckID = "11111111-1111-1111-1111-111111111111"

func TestVocabularyDeckService_GetDetail(t *testing.T) {
	personal := &models.VocabularyDeck{ID: deckID, Name: "Core", UserID: "user_1"}
	draftSystem := &models.VocabularyDeck{ID: deckID, Name: "Curated", UserID: "admin_1", IsDefault: true, Status: enums.PublishStatusDraft}
	publishedSystem := &models.VocabularyDeck{ID: deckID, Name: "Curated", UserID: "admin_1", IsDefault: true, Status: enums.PublishStatusPublished}

	tests := []struct {
		name       string
		deck       *models.VocabularyDeck
		callerID   string
		isAdmin    bool
		wantErr    bool
		wantStatus int
	}{
		{name: "owner sees their personal deck", deck: personal, callerID: "user_1"},
		{name: "another user blocked from a personal deck: 403", deck: personal, callerID: "user_9", wantErr: true, wantStatus: http.StatusForbidden},
		{name: "anyone sees a published system deck", deck: publishedSystem, callerID: "user_9"},
		{name: "non-admin blocked from a draft system deck: 404 (existence hidden)", deck: draftSystem, callerID: "user_9", wantErr: true, wantStatus: http.StatusNotFound},
		{name: "admin sees a draft system deck", deck: draftSystem, callerID: "admin_1", isAdmin: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupDeck(t)
			mockRepo.On("GetDetail", mock.Anything, deckID).Return(tt.deck, nil)

			res, err := svc.GetDetail(context.Background(), tt.callerID, deckID, tt.isAdmin)

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

func TestVocabularyDeckService_GetDetail_NotFound(t *testing.T) {
	svc, mockRepo := setupDeck(t)
	mockRepo.On("GetDetail", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound)

	res, err := svc.GetDetail(context.Background(), "user_1", "missing", false)

	requireAppError(t, err, http.StatusNotFound)
	assert.Nil(t, res)
}

func TestVocabularyDeckService_Create(t *testing.T) {
	svc, mockRepo := setupDeck(t)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(d *models.VocabularyDeck) bool {
		return d.UserID == "user_1" && !d.IsDefault && d.Status == enums.PublishStatusPublished
	})).Return(nil)

	res, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyDeck{Name: "My deck"})

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "user_1", res.UserID)
	assert.False(t, res.IsDefault)
	assert.Equal(t, enums.PublishStatusPublished, res.Status)
}

func TestVocabularyDeckService_CreateSystemDeck(t *testing.T) {
	svc, mockRepo := setupDeck(t)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(d *models.VocabularyDeck) bool {
		return d.UserID == "admin_1" && d.IsDefault && d.Status == enums.PublishStatusDraft
	})).Return(nil)

	res, err := svc.CreateSystemDeck(context.Background(), "admin_1", requests.CreateVocabularyDeck{Name: "Business English"})

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.IsDefault)
	assert.Equal(t, enums.PublishStatusDraft, res.Status)
}

func TestVocabularyDeckService_Update(t *testing.T) {
	personal := &models.VocabularyDeck{ID: deckID, Name: "Core", UserID: "user_1"}
	systemDeck := &models.VocabularyDeck{ID: deckID, Name: "Curated", UserID: "admin_1", IsDefault: true, Status: enums.PublishStatusDraft}

	tests := []struct {
		name       string
		deck       *models.VocabularyDeck
		callerID   string
		isAdmin    bool
		req        requests.UpdateVocabularyDeck
		wantErr    bool
		wantStatus int
	}{
		{
			name: "owner updates their personal deck",
			deck: personal, callerID: "user_1",
			req: requests.UpdateVocabularyDeck{Name: strPtr("Renamed")},
		},
		{
			name: "non-owner blocked from a personal deck: 403",
			deck: personal, callerID: "user_9",
			req:     requests.UpdateVocabularyDeck{Name: strPtr("Renamed")},
			wantErr: true, wantStatus: http.StatusForbidden,
		},
		{
			name: "non-admin blocked from a system deck even if it has an id: 403",
			deck: systemDeck, callerID: "user_9",
			req:     requests.UpdateVocabularyDeck{Name: strPtr("Renamed")},
			wantErr: true, wantStatus: http.StatusForbidden,
		},
		{
			name: "admin publishes a system deck",
			deck: systemDeck, callerID: "admin_1", isAdmin: true,
			req: requests.UpdateVocabularyDeck{Status: publishPtr(enums.PublishStatusPublished)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupDeck(t)
			mockRepo.On("GetByID", mock.Anything, deckID).Return(tt.deck, nil)
			if !tt.wantErr {
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			}

			res, err := svc.Update(context.Background(), tt.callerID, deckID, tt.isAdmin, tt.req)

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

func TestVocabularyDeckService_Delete(t *testing.T) {
	personal := &models.VocabularyDeck{ID: deckID, Name: "Core", UserID: "user_1"}
	systemDeck := &models.VocabularyDeck{ID: deckID, Name: "Curated", UserID: "admin_1", IsDefault: true}

	tests := []struct {
		name       string
		deck       *models.VocabularyDeck
		callerID   string
		isAdmin    bool
		wantErr    bool
		wantStatus int
	}{
		{name: "owner deletes their personal deck", deck: personal, callerID: "user_1"},
		{name: "non-owner blocked: 403", deck: personal, callerID: "user_9", wantErr: true, wantStatus: http.StatusForbidden},
		{name: "non-admin blocked from a system deck: 403", deck: systemDeck, callerID: "user_1", wantErr: true, wantStatus: http.StatusForbidden},
		{name: "admin deletes a system deck", deck: systemDeck, callerID: "admin_1", isAdmin: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupDeck(t)
			mockRepo.On("GetByID", mock.Anything, deckID).Return(tt.deck, nil)
			if !tt.wantErr {
				mockRepo.On("Delete", mock.Anything, deckID).Return(nil)
			}

			err := svc.Delete(context.Background(), tt.callerID, deckID, tt.isAdmin)

			if tt.wantErr {
				requireAppError(t, err, tt.wantStatus)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func strPtr(s string) *string                               { return &s }
func publishPtr(s enums.PublishStatus) *enums.PublishStatus { return &s }
