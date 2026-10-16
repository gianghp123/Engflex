package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

const (
	deckID   = "44444444-4444-4444-4444-444444444444"
	deckItem = "55555555-5555-5555-5555-555555555555"
	wordID   = "66666666-6666-6666-6666-666666666666"
)

// TestVocabularyItemService_AddWord_JoinsDeck pins the composite: adding a word
// with a deckId enrols it AND puts it in that deck, and a re-add is a no-op
// membership (the deck row already exists), never a 409.
func TestVocabularyItemService_AddWord_JoinsDeck(t *testing.T) {
	tests := []struct {
		name          string
		existingRow   *models.VocabularyDeckItem
		wantDeckWrite bool
	}{
		{name: "new membership is created", existingRow: nil, wantDeckWrite: true},
		{name: "existing membership is reused", existingRow: &models.VocabularyDeckItem{ID: deckItem, DeckID: deckID, ItemID: wordID}, wantDeckWrite: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockVocabularyItemRepository(t)
			userVocab := mocks.NewMockUserVocabularyRepository(t)
			decks := mocks.NewMockVocabularyDeckRepository(t)
			deckItems := mocks.NewMockVocabularyDeckItemRepository(t)

			existing := &models.VocabularyItem{ID: wordID, Term: "bank"}
			repo.On("GetPrivateByTermKey", mock.Anything, "user_1", "bank").Return(existing, nil)
			userVocab.On("Get", mock.Anything, "user_1", wordID).Return(nil, gorm.ErrRecordNotFound)
			userVocab.On("CountByUser", mock.Anything, "user_1").Return(int64(0), nil)
			userVocab.On("Upsert", mock.Anything, mock.Anything, false).Return(nil)
			repo.On("GetByID", mock.Anything, "user_1", wordID).Return(existing, nil)

			uid := "user_1"
			decks.On("GetByID", mock.Anything, deckID).Return(&models.VocabularyDeck{ID: deckID, UserID: &uid, Name: "Business"}, nil)
			if tt.existingRow != nil {
				deckItems.On("GetByDeckAndItem", mock.Anything, deckID, wordID).Return(tt.existingRow, nil)
			} else {
				deckItems.On("GetByDeckAndItem", mock.Anything, deckID, wordID).Return(nil, gorm.ErrRecordNotFound)
				deckItems.On("CountByDeck", mock.Anything, deckID).Return(int64(0), nil)
				deckItems.On("Create", mock.Anything, mock.Anything).Return(nil)
			}

			svc := services.NewVocabularyItemService(repo, userVocab, decks, deckItems)
			target := deckID
			_, err := svc.AddWord(context.Background(), "user_1", requests.AddUserVocabularyWord{
				Term: "bank", CEFR: "A1", DeckID: &target,
			})
			require.NoError(t, err)

			if tt.wantDeckWrite {
				deckItems.AssertCalled(t, "Create", mock.Anything, mock.Anything)
			} else {
				deckItems.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			}
		})
	}
}

// TestVocabularyItemService_AddWord_RejectsUnwritableDeck pins ownership: a deck
// that is not the caller's — or a curated one — is refused before ANY write, so
// a rejected request leaves no side effect (the plan's test name claims exactly
// that: "nothing is written").
func TestVocabularyItemService_AddWord_RejectsUnwritableDeck(t *testing.T) {
	other := "user_2"
	tests := []struct {
		name string
		deck *models.VocabularyDeck
	}{
		{"foreign deck", &models.VocabularyDeck{ID: deckID, UserID: &other, Name: "Theirs"}},
		{"curated deck", &models.VocabularyDeck{ID: deckID, UserID: nil, Name: "Curated"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockVocabularyItemRepository(t)
			userVocab := mocks.NewMockUserVocabularyRepository(t)
			decks := mocks.NewMockVocabularyDeckRepository(t)
			deckItems := mocks.NewMockVocabularyDeckItemRepository(t)

			decks.On("GetByID", mock.Anything, deckID).Return(tt.deck, nil)

			svc := services.NewVocabularyItemService(repo, userVocab, decks, deckItems)
			target := deckID
			_, err := svc.AddWord(context.Background(), "user_1", requests.AddUserVocabularyWord{
				Term: "bank", CEFR: "A1", DeckID: &target,
			})
			requireAppError(t, err, http.StatusForbidden)
			// The deck is validated first: nothing at all is written.
			repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			userVocab.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything)
			deckItems.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		})
	}
}
