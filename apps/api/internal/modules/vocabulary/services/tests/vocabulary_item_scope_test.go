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
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

func TestVocabularyItemService_GetByID_ScopesToUser(t *testing.T) {
	id := "11111111-1111-1111-1111-111111111111"
	repo := mocks.NewMockVocabularyItemRepository(t)
	// The repo receives the caller's userID and must be asked with it.
	repo.On("GetByID", mock.Anything, "user_1", id).Return(&models.VocabularyItem{ID: id, Term: "bank"}, nil)
	repo.On("GetByID", mock.Anything, "user_2", id).Return(nil, gormErrNotFound())
	svc := services.NewVocabularyItemService(repo, mocks.NewMockUserVocabularyRepository(t), mocks.NewMockVocabularyDeckRepository(t), mocks.NewMockVocabularyDeckItemRepository(t))

	_, err := svc.GetByID(context.Background(), "user_1", id)
	require.NoError(t, err)

	_, err = svc.GetByID(context.Background(), "user_2", id)
	requireAppError(t, err, http.StatusNotFound)
	repo.AssertExpectations(t)
}

func TestVocabularyItemService_Create_StampsOwner(t *testing.T) {
	repo := mocks.NewMockVocabularyItemRepository(t)
	var saved *models.VocabularyItem
	repo.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		saved = a.Get(1).(*models.VocabularyItem)
	}).Return(nil)
	svc := services.NewVocabularyItemService(repo, mocks.NewMockUserVocabularyRepository(t), mocks.NewMockVocabularyDeckRepository(t), mocks.NewMockVocabularyDeckItemRepository(t))

	got, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyItem{
		Term: "serendipity", CEFR: "B2",
	})
	require.NoError(t, err)
	require.NotNil(t, saved.CreatedByUserID)
	require.Equal(t, "user_1", *saved.CreatedByUserID)
	require.Equal(t, "user_1", gotUUIDOwner(got)) // helper below
}

func TestVocabularyItemService_Create_MapsPartOfSpeech(t *testing.T) {
	repo := mocks.NewMockVocabularyItemRepository(t)
	var saved *models.VocabularyItem
	repo.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		saved = a.Get(1).(*models.VocabularyItem)
	}).Return(nil)
	svc := services.NewVocabularyItemService(repo, mocks.NewMockUserVocabularyRepository(t), mocks.NewMockVocabularyDeckRepository(t), mocks.NewMockVocabularyDeckItemRepository(t))

	pos := "phrasal_verb"
	_, err := svc.Create(context.Background(), "user_1", requests.CreateVocabularyItem{
		Term: "give up", CEFR: "B2", PartOfSpeech: &pos,
	})
	require.NoError(t, err)
	require.NotNil(t, saved.PartOfSpeech)
	assert.Equal(t, enums.VocabularyPartOfSpeechPhrasalVerb, *saved.PartOfSpeech)
}

func gormErrNotFound() error { return gorm.ErrRecordNotFound }
func gotUUIDOwner(m *models.VocabularyItem) string {
	if m.CreatedByUserID == nil {
		return ""
	}
	return *m.CreatedByUserID
}
