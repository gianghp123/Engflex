package tests

import (
	"testing"

	"engflex-api/internal/common"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDeck(t *testing.T) (services.VocabularyDeckService, *mocks.MockVocabularyDeckRepository) {
	t.Helper()
	mockRepo := mocks.NewMockVocabularyDeckRepository(t)
	svc := services.NewVocabularyDeckService(mockRepo, mocks.NewMockVocabularyDeckItemRepository(t))
	return svc, mockRepo
}

func setupDeckItem(t *testing.T) (services.VocabularyDeckItemService, *mocks.MockVocabularyDeckItemRepository, *mocks.MockVocabularyDeckRepository, *mocks.MockVocabularyItemRepository) {
	t.Helper()
	itemRepo := mocks.NewMockVocabularyDeckItemRepository(t)
	deckRepo := mocks.NewMockVocabularyDeckRepository(t)
	entryRepo := mocks.NewMockVocabularyItemRepository(t)
	svc := services.NewVocabularyDeckItemService(itemRepo, deckRepo, entryRepo)
	return svc, itemRepo, deckRepo, entryRepo
}

func requireAppError(t *testing.T, err error, expectedStatus int) *common.AppError {
	t.Helper()
	require.Error(t, err)
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, expectedStatus, appErr.Status)
	return appErr
}
