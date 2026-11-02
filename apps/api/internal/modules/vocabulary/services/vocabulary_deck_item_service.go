package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/utils"
)

// VocabularyDeckItemService owns flashcard rows. Every method takes the
// caller's userID and isAdmin, and proves deck access through the deck
// repository before touching rows: cards inherit their deck's access rules
// (canReadDeck/canManageDeck, below), and ids are unguessable but not
// unshareable, so the check cannot live in the controller alone.
type VocabularyDeckItemService interface {
	ListByDeck(ctx context.Context, userID, deckID string, isAdmin bool) ([]*models.VocabularyDeckItem, error)
	GetByID(ctx context.Context, userID, id string, isAdmin bool) (*models.VocabularyDeckItem, error)
	Create(ctx context.Context, userID string, isAdmin bool, req requests.CreateVocabularyDeckItem) (*models.VocabularyDeckItem, error)
	Update(ctx context.Context, userID, id string, isAdmin bool, req requests.UpdateVocabularyDeckItem) (*models.VocabularyDeckItem, error)
	Delete(ctx context.Context, userID, id string, isAdmin bool) error
}

type vocabularyDeckItemService struct {
	items      repositories.VocabularyDeckItemRepository
	decks      repositories.VocabularyDeckRepository
	dictionary repositories.VocabularyItemRepository
	premium    common.PremiumChecker
}

// NewVocabularyDeckItemService builds the service over both repositories.
func NewVocabularyDeckItemService(items repositories.VocabularyDeckItemRepository, decks repositories.VocabularyDeckRepository, dictionary repositories.VocabularyItemRepository, premium common.PremiumChecker) VocabularyDeckItemService {
	return &vocabularyDeckItemService{items: items, decks: decks, dictionary: dictionary, premium: premium}
}

// requireRead mirrors vocabularyDeckService.GetDetail's access split: a
// draft system deck is invisible to non-admins (404, existence hidden) once
// published it's readable by anyone; a personal deck stays owner-only (403
// — ownership, not existence, is what's being withheld).
func (s *vocabularyDeckItemService) requireRead(ctx context.Context, userID, deckID string, isAdmin bool) error {
	d, err := s.decks.GetByID(ctx, deckID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "deck lookup for read access failed", appErr, "deckID", deckID)
		return appErr
	}
	if d.IsDefault {
		if d.Status != enums.PublishStatusPublished && !isAdmin {
			return common.NotFound("vocabulary deck not found")
		}
		return common.RequirePremiumAccess(ctx, s.premium, userID, isAdmin, d.IsPremium)
	}
	if d.UserID != userID {
		return common.Forbidden("not your deck")
	}
	return nil
}

func (s *vocabularyDeckItemService) requireManage(ctx context.Context, userID, deckID string, isAdmin bool) error {
	d, err := s.decks.GetByID(ctx, deckID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "deck lookup for manage access failed", appErr, "deckID", deckID)
		return appErr
	}
	if !canManageDeck(d, userID, isAdmin) {
		return common.Forbidden("not your deck")
	}
	return nil
}

func (s *vocabularyDeckItemService) ListByDeck(ctx context.Context, userID, deckID string, isAdmin bool) ([]*models.VocabularyDeckItem, error) {
	if err := s.requireRead(ctx, userID, deckID, isAdmin); err != nil {
		return nil, err
	}
	items, err := s.items.ListByDeck(ctx, deckID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "list vocabulary deck items failed", appErr, "deckID", deckID)
		return nil, appErr
	}
	for _, item := range items {
		if err := s.checkItem(ctx, userID, item.Item, isAdmin); err != nil {
			return nil, err
		}
	}
	slog.InfoContext(ctx, "vocabulary deck items listed", "count", len(items), "deckID", deckID)
	return items, nil
}

func (s *vocabularyDeckItemService) GetByID(ctx context.Context, userID, id string, isAdmin bool) (*models.VocabularyDeckItem, error) {
	m, err := s.items.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "get vocabulary deck item failed", appErr, "id", id)
		return nil, appErr
	}
	if err := s.requireRead(ctx, userID, m.DeckID, isAdmin); err != nil {
		return nil, err
	}
	if err := s.checkItem(ctx, userID, m.Item, isAdmin); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "vocabulary deck item found", "id", m.ID)
	return m, nil
}

