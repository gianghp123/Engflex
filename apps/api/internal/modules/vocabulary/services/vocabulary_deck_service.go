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

// VocabularyDeckService owns flashcard decks. Curated decks (user_id NULL)
// are readable by everyone but writable by no one; private decks are
// readable and writable only by their owner.
type VocabularyDeckService interface {
	ListVisible(ctx context.Context, userID string, limit, offset int) ([]*models.VocabularyDeck, int64, error)
	ListVisibleWithPreview(ctx context.Context, userID string, previewK, limit, offset int) ([]*models.VocabularyDeck, int64, error)
	GetDetail(ctx context.Context, userID, id string) (*models.VocabularyDeck, error)
	Create(ctx context.Context, userID string, req requests.CreateVocabularyDeck) (*models.VocabularyDeck, error)
	Update(ctx context.Context, userID, id string, req requests.UpdateVocabularyDeck) (*models.VocabularyDeck, error)
	Delete(ctx context.Context, userID, id string) error
}

type vocabularyDeckService struct {
	repo      repositories.VocabularyDeckRepository
	deckItems repositories.VocabularyDeckItemRepository
}

// NewVocabularyDeckService builds the service over both repositories.
func NewVocabularyDeckService(repo repositories.VocabularyDeckRepository, deckItems repositories.VocabularyDeckItemRepository) VocabularyDeckService {
	return &vocabularyDeckService{repo: repo, deckItems: deckItems}
}

func (s *vocabularyDeckService) ListVisible(ctx context.Context, userID string, limit, offset int) ([]*models.VocabularyDeck, int64, error) {
	items, err := s.repo.ListVisible(ctx, userID, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "list vocabulary decks failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.CountVisible(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "count vocabulary decks failed", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "vocabulary decks listed", "count", len(items), "total", total)
	return items, total, nil
}

// ListVisibleWithPreview returns every visible deck with up to previewK of its
// words attached, each carrying the caller's review state. previewK is clamped
// because the landing is a preview; full rows stay on the deck detail.
func (s *vocabularyDeckService) ListVisibleWithPreview(ctx context.Context, userID string, previewK, limit, offset int) ([]*models.VocabularyDeck, int64, error) {
	if previewK < 1 {
		previewK = 6
	}
	if previewK > 12 {
		previewK = 12
	}
	decks, err := s.repo.ListVisible(ctx, userID, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "list vocabulary decks failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.CountVisible(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "count vocabulary decks failed", appErr)
		return nil, 0, appErr
	}
	// Every deck keeps a non-nil Items, so a deck with no words still renders.
	for _, d := range decks {
		d.Items = make([]*models.VocabularyDeckItem, 0)
	}
	if len(decks) > 0 {
		ids := make([]string, 0, len(decks))
		byID := make(map[string]*models.VocabularyDeck, len(decks))
		for _, d := range decks {
			ids = append(ids, d.ID)
			byID[d.ID] = d
		}
		rows, rErr := s.deckItems.ListPreviewByDeckIDs(ctx, userID, ids, previewK)
		if rErr != nil {
			appErr := common.FromDBError(rErr, "vocabulary deck item")
			logger.Report(ctx, "list deck previews failed", appErr)
			return nil, 0, appErr
		}
		for _, row := range rows {
			if d := byID[row.DeckID]; d != nil {
				d.Items = append(d.Items, row)
			}
		}
	}
	slog.InfoContext(ctx, "vocabulary decks listed with preview", "count", len(decks), "total", total, "previewK", previewK)
	return decks, total, nil
}

func (s *vocabularyDeckService) GetDetail(ctx context.Context, userID, id string) (*models.VocabularyDeck, error) {
	m, err := s.repo.GetDetail(ctx, userID, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "get vocabulary deck detail failed", appErr, "id", id)
		return nil, appErr
	}
	if m.UserID != nil && *m.UserID != userID {
		return nil, common.Forbidden("not your deck")
	}
	slog.InfoContext(ctx, "vocabulary deck found", "id", m.ID)
	return m, nil
}

func (s *vocabularyDeckService) Create(ctx context.Context, userID string, req requests.CreateVocabularyDeck) (*models.VocabularyDeck, error) {
	// Names are unique per user (lower(name) index): trim before the blank
	// guard so "   " can never reach the unique index as a distinct name.
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, common.BadRequest("name is required")
	}
	req.Name = name
	// Product cap; curated decks (user_id NULL) are shared and never count.
	// Checked without a transaction: concurrent creates could land one over —
	// the cap is a product guard, not a constraint.
	owned, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "count user decks failed", appErr)
		return nil, appErr
	}
	if owned >= common.MaxUserDecks {
		return nil, common.UnprocessableEntity("deck limit reached")
	}
	d := &models.VocabularyDeck{}
	if err := utils.Map(d, req); err != nil {
		logger.Report(ctx, "map create vocabulary deck failed", err)
		return nil, common.Internal()
	}
	uid := userID
	d.UserID = &uid
	if err := s.repo.Create(ctx, d); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "create vocabulary deck failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary deck created", "id", d.ID)
	return d, nil
}

func (s *vocabularyDeckService) Update(ctx context.Context, userID, id string, req requests.UpdateVocabularyDeck) (*models.VocabularyDeck, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "get vocabulary deck for update failed", appErr, "id", id)
		return nil, appErr
	}
	if d.UserID == nil {
		return nil, common.Forbidden("curated deck is read-only")
	}
	if *d.UserID != userID {
		return nil, common.Forbidden("not your deck")
	}
	// Partial update: UpdateVocabularyDeck fields are all pointers and copier
	// copies nil *string over *string (verified live), so only non-nil fields
	// are assigned — a blanket utils.Map here would detach CategoryID whenever
	// the body omits it.
	if req.CategoryID != nil {
		d.CategoryID = req.CategoryID
	}
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, common.BadRequest("name is required")
		}
		d.Name = name
	}
	if req.Description != nil {
		d.Description = *req.Description
	}
	if req.ThumbnailURL != nil {
		d.ThumbnailURL = *req.ThumbnailURL
	}
	if req.Level != nil {
		d.Level = *req.Level
	}
	if err := s.repo.Update(ctx, d); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "update vocabulary deck failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary deck updated", "id", id)
	return d, nil
}

func (s *vocabularyDeckService) Delete(ctx context.Context, userID, id string) error {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "get vocabulary deck for delete failed", appErr, "id", id)
		return appErr
	}
	if d.UserID == nil {
		return common.Forbidden("curated deck is read-only")
	}
	if *d.UserID != userID {
		return common.Forbidden("not your deck")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "delete vocabulary deck failed", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "vocabulary deck deleted", "id", id)
	return nil
}
