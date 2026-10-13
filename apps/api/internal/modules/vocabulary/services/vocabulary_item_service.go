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

// VocabularyItemService owns the shared dictionary: admin-curated entries
// (CreatedByUserID nil, gated by Status) and user-submitted custom words
// (CreatedByUserID set, always published — no review step for personal
// words). Update/Delete only ever touch admin-curated entries; a user's
// custom word is never reachable through them.
type VocabularyItemService interface {
	List(ctx context.Context, limit, offset int, isAdmin bool) ([]*models.VocabularyItem, int64, error)
	GetByID(ctx context.Context, id string, isAdmin bool) (*models.VocabularyItem, error)
	Create(ctx context.Context, userID string, req requests.CreateVocabularyItem) (*models.VocabularyItem, error)
	CreateSystemEntry(ctx context.Context, req requests.CreateVocabularyItem) (*models.VocabularyItem, error)
	Update(ctx context.Context, id string, req requests.UpdateVocabularyItem) (*models.VocabularyItem, error)
	Delete(ctx context.Context, id string) error
}

type vocabularyItemService struct {
	repo repositories.VocabularyItemRepository
}

// NewVocabularyItemService builds the service over the repository interface.
func NewVocabularyItemService(repo repositories.VocabularyItemRepository) VocabularyItemService {
	return &vocabularyItemService{repo: repo}
}

func (s *vocabularyItemService) List(ctx context.Context, limit, offset int, isAdmin bool) ([]*models.VocabularyItem, int64, error) {
	items, err := s.repo.List(ctx, limit, offset, !isAdmin)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "list vocabulary items failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.Count(ctx, !isAdmin)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "count vocabulary items failed", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "vocabulary items listed", "count", len(items), "total", total)
	return items, total, nil
}

// GetByID: a draft entry is invisible to non-admins (404, not just hidden
// from lists) — its existence isn't disclosed either, same rule as a draft
// system deck.
func (s *vocabularyItemService) GetByID(ctx context.Context, id string, isAdmin bool) (*models.VocabularyItem, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "get vocabulary item failed", appErr, "id", id)
		return nil, appErr
	}
	if m.Status != enums.PublishStatusPublished && !isAdmin {
		return nil, common.NotFound("vocabulary item not found")
	}
	slog.InfoContext(ctx, "vocabulary item found", "id", m.ID)
	return m, nil
}

// Create adds the caller's own custom word. Always published: no review
// step for personal words.
func (s *vocabularyItemService) Create(ctx context.Context, userID string, req requests.CreateVocabularyItem) (*models.VocabularyItem, error) {
	i := &models.VocabularyItem{}
	if err := utils.Map(i, req); err != nil {
		logger.Report(ctx, "map create vocabulary item failed", err)
		return nil, common.Internal()
	}
	i.CEFR = enums.CEFR(req.CEFR)
	if req.Domain != nil {
		d := enums.VocabularyDomain(*req.Domain)
		i.Domain = &d
	}
	if userID != "" {
		uid := userID
		i.CreatedByUserID = &uid
	}
	i.Status = enums.PublishStatusPublished
	if err := s.repo.Create(ctx, i); err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "create vocabulary item failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary item created", "id", i.ID)
	return i, nil
}

// CreateSystemEntry adds a shared dictionary entry (admin-only), starting
// as draft until explicitly published.
func (s *vocabularyItemService) CreateSystemEntry(ctx context.Context, req requests.CreateVocabularyItem) (*models.VocabularyItem, error) {
	i := &models.VocabularyItem{}
	if err := utils.Map(i, req); err != nil {
		logger.Report(ctx, "map create system vocabulary item failed", err)
		return nil, common.Internal()
	}
	i.CEFR = enums.CEFR(req.CEFR)
	if req.Domain != nil {
		d := enums.VocabularyDomain(*req.Domain)
		i.Domain = &d
	}
	i.Status = enums.PublishStatusDraft
	if err := s.repo.Create(ctx, i); err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "create system vocabulary item failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "system vocabulary item created", "id", i.ID)
	return i, nil
}

// Update edits a shared dictionary entry. Rejects a user's custom word
// (CreatedByUserID set) even though the route is admin-gated — an admin
// account is still not the entry's owner, and ownership stays with the
// user who submitted it.
func (s *vocabularyItemService) Update(ctx context.Context, id string, req requests.UpdateVocabularyItem) (*models.VocabularyItem, error) {
	i, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "get vocabulary item for update failed", appErr, "id", id)
		return nil, appErr
	}
	if i.CreatedByUserID != nil {
		return nil, common.Forbidden("cannot modify a user-submitted word")
	}
	if req.Term != nil {
		i.Term = *req.Term
	}
	if req.Definition != nil {
		i.Definition = *req.Definition
	}
	if req.CEFR != nil {
		i.CEFR = enums.CEFR(*req.CEFR)
	}
	if req.Domain != nil {
		d := enums.VocabularyDomain(*req.Domain)
		i.Domain = &d
	}
	if req.Status != nil {
		i.Status = *req.Status
	}
	if err := s.repo.Update(ctx, i); err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "update vocabulary item failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary item updated", "id", i.ID)
	return i, nil
}

func (s *vocabularyItemService) Delete(ctx context.Context, id string) error {
	i, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "get vocabulary item for delete failed", appErr, "id", id)
		return appErr
	}
	if i.CreatedByUserID != nil {
		return common.Forbidden("cannot delete a user-submitted word")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "delete vocabulary item failed", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "vocabulary item deleted", "id", id)
	return nil
}
