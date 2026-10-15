package services

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/utils"
)

// VocabularyDeckItemService owns flashcard rows. Every method takes the
// caller's userID and proves the deck is private and owned through the deck
// repository before touching rows: cards inherit their deck's owner, curated
// decks are read-only, and ids are unguessable but not unshareable, so the
// check cannot live in the controller alone.
type VocabularyDeckItemService interface {
	List(ctx context.Context, userID, deckID string) ([]*models.VocabularyDeckItem, error)
	GetByID(ctx context.Context, userID, id string) (*models.VocabularyDeckItem, error)
	Create(ctx context.Context, userID string, req requests.CreateVocabularyDeckItem) (*models.VocabularyDeckItem, error)
	Update(ctx context.Context, userID, id string, req requests.UpdateVocabularyDeckItem) (*models.VocabularyDeckItem, error)
	Delete(ctx context.Context, userID, id string) error
}

type vocabularyDeckItemService struct {
	items   repositories.VocabularyDeckItemRepository
	decks   repositories.VocabularyDeckRepository
	entries repositories.VocabularyItemRepository
}

// NewVocabularyDeckItemService builds the service over all three repositories.
func NewVocabularyDeckItemService(items repositories.VocabularyDeckItemRepository, decks repositories.VocabularyDeckRepository, entries repositories.VocabularyItemRepository) VocabularyDeckItemService {
	return &vocabularyDeckItemService{items: items, decks: decks, entries: entries}
}

func (s *vocabularyDeckItemService) List(ctx context.Context, userID, deckID string) ([]*models.VocabularyDeckItem, error) {
	if _, err := requireWritableDeck(ctx, s.decks, userID, deckID); err != nil {
		return nil, err
	}
	rows, err := s.items.ListByDeck(ctx, userID, deckID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "list vocabulary deck items failed", appErr, "deckID", deckID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary deck items listed", "deckID", deckID, "count", len(rows))
	return rows, nil
}

func (s *vocabularyDeckItemService) GetByID(ctx context.Context, userID, id string) (*models.VocabularyDeckItem, error) {
	m, err := s.items.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "get vocabulary deck item failed", appErr, "id", id)
		return nil, appErr
	}
	if _, err := requireWritableDeck(ctx, s.decks, userID, m.DeckID); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "vocabulary deck item found", "id", m.ID)
	return m, nil
}

func (s *vocabularyDeckItemService) Create(ctx context.Context, userID string, req requests.CreateVocabularyDeckItem) (*models.VocabularyDeckItem, error) {
	if _, err := requireWritableDeck(ctx, s.decks, userID, req.DeckID); err != nil {
		return nil, err
	}
	// The referenced entry must be visible to the caller (curated or their own).
	if _, err := s.entries.GetByID(ctx, userID, req.ItemID); err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "deck item entry lookup failed", appErr, "itemID", req.ItemID)
		return nil, appErr
	}
	i := &models.VocabularyDeckItem{}
	if err := utils.Map(i, req); err != nil {
		logger.Report(ctx, "map create vocabulary deck item failed", err)
		return nil, common.Internal()
	}
	saved, err := ensureDeckItem(ctx, s.items, i)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "vocabulary deck item created", "id", saved.ID)
	return saved, nil
}

func (s *vocabularyDeckItemService) Update(ctx context.Context, userID, id string, req requests.UpdateVocabularyDeckItem) (*models.VocabularyDeckItem, error) {
	i, err := s.items.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "get vocabulary deck item for update failed", appErr, "id", id)
		return nil, appErr
	}
	if _, err := requireWritableDeck(ctx, s.decks, userID, i.DeckID); err != nil {
		return nil, err
	}
	if req.ExampleSentence != nil {
		i.ExampleSentence = *req.ExampleSentence
	}
	if req.Note != nil {
		i.Note = *req.Note
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
	if _, err := requireWritableDeck(ctx, s.decks, userID, i.DeckID); err != nil {
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

// requireWritableDeck proves the deck is private and owned by the caller.
// Curated decks (user_id NULL) are read-only for everyone.
func requireWritableDeck(ctx context.Context, decks repositories.VocabularyDeckRepository, userID, deckID string) (*models.VocabularyDeck, error) {
	d, err := decks.GetByID(ctx, deckID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "deck lookup for ownership failed", appErr, "deckID", deckID)
		return nil, appErr
	}
	if d.UserID == nil {
		return nil, common.Forbidden("curated deck is read-only")
	}
	if *d.UserID != userID {
		return nil, common.Forbidden("not your deck")
	}
	return d, nil
}

// ensureDeckItem returns the deck's row for an entry, creating `row` if absent.
// Idempotent: re-adding a word to the same deck returns the existing row rather
// than losing the unique (deck_id, item_id) index and surfacing a 409. On the
// create path `row` is inserted whole, so its presentation fields survive.
func ensureDeckItem(ctx context.Context, items repositories.VocabularyDeckItemRepository, row *models.VocabularyDeckItem) (*models.VocabularyDeckItem, error) {
	existing, err := items.GetByDeckAndItem(ctx, row.DeckID, row.ItemID)
	switch {
	case err == nil:
		return existing, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		appErr := common.FromDBError(err, "vocabulary deck item")
		logger.Report(ctx, "deck item membership lookup failed", appErr, "deckID", row.DeckID, "itemID", row.ItemID)
		return nil, appErr
	}
	if cErr := items.Create(ctx, row); cErr != nil {
		// A concurrent add can still lose the unique index; the row exists, so
		// re-read and return it instead of surfacing a conflict.
		if again, rErr := items.GetByDeckAndItem(ctx, row.DeckID, row.ItemID); rErr == nil {
			return again, nil
		}
		appErr := common.FromDBError(cErr, "vocabulary deck item")
		logger.Report(ctx, "create deck item membership failed", appErr, "deckID", row.DeckID, "itemID", row.ItemID)
		return nil, appErr
	}
	return row, nil
}
