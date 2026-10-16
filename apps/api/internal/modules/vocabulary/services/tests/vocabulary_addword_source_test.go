package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

// TestAddWordSource pins the origin prefill: a save started from content (the
// context dock) records sourceType/sourceId on the NEW enrolment; an omitted
// source stays manual. A re-add never rewrites the first save's source.
func TestAddWordSource(t *testing.T) {
	lessonID := "11111111-1111-1111-1111-111111111111"
	word := &models.VocabularyItem{ID: "i1", Term: "punctual", CEFR: enums.CEFRB2}

	tests := []struct {
		name   string
		req    requests.AddUserVocabularyWord
		assert func(t *testing.T, u *models.UserVocabulary)
	}{
		{
			name: "lesson source rides the new enrolment",
			req:  requests.AddUserVocabularyWord{Term: "punctual", CEFR: "B2", SourceType: "lesson", SourceID: &lessonID},
			assert: func(t *testing.T, u *models.UserVocabulary) {
				require.Equal(t, enums.VocabularySourceLesson, u.SourceType)
				require.NotNil(t, u.SourceID)
				require.Equal(t, lessonID, *u.SourceID)
			},
		},
		{
			name: "omitted source stays manual",
			req:  requests.AddUserVocabularyWord{Term: "punctual", CEFR: "B2"},
			assert: func(t *testing.T, u *models.UserVocabulary) {
				require.Equal(t, enums.VocabularySourceManual, u.SourceType)
				require.Nil(t, u.SourceID)
			},
		},
		{
			name: "video source rides the new enrolment",
			req:  requests.AddUserVocabularyWord{Term: "punctual", CEFR: "B2", SourceType: "video", SourceID: &lessonID},
			assert: func(t *testing.T, u *models.UserVocabulary) {
				require.Equal(t, enums.VocabularySourceVideo, u.SourceType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := mocks.NewMockVocabularyItemRepository(t)
			user := mocks.NewMockUserVocabularyRepository(t)
			decks := mocks.NewMockVocabularyDeckRepository(t)
			deckItems := mocks.NewMockVocabularyDeckItemRepository(t)

			items.On("GetPrivateByTermKey", mock.Anything, "u1", mock.Anything).Return(nil, gorm.ErrRecordNotFound).Once()
			items.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
			user.On("Get", mock.Anything, "u1", mock.Anything).Return(nil, gorm.ErrRecordNotFound).Once()
			user.On("CountByUser", mock.Anything, "u1").Return(int64(0), nil).Once()
			user.On("Upsert", mock.Anything, mock.MatchedBy(func(u *models.UserVocabulary) bool {
				tt.assert(t, u)
				return true
			}), false).Return(nil).Once()
			items.On("GetByID", mock.Anything, "u1", mock.Anything).Return(word, nil).Once()

			svc := services.NewVocabularyItemService(items, user, decks, deckItems)
			got, err := svc.AddWord(context.Background(), "u1", tt.req)
			require.NoError(t, err)
			require.Equal(t, word, got)
		})
	}
}