func (s *vocabularyDeckItemService) Create(ctx context.Context, userID string, isAdmin bool, req requests.CreateVocabularyDeckItem) (*models.VocabularyDeckItem, error) {
	if err := s.requireManage(ctx, userID, req.DeckID, isAdmin); err != nil {
		return nil, err
	}
	i := &models.VocabularyDeckItem{}
	if err := utils.Map(i, req); err != nil {
		logger.Report(ctx, "map create vocabulary deck item failed", err)
		return nil, common.Internal()
	}
	if err := s.requireItem(ctx, userID, req.DeckID, req.ItemID, isAdmin); err != nil {
		return nil, err
	}
	if err := s.items.Create(ctx, i); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "create vocabulary deck item failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary deck item created", "id", i.ID)
	return s.GetByID(ctx, userID, i.ID, isAdmin)
}

func (s *vocabularyDeckItemService) Update(ctx context.Context, userID, id string, isAdmin bool, req requests.UpdateVocabularyDeckItem) (*models.VocabularyDeckItem, error) {
	i, err := s.items.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "get vocabulary deck item for update failed", appErr, "id", id)
		return nil, appErr
	}
	if err := s.requireManage(ctx, userID, i.DeckID, isAdmin); err != nil {
		return nil, err
	}
	if err := utils.Map(i, req); err != nil {
		logger.Report(ctx, "map update vocabulary deck item failed", err, "id", id)
		return nil, common.Internal()
	}
	if err := s.requireItem(ctx, userID, i.DeckID, i.ItemID, isAdmin); err != nil {
		return nil, err
	}
	if err := s.items.Update(ctx, i); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "update vocabulary deck item failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary deck item updated", "id", id)
	return s.GetByID(ctx, userID, i.ID, isAdmin)
}

func (s *vocabularyDeckItemService) Delete(ctx context.Context, userID, id string, isAdmin bool) error {
	i, err := s.items.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "get vocabulary deck item for delete failed", appErr, "id", id)
		return appErr
	}
	if err := s.requireManage(ctx, userID, i.DeckID, isAdmin); err != nil {
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

// requireItem applies dictionary visibility without granting access to another user's words.
func (s *vocabularyDeckItemService) requireItem(ctx context.Context, userID, deckID, itemID string, isAdmin bool) error {
	item, err := s.dictionary.GetByID(ctx, itemID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "load dictionary reference failed", appErr, "itemID", itemID)
		return appErr
	}
	if item.CreatedByUserID != nil {
		if *item.CreatedByUserID != userID {
			return common.NotFound("vocabulary item not found")
		}
		deck, err := s.decks.GetByID(ctx, deckID)
		if err != nil {
			appErr := common.FromDBError(err, "vocabulary deck")
			logger.Report(ctx, "load deck reference policy failed", appErr)
			return appErr
		}
		if deck.IsDefault {
			return common.UnprocessableEntity("system decks require shared dictionary entries")
		}
	} else if item.Status != enums.PublishStatusPublished && !isAdmin {
		return common.NotFound("vocabulary item not found")
	}
	return common.RequirePremiumAccess(ctx, s.premium, userID, isAdmin, item.IsPremium)
}

func (s *vocabularyDeckItemService) checkItem(ctx context.Context, userID string, item *models.VocabularyItem, isAdmin bool) error {
	if item == nil {
		appErr := common.Internal()
		logger.Report(ctx, "dictionary relation missing", appErr)
		return appErr
	}
	if item.CreatedByUserID != nil && *item.CreatedByUserID != userID {
		return common.NotFound("vocabulary item not found")
	}
	if item.CreatedByUserID == nil && item.Status != enums.PublishStatusPublished && !isAdmin {
		return common.NotFound("vocabulary item not found")
	}
	return common.RequirePremiumAccess(ctx, s.premium, userID, isAdmin, item.IsPremium)
}
