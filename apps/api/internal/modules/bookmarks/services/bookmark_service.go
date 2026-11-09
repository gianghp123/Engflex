package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
)

type BookmarkService interface {
	ListByUser(ctx context.Context, userID, contentType string) ([]*models.Bookmark, error)
	Save(ctx context.Context, userID string, contentType enums.ContentType, contentID string, isAdmin bool) (*models.Bookmark, error)
	Unsave(ctx context.Context, userID, contentType, contentID string) error
	UpdateNote(ctx context.Context, userID, contentType, contentID string, note *string) error
	Library(ctx context.Context, userID, contentType, search string, isAdmin bool, limit, offset int) (*models.SavedLibrary, int64, error)
}

type bookmarkService struct {
	repo    repositories.BookmarkRepository
	library repositories.SavedItemRepository
	premium common.PremiumChecker
}

func NewBookmarkService(repo repositories.BookmarkRepository, library repositories.SavedItemRepository, premium common.PremiumChecker) BookmarkService {
	return &bookmarkService{repo: repo, library: library, premium: premium}
}

func (s *bookmarkService) ListByUser(ctx context.Context, userID, contentType string) ([]*models.Bookmark, error) {
	items, err := s.repo.ListByUser(ctx, userID, contentType)
	if err != nil {
		appErr := common.FromDBError(err, "bookmark")
		logger.Report(ctx, "list bookmarks failed", appErr, "userID", userID, "contentType", contentType)
		return nil, appErr
	}
	slog.InfoContext(ctx, "bookmarks listed", "count", len(items), "contentType", contentType)
	return items, nil
}

func (s *bookmarkService) Save(ctx context.Context, userID string, contentType enums.ContentType, contentID string, isAdmin bool) (*models.Bookmark, error) {
	if err := s.checkTarget(ctx, userID, contentType, contentID, isAdmin); err != nil {
		return nil, err
	}
	b := &models.Bookmark{UserID: userID, ContentType: contentType, ContentID: contentID}
	if err := s.repo.Create(ctx, b); err != nil {
		appErr := common.FromDBError(err, "bookmark")
		logger.Report(ctx, "save bookmark failed", appErr, "contentType", contentType, "contentID", contentID)
		return nil, appErr
	}
	got, err := s.repo.Get(ctx, userID, string(contentType), contentID)
	if err != nil {
		appErr := common.FromDBError(err, "bookmark")
		logger.Report(ctx, "read saved bookmark failed", appErr, "contentType", contentType, "contentID", contentID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "bookmark saved", "contentType", contentType, "contentID", contentID)
	return got, nil
}

func (s *bookmarkService) Unsave(ctx context.Context, userID, contentType, contentID string) error {
	if err := s.repo.Delete(ctx, userID, contentType, contentID); err != nil {
		appErr := common.FromDBError(err, "bookmark")
		logger.Report(ctx, "unsave bookmark failed", appErr, "contentType", contentType, "contentID", contentID)
		return appErr
	}
	slog.InfoContext(ctx, "bookmark unsaved", "contentType", contentType, "contentID", contentID)
	return nil
}

func (s *bookmarkService) UpdateNote(ctx context.Context, userID, contentType, contentID string, note *string) error {
	if err := s.repo.UpdateNote(ctx, userID, contentType, contentID, note); err != nil {
		appErr := common.FromDBError(err, "bookmark")
		logger.Report(ctx, "bookmark note update failed", appErr, "contentType", contentType, "contentID", contentID)
		return appErr
	}
	slog.InfoContext(ctx, "bookmark note updated", "contentType", contentType, "contentID", contentID)
	return nil
}

func (s *bookmarkService) Library(ctx context.Context, userID, contentType, search string, isAdmin bool, limit, offset int) (*models.SavedLibrary, int64, error) {
	isPremium := false
	if !isAdmin {
		var err error
		isPremium, err = s.premium.IsPremium(ctx, userID)
		if err != nil {
			logger.Report(ctx, "check saved library access failed", err)
			return nil, 0, err
		}
	}
	library, err := s.library.LibraryStats(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "saved library")
		logger.Report(ctx, "read saved library stats failed", appErr)
		return nil, 0, appErr
	}
	library.Items, err = s.library.ListLibrary(ctx, userID, contentType, search, isAdmin, isPremium, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "saved library")
		logger.Report(ctx, "list saved library failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.library.CountLibrary(ctx, userID, contentType, search, isAdmin, isPremium)
	if err != nil {
		appErr := common.FromDBError(err, "saved library")
		logger.Report(ctx, "count saved library failed", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "saved library listed", "count", len(library.Items))
	return library, total, nil
}

func (s *bookmarkService) checkTarget(ctx context.Context, userID string, contentType enums.ContentType, contentID string, isAdmin bool) error {
	item, err := s.library.GetTarget(ctx, string(contentType), contentID)
	if err != nil {
		appErr := common.FromDBError(err, "bookmark content")
		logger.Report(ctx, "bookmark target lookup failed", appErr, "contentID", contentID)
		return appErr
	}
	isPremium := false
	switch {
	case item.Lesson != nil:
		if !isAdmin && item.Lesson.Status != enums.PublishStatusPublished {
			return common.NotFound("bookmark content not found")
		}
		isPremium = item.Lesson.IsPremium
	case item.VideoExercise != nil:
		if !isAdmin && item.VideoExercise.Status != enums.PublishStatusPublished {
			return common.NotFound("bookmark content not found")
		}
		isPremium = item.VideoExercise.IsPremium
	case item.Story == nil:
		return common.NotFound("bookmark content not found")
	}
	if err := common.RequirePremiumAccess(ctx, s.premium, userID, isAdmin, isPremium); err != nil {
		logger.Report(ctx, "bookmark target access failed", err, "contentID", contentID)
		return err
	}
	slog.InfoContext(ctx, "bookmark target verified", "contentID", contentID)
	return nil
}
