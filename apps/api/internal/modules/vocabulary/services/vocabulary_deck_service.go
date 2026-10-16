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

// VocabularyDeckService owns flashcard decks: personal (owner-managed) and
// system-curated (admin-managed, gated by Status for everyone else).
type VocabularyDeckService interface {
	ListByUser(ctx context.Context, userID string, isAdmin bool, limit, offset int) ([]*models.VocabularyDeck, int64, error)
	GetDetail(ctx context.Context, userID, id string, isAdmin bool) (*models.VocabularyDeck, error)
	Create(ctx context.Context, userID string, req requests.CreateVocabularyDeck) (*models.VocabularyDeck, error)
	CreateSystemDeck(ctx context.Context, adminUserID string, req requests.CreateVocabularyDeck) (*models.VocabularyDeck, error)
	Update(ctx context.Context, userID, id string, isAdmin bool, req requests.UpdateVocabularyDeck) (*models.VocabularyDeck, error)
	Delete(ctx context.Context, userID, id string, isAdmin bool) error
}

// canManage enforces: system decks (IsDefault) are admin-only; personal
// decks are owner-only.
func canManageDeck(d *models.VocabularyDeck, userID string, isAdmin bool) bool {
	if d.IsDefault {
		return isAdmin
	}
	return d.UserID == userID
}

type vocabularyDeckService struct {
	repo repositories.VocabularyDeckRepository
}

// NewVocabularyDeckService builds the service over the repository interface.
func NewVocabularyDeckService(repo repositories.VocabularyDeckRepository) VocabularyDeckService {
	return &vocabularyDeckService{repo: repo}
}

func (s *vocabularyDeckService) ListByUser(ctx context.Context, userID string, isAdmin bool, limit, offset int) ([]*models.VocabularyDeck, int64, error) {
	items, err := s.repo.ListByUser(ctx, userID, isAdmin, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "list vocabulary decks failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.CountByUser(ctx, userID, isAdmin)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "count vocabulary decks failed", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "vocabulary decks listed", "count", len(items), "total", total)
	return items, total, nil
}

// GetDetail: a draft system deck is invisible to non-admins (404, not 403 —
// its existence itself isn't disclosed), matching the "what's allowed to
// show to learners" scope. Personal decks keep the prior owner-only check.
func (s *vocabularyDeckService) GetDetail(ctx context.Context, userID, id string, isAdmin bool) (*models.VocabularyDeck, error) {
	m, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "get vocabulary deck detail failed", appErr, "id", id)
		return nil, appErr
	}
	if m.IsDefault {
		if m.Status != enums.PublishStatusPublished && !isAdmin {
			return nil, common.NotFound("vocabulary deck not found")
		}
	} else if m.UserID != userID {
		return nil, common.Forbidden("not your deck")
	}
	slog.InfoContext(ctx, "vocabulary deck found", "id", m.ID)
	return m, nil
}

// Create makes a personal deck for the caller. Always published: there is
// no review step for a learner's own content.
func (s *vocabularyDeckService) Create(ctx context.Context, userID string, req requests.CreateVocabularyDeck) (*models.VocabularyDeck, error) {
	d := &models.VocabularyDeck{}
	if err := utils.Map(d, req); err != nil {
		logger.Report(ctx, "map create vocabulary deck failed", err)
		return nil, common.Internal()
	}
	d.UserID = userID
	d.Status = enums.PublishStatusPublished
	if err := s.repo.Create(ctx, d); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "create vocabulary deck failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary deck created", "id", d.ID)
	return d, nil
}

// CreateSystemDeck makes an admin-curated deck, starting as draft until
// explicitly published.
func (s *vocabularyDeckService) CreateSystemDeck(ctx context.Context, adminUserID string, req requests.CreateVocabularyDeck) (*models.VocabularyDeck, error) {
	d := &models.VocabularyDeck{}
	if err := utils.Map(d, req); err != nil {
		logger.Report(ctx, "map create system vocabulary deck failed", err)
		return nil, common.Internal()
	}
	d.UserID = adminUserID
	d.IsDefault = true
	d.Status = enums.PublishStatusDraft
	if err := s.repo.Create(ctx, d); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "create system vocabulary deck failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "system vocabulary deck created", "id", d.ID)
	return d, nil
}

func (s *vocabularyDeckService) Update(ctx context.Context, userID, id string, isAdmin bool, req requests.UpdateVocabularyDeck) (*models.VocabularyDeck, error) {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "get vocabulary deck for update failed", appErr, "id", id)
		return nil, appErr
	}
	if !canManageDeck(d, userID, isAdmin) {
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
		d.Name = *req.Name
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
	// Status only means anything for a system deck, and only an admin can
	// publish/unpublish one — canManageDeck above already guarantees that.
	if req.Status != nil && d.IsDefault {
		d.Status = *req.Status
	}
	if err := s.repo.Update(ctx, d); err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "update vocabulary deck failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "vocabulary deck updated", "id", id)
	return d, nil
}

func (s *vocabularyDeckService) Delete(ctx context.Context, userID, id string, isAdmin bool) error {
	d, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary deck")
		logger.Report(ctx, "get vocabulary deck for delete failed", appErr, "id", id)
		return appErr
	}
	if !canManageDeck(d, userID, isAdmin) {
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
