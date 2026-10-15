package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/services"
)

// TestVocabularyDeckService_ListVisibleWithPreview pins that every visible deck
// comes back with its own preview attached — including a deck with none, which
// must still appear (the landing renders a section per deck).
func TestVocabularyDeckService_ListVisibleWithPreview(t *testing.T) {
	repo := mocks.NewMockVocabularyDeckRepository(t)
	deckItems := mocks.NewMockVocabularyDeckItemRepository(t)

	uid := "user_1"
	full := &models.VocabularyDeck{ID: deckID, UserID: &uid, Name: "Business"}
	empty := &models.VocabularyDeck{ID: "77777777-7777-7777-7777-777777777777", UserID: &uid, Name: "Empty"}

	repo.On("ListVisible", mock.Anything, "user_1", 20, 0).Return([]*models.VocabularyDeck{full, empty}, nil)
	repo.On("CountVisible", mock.Anything, "user_1").Return(int64(2), nil)
	deckItems.On("ListPreviewByDeckIDs", mock.Anything, "user_1", []string{full.ID, empty.ID}, 6).
		Return([]*models.VocabularyDeckItem{
			{ID: deckItem, DeckID: full.ID, ItemID: wordID, Item: &models.VocabularyItem{ID: wordID, Term: "bank"}},
		}, nil)

	svc := services.NewVocabularyDeckService(repo, deckItems)
	got, total, err := svc.ListVisibleWithPreview(context.Background(), "user_1", 6, 20, 0)
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, got, 2)
	require.Len(t, got[0].Items, 1)
	assert.Equal(t, "bank", got[0].Items[0].Item.Term)
	assert.Empty(t, got[1].Items) // the empty deck still comes back
}
