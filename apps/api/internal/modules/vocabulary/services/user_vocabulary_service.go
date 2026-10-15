package services

import (
	"context"
	"log/slog"
	"time"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
)

// NextIntervalDays advances the review interval. grade: 0 = again, 1 = hard,
// 2 = good, 3 = easy. A zero interval means "due again now".
func NextIntervalDays(current, reviewCount, grade int) int {
	switch grade {
	case 0:
		return 0
	case 1:
		if current < 1 {
			return 1
		}
		return current
	case 2:
		if current < 1 {
			return 1
		}
		return current * 2
	default:
		if current < 1 {
			return 3
		}
		return current * 4
	}
}

// UserVocabularyService owns per-user word state: saving, reviewing,
// mastery, and unsaving. The saved-words hub lives on
// VocabularyItemService.Hub.
type UserVocabularyService interface {
	Save(ctx context.Context, userID string, req requests.SaveUserVocabulary) (*models.UserVocabulary, error)
	Review(ctx context.Context, userID, itemID string, grade int) (*models.UserVocabulary, error)
	Patch(ctx context.Context, userID, itemID string, req requests.PatchUserVocabulary) (*models.UserVocabulary, error)
	Unsave(ctx context.Context, userID, itemID string) error
}

type userVocabularyService struct {
	repo    repositories.UserVocabularyRepository
	entries repositories.VocabularyItemRepository
}

// NewUserVocabularyService builds the service over both repositories.
func NewUserVocabularyService(repo repositories.UserVocabularyRepository, entries repositories.VocabularyItemRepository) UserVocabularyService {
	return &userVocabularyService{repo: repo, entries: entries}
}

// Save stores per-user state idempotently via Upsert. The entry must be
// visible to the caller (curated or their own) — saving an invisible id is a
// 404, mirroring the deck-item Create pre-check. The first save enrols the
// word with a due date; re-saves never reset the review schedule.
func (s *userVocabularyService) Save(ctx context.Context, userID string, req requests.SaveUserVocabulary) (*models.UserVocabulary, error) {
	if _, err := s.entries.GetByID(ctx, userID, req.ItemID); err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "save vocabulary item lookup failed", appErr, "itemID", req.ItemID)
		return nil, appErr
	}
	u := &models.UserVocabulary{UserID: userID, ItemID: req.ItemID}
	if req.SourceID != nil {
		u.SourceID = req.SourceID
	}
	if req.Note != nil {
		u.Note = req.Note
	}
	u.SourceType = enums.VocabularySource(req.SourceType)
	updateMastered := req.Mastered != nil
	if updateMastered {
		u.Mastered = *req.Mastered
	}
	now := time.Now().UTC()
	u.SRSDueAt = &now // insert-only: the conflict UPDATE excludes the schedule columns
	if err := s.repo.Upsert(ctx, u, updateMastered); err != nil {
		appErr := common.FromDBError(err, "vocabulary item")
		logger.Report(ctx, "save user vocabulary failed", appErr, "itemID", req.ItemID)
		return nil, appErr
	}
	got, err := s.repo.Get(ctx, userID, req.ItemID)
	if err != nil {
		appErr := common.FromDBError(err, "user vocabulary")
		logger.Report(ctx, "read saved user vocabulary failed", appErr, "itemID", req.ItemID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "user vocabulary saved", "itemID", req.ItemID)
	return got, nil
}

// Review advances the SRS schedule for one saved word.
func (s *userVocabularyService) Review(ctx context.Context, userID, itemID string, grade int) (*models.UserVocabulary, error) {
	u, err := s.repo.Get(ctx, userID, itemID)
	if err != nil {
		appErr := common.FromDBError(err, "user vocabulary")
		logger.Report(ctx, "get user vocabulary for review failed", appErr, "itemID", itemID)
		return nil, appErr
	}
	now := time.Now().UTC()
	u.IntervalDays = NextIntervalDays(u.IntervalDays, u.ReviewCount, grade)
	u.ReviewCount++
	u.LastReviewedAt = &now
	due := now.AddDate(0, 0, u.IntervalDays)
	u.SRSDueAt = &due
	if err := s.repo.UpdateSchedule(ctx, u); err != nil {
		appErr := common.FromDBError(err, "user vocabulary")
		logger.Report(ctx, "update review schedule failed", appErr, "itemID", itemID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "user vocabulary reviewed", "itemID", itemID, "grade", grade, "intervalDays", u.IntervalDays)
	return u, nil
}

// Patch applies a partial update (mastery and/or note) to one saved word.
func (s *userVocabularyService) Patch(ctx context.Context, userID, itemID string, req requests.PatchUserVocabulary) (*models.UserVocabulary, error) {
	if req.Mastered == nil && req.Note == nil {
		return nil, common.BadRequest("nothing to update")
	}
	if err := s.repo.Patch(ctx, userID, itemID, req.Mastered, req.Note); err != nil {
		appErr := common.FromDBError(err, "user vocabulary")
		logger.Report(ctx, "patch user vocabulary failed", appErr, "itemID", itemID)
		return nil, appErr
	}
	u, err := s.repo.Get(ctx, userID, itemID)
	if err != nil {
		appErr := common.FromDBError(err, "user vocabulary")
		logger.Report(ctx, "read user vocabulary after patch failed", appErr, "itemID", itemID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "user vocabulary patched", "itemID", itemID)
	return u, nil
}

func (s *userVocabularyService) Unsave(ctx context.Context, userID, itemID string) error {
	if err := s.repo.Delete(ctx, userID, itemID); err != nil {
		appErr := common.FromDBError(err, "user vocabulary")
		logger.Report(ctx, "unsave user vocabulary failed", appErr, "itemID", itemID)
		return appErr
	}
	slog.InfoContext(ctx, "user vocabulary unsaved", "itemID", itemID)
	return nil
}
