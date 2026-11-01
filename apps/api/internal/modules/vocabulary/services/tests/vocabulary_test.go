package tests

import (
	"context"
	"testing"

	"engflex-api/internal/common"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakePremiumChecker defaults to "not premium" — most tests exercise
// free-content behavior and don't care about the subscription lookup.
type fakePremiumChecker struct {
	isPremium bool
}

func (f fakePremiumChecker) IsPremium(ctx context.Context, userID string) (bool, error) {
	return f.isPremium, nil
}

func setupDeck(t *testing.T) (services.VocabularyDeckService, *mocks.MockVocabularyDeckRepository) {
	t.Helper()
	return setupDeckWithPremium(t, false)
}

func setupDeckWithPremium(t *testing.T, isPremium bool) (services.VocabularyDeckService, *mocks.MockVocabularyDeckRepository) {
	t.Helper()
	mockRepo := mocks.NewMockVocabularyDeckRepository(t)
	svc := services.NewVocabularyDeckService(mockRepo, fakePremiumChecker{isPremium: isPremium})
	return svc, mockRepo
}

func requireAppError(t *testing.T, err error, expectedStatus int) *common.AppError {
	t.Helper()
	require.Error(t, err)
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, expectedStatus, appErr.Status)
	return appErr
}
