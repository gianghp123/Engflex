package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
)

func TestVocabularyDeckService_GetDetail(t *testing.T) {
	deckID := "11111111-1111-1111-1111-111111111111"
	owner := "user_1"
	owned := &models.VocabularyDeck{
		ID: deckID, Name: "Core", UserID: &owner,
		Category: &models.VocabularyCategory{ID: "22222222-2222-2222-2222-222222222222", Name: "General"},
	}
	curated := &models.VocabularyDeck{
		ID: deckID, Name: "TOEIC 600", UserID: nil,
		Category: &models.VocabularyCategory{ID: "22222222-2222-2222-2222-222222222222", Name: "General"},
	}

	tests := []struct {
		name      string
		user      string
		id        string
		setupMock func(m *mocks.MockVocabularyDeckRepository)
		wantErr   bool
		status    int
	}{
		{
			name: "success: owned deck with category",
			user: owner, id: deckID,
			setupMock: func(m *mocks.MockVocabularyDeckRepository) {
				m.On("GetDetail", mock.Anything, owner, deckID).Return(owned, nil)
			},
		},
		{
			name: "success: curated deck is readable by anyone",
			user: "user_9", id: deckID,
			setupMock: func(m *mocks.MockVocabularyDeckRepository) {
				m.On("GetDetail", mock.Anything, "user_9", deckID).Return(curated, nil)
			},
		},
		{
			name: "not found: missing id returns 404",
			user: owner, id: "missing",
			setupMock: func(m *mocks.MockVocabularyDeckRepository) {
				m.On("GetDetail", mock.Anything, owner, "missing").Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: true, status: http.StatusNotFound,
		},
		{
			name: "forbidden: foreign private deck returns 403",
			user: "user_9", id: deckID,
			setupMock: func(m *mocks.MockVocabularyDeckRepository) {
				m.On("GetDetail", mock.Anything, "user_9", deckID).Return(owned, nil)
			},
			wantErr: true, status: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, mockRepo := setupDeck(t)
			tt.setupMock(mockRepo)

			res, err := svc.GetDetail(context.Background(), tt.user, tt.id)

			if tt.wantErr {
				requireAppError(t, err, tt.status)
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotNil(t, res.Category)
			}
		})
	}
}
