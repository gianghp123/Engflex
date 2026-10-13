package services

import (
	"context"
	"log/slog"
	"strings"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/utils"
)

// VocabularyDeckItemService owns flashcard rows. Every method takes the
// caller's userID and proves deck ownership through the deck repository
// before touching rows: cards inherit their deck's owner, and ids are
// unguessable but not unshareable, so the check cannot live in the
// controller alone.
type VocabularyDeckItemService interface {
	ListByDeck(ctx context.Context, userID, deckID string) ([]*models.VocabularyDeckItem, error)
	GetByID(ctx context.Context, userID, id string) (*models.VocabularyDeckItem, error)
	Create(ctx context.Context, userID string, req requests.CreateVocabularyDeckItem) (*models.VocabularyDeckItem, error)
	Update(ctx context.Context, userID, id string, req requests.UpdateVocabularyDeckItem) (*models.VocabularyDeckItem, error)
	Delete(ctx context.Context, userID, id string) error
}

type vocabularyDeckItemService struct {
	items repositories.VocabularyDeckItemRepository
	decks repositories.VocabularyDeckRepository
}

// NewVocabularyDeckItemService builds the service over both repositories.
func NewVocabularyDeckItemService(items repositories.VocabularyDeckItemRepository, decks repositories.VocabularyDeckRepository) VocabularyDeckItemService {
	return &vocabularyDeckItemService{items: items, decks: decks}
}

func (s *vocabularyDeckItemService) requireOwner(ctx context.Context, userID, deckID string) error {
	d, err := s.decks.GetByID(ctx, deckID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "deck lookup for ownership failed", appErr, "deckID", deckID)
		return appErr
	}
	if d.UserID != userID {
		return common.Forbidden("not your deck")
	}
	return nil
}

func (s *vocabularyDeckItemService) ListByDeck(ctx context.Context, userID, deckID string) ([]*models.VocabularyDeckItem, error) {
	if err := s.requireOwner(ctx, userID, deckID); err != nil {
		return nil, err
	}
	items, err := s.items.ListByDeck(ctx, deckID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "list vocabulary deck items failed", appErr, "deckID", deckID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary deck items listed", "count", len(items), "deckID", deckID)
	return items, nil
}

func (s *vocabularyDeckItemService) GetByID(ctx context.Context, userID, id string) (*models.VocabularyDeckItem, error) {
	m, err := s.items.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "get vocabulary deck item failed", appErr, "id", id)
		return nil, appErr
	}
	if err := s.requireOwner(ctx, userID, m.DeckID); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "vocabulary deck item found", "id", m.ID)
	return m, nil
}

func (s *vocabularyDeckItemService) Create(ctx context.Context, userID string, req requests.CreateVocabularyDeckItem) (*models.VocabularyDeckItem, error) {
	if err := s.requireOwner(ctx, userID, req.DeckID); err != nil {
		return nil, err
	}
	i := &models.VocabularyDeckItem{}
	if err := utils.Map(i, req); err != nil {
		logger.Report(ctx, "map create vocabulary deck item failed", err)
		return nil, common.Internal()
	}
	i.NormalizedPhrase = strings.ToLower(strings.TrimSpace(i.Phrase))
	if err := s.items.Create(ctx, i); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "create vocabulary deck item failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary deck item created", "id", i.ID)
	return i, nil
}

func (s *vocabularyDeckItemService) Update(ctx context.Context, userID, id string, req requests.UpdateVocabularyDeckItem) (*models.VocabularyDeckItem, error) {
	i, err := s.items.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "get vocabulary deck item for update failed", appErr, "id", id)
		return nil, appErr
	}
	if err := s.requireOwner(ctx, userID, i.DeckID); err != nil {
		return nil, err
	}
	if err := utils.Map(i, req); err != nil {
		logger.Report(ctx, "map update vocabulary deck item failed", err, "id", id)
		return nil, common.Internal()
	}
	if req.Phrase != nil {
		i.NormalizedPhrase = strings.ToLower(strings.TrimSpace(i.Phrase))
	}
	if err := s.items.Update(ctx, i); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "update vocabulary deck item failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary deck item updated", "id", id)
	return i, nil
}

func (s *vocabularyDeckItemService) Delete(ctx context.Context, userID, id string) error {
	i, err := s.items.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "get vocabulary deck item for delete failed", appErr, "id", id)
		return appErr
	}
	if err := s.requireOwner(ctx, userID, i.DeckID); err != nil {
		return err
	}
	if err := s.items.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "delete vocabulary deck item failed", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "vocabulary deck item deleted", "id", id)
	return nil
}
