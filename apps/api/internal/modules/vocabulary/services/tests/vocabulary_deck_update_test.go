package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

func TestVocabularyDeckService_Update_PreservesCategoryWhenOmitted(t *testing.T) {
	owner := "user_1"
	catID := "22222222-2222-2222-2222-222222222222"
	repo := mocks.NewMockVocabularyDeckRepository(t)
	repo.On("GetByID", mock.Anything, deckID).
		Return(&models.VocabularyDeck{ID: deckID, UserID: owner, CategoryID: &catID, Name: "Core"}, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)
	svc := services.NewVocabularyDeckService(repo)

	got, err := svc.Update(context.Background(), owner, deckID, false, requests.UpdateVocabularyDeck{})

	require.NoError(t, err)
	require.NotNil(t, got.CategoryID)
	assert.Equal(t, catID, *got.CategoryID)
}
