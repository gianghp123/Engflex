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

// UserVocabularyService owns per-user word state.
type UserVocabularyService interface {
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.UserVocabulary, error)
	Save(ctx context.Context, userID string, req requests.SaveUserVocabulary) (*models.UserVocabulary, error)
	Unsave(ctx context.Context, userID, itemID string) error
}

type userVocabularyService struct {
	repo  repositories.UserVocabularyRepository
	items repositories.VocabularyItemRepository
}

// NewUserVocabularyService builds the service over both repository
// interfaces: items is needed to gate Save on the referenced item's
// publish status (SRS must never pick up a draft entry).
func NewUserVocabularyService(repo repositories.UserVocabularyRepository, items repositories.VocabularyItemRepository) UserVocabularyService {
	return &userVocabularyService{repo: repo, items: items}
}

func (s *userVocabularyService) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.UserVocabulary, error) {
	items, err := s.repo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "user vocabulary")
		logger.Report(ctx, "list user vocabulary failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "user vocabulary listed", "count", len(items))
	return items, nil
}

// Save stores per-user state idempotently via Upsert. SRS must only ever
// pick up published words: a draft entry is rejected here even though the
// item id itself is valid, so a half-reviewed dictionary entry can never
// reach a learner's deck through the save path.
func (s *userVocabularyService) Save(ctx context.Context, userID string, req requests.SaveUserVocabulary) (*models.UserVocabulary, error) {
	item, err := s.items.GetByID(ctx, req.ItemID)
	if err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "get vocabulary item for save failed", appErr, "itemID", req.ItemID)
		return nil, appErr
	}
	if item.Status != enums.PublishStatusPublished {
		return nil, common.UnprocessableEntity("word is not published yet")
	}

	u := &models.UserVocabulary{UserID: userID}
	if err := utils.Map(u, req); err != nil {
		logger.Report(ctx, "map save user vocabulary failed", err)
		return nil, common.Internal()
	}
	u.SourceType = enums.VocabularySource(req.SourceType)
	if err := s.repo.Upsert(ctx, u); err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "save user vocabulary failed", appErr)
		return nil, appErr
	}
	got, err := s.repo.Get(ctx, userID, req.ItemID)
	if err != nil {
		appErr := common.FromDBError(err, "user vocabulary")
		logger.Report(ctx, "read saved user vocabulary failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "user vocabulary saved", "itemID", req.ItemID)
	return got, nil
}

func (s *userVocabularyService) Unsave(ctx context.Context, userID, itemID string) error {
	if err := s.repo.Delete(ctx, userID, itemID); err != nil {
		appErr := common.FromDBError(err, "user vocabulary")
		logger.Report(ctx, "unsave user vocabulary failed", appErr)
		return appErr
	}
	slog.InfoContext(ctx, "user vocabulary unsaved", "itemID", itemID)
	return nil
}
