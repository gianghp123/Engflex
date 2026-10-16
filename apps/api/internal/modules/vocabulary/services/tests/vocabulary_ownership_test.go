package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

func TestVocabularyDeckItemService_ListByDeck_ForeignDeck(t *testing.T) {
	owner := "user_1"
	other := "user_9"
	deckRepo := mocks.NewMockVocabularyDeckRepository(t)
	deckRepo.On("GetByID", mock.Anything, deckID).
		Return(&models.VocabularyDeck{ID: deckID, UserID: owner, Name: "Core"}, nil)
	itemRepo := mocks.NewMockVocabularyDeckItemRepository(t)
	svc := services.NewVocabularyDeckItemService(itemRepo, deckRepo)

	res, err := svc.ListByDeck(context.Background(), other, deckID, false)

	requireAppError(t, err, http.StatusForbidden)
	assert.Nil(t, res)
	itemRepo.AssertNotCalled(t, "ListByDeck", mock.Anything, mock.Anything)
}

// A draft system deck's items stay invisible to a non-admin learner (404,
// not 403 — same "existence not disclosed" rule as the deck itself).
func TestVocabularyDeckItemService_ListByDeck_DraftSystemDeckHiddenFromLearner(t *testing.T) {
	admin := "admin_1"
	learner := "user_9"
	deckRepo := mocks.NewMockVocabularyDeckRepository(t)
	deckRepo.On("GetByID", mock.Anything, deckID).
		Return(&models.VocabularyDeck{ID: deckID, UserID: admin, IsDefault: true, Status: enums.PublishStatusDraft}, nil)
	itemRepo := mocks.NewMockVocabularyDeckItemRepository(t)
	svc := services.NewVocabularyDeckItemService(itemRepo, deckRepo)

	res, err := svc.ListByDeck(context.Background(), learner, deckID, false)

	requireAppError(t, err, http.StatusNotFound)
	assert.Nil(t, res)
	itemRepo.AssertNotCalled(t, "ListByDeck", mock.Anything, mock.Anything)
}

// Once published, any authenticated user can read a system deck's items —
// that's the point of publishing it.
func TestVocabularyDeckItemService_ListByDeck_PublishedSystemDeckReadableByAnyone(t *testing.T) {
	admin := "admin_1"
	learner := "user_9"
	deckRepo := mocks.NewMockVocabularyDeckRepository(t)
	deckRepo.On("GetByID", mock.Anything, deckID).
		Return(&models.VocabularyDeck{ID: deckID, UserID: admin, IsDefault: true, Status: enums.PublishStatusPublished}, nil)
	itemRepo := mocks.NewMockVocabularyDeckItemRepository(t)
	itemRepo.On("ListByDeck", mock.Anything, deckID).Return(make([]*models.VocabularyDeckItem, 0), nil)
	svc := services.NewVocabularyDeckItemService(itemRepo, deckRepo)

	res, err := svc.ListByDeck(context.Background(), learner, deckID, false)

	require.NoError(t, err)
	assert.NotNil(t, res)
}

// Any admin — not just the one who created the system deck — can manage its
// items, mirroring canManageDeck's IsDefault split for the deck itself.
func TestVocabularyDeckItemService_Create_AnyAdminManagesSystemDeck(t *testing.T) {
	creator := "admin_1"
	otherAdmin := "admin_2"
	deckRepo := mocks.NewMockVocabularyDeckRepository(t)
	deckRepo.On("GetByID", mock.Anything, deckID).
		Return(&models.VocabularyDeck{ID: deckID, UserID: creator, IsDefault: true, Status: enums.PublishStatusDraft}, nil)
	itemRepo := mocks.NewMockVocabularyDeckItemRepository(t)
	itemRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	svc := services.NewVocabularyDeckItemService(itemRepo, deckRepo)

	res, err := svc.Create(context.Background(), otherAdmin, true, requests.CreateVocabularyDeckItem{DeckID: deckID, Phrase: "hi", Meaning: "chao"})

	require.NoError(t, err)
	assert.NotNil(t, res)
}
