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
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

// TestVocabularyItemService_AddWord_NormalizesTerm pins that the add path stores
// the tidied term and looks the row up by normalized key, so "  Give   Up  "
// finds and creates the same word as "give up".
func TestVocabularyItemService_AddWord_NormalizesTerm(t *testing.T) {
	repo := mocks.NewMockVocabularyItemRepository(t)
	userVocab := mocks.NewMockUserVocabularyRepository(t)
	var saved *models.VocabularyItem

	repo.On("GetPrivateByTermKey", mock.Anything, "user_1", "give up").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		saved = a.Get(1).(*models.VocabularyItem)
	}).Return(nil)
	userVocab.On("Get", mock.Anything, "user_1", mock.Anything).Return(nil, gorm.ErrRecordNotFound)
	userVocab.On("CountByUser", mock.Anything, "user_1").Return(int64(0), nil)
	userVocab.On("Upsert", mock.Anything, mock.Anything, false).Return(nil)
	repo.On("GetByID", mock.Anything, "user_1", mock.Anything).Return(&models.VocabularyItem{Term: "give up"}, nil)

	// Four mocks: Task 2 added the two deck repositories to the constructor.
	svc := services.NewVocabularyItemService(repo, userVocab, mocks.NewMockVocabularyDeckRepository(t), mocks.NewMockVocabularyDeckItemRepository(t))

	_, err := svc.AddWord(context.Background(), "user_1", requests.AddUserVocabularyWord{
		Term: "  Give   Up  ", CEFR: "B2",
	})
	require.NoError(t, err)
	require.NotNil(t, saved)
	// TidyTerm trims and collapses, but keeps the user's casing: the stored term
	// is display text, and the case-folded identity is the key used for lookup.
	assert.Equal(t, "Give Up", saved.Term)
	repo.AssertCalled(t, "GetPrivateByTermKey", mock.Anything, "user_1", "give up")
}

// TestVocabularyItemService_RejectsBlankTerm pins that a term which is only
// whitespace is refused: TidyTerm collapses it to "", and an empty term is not
// a word. Both write paths must reject it, and neither may touch the database.
func TestVocabularyItemService_RejectsBlankTerm(t *testing.T) {
	tests := []struct {
		name string
		call func(svc services.VocabularyItemService) error
	}{
		{
			name: "AddWord",
			call: func(svc services.VocabularyItemService) error {
				_, err := svc.AddWord(context.Background(), "user_1", requests.AddUserVocabularyWord{Term: "   ", CEFR: "B2"})
				return err
			},
		},
		{
			name: "Create",
			call: func(svc services.VocabularyItemService) error {
				_, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyItem{Term: " \t\n ", CEFR: "B2"})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockVocabularyItemRepository(t)
			userVocab := mocks.NewMockUserVocabularyRepository(t)
			svc := services.NewVocabularyItemService(repo, userVocab, mocks.NewMockVocabularyDeckRepository(t), mocks.NewMockVocabularyDeckItemRepository(t))

			err := tt.call(svc)
			requireAppError(t, err, http.StatusBadRequest)
			repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			userVocab.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}
