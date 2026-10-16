package services

import (
	"context"
	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/utils"
	"log/slog"
)

type VideoCategoryService interface {
	List(ctx context.Context, limit, offset int) ([]*models.VideoCategory, int64, error)
	ListWithPreview(ctx context.Context, previewK int) ([]*models.VideoCategory, error)
	GetByID(ctx context.Context, id string) (*models.VideoCategory, error)
	Create(ctx context.Context, req requests.CreateVideoCategory) (*models.VideoCategory, error)
	Update(ctx context.Context, id string, req requests.UpdateVideoCategory) (*models.VideoCategory, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type videoCategoryService struct {
	repo repositories.VideoCategoryRepository
}

func NewVideoCategoryService(repo repositories.VideoCategoryRepository) VideoCategoryService {
	return &videoCategoryService{repo: repo}
}

func (s *videoCategoryService) Create(ctx context.Context, req requests.CreateVideoCategory) (*models.VideoCategory, error) {
	category := &models.VideoCategory{}
	if err := utils.Map(category, req); err != nil {
		logger.Report(ctx, "map create video category failed", err, "name", req.Name)
		return nil, common.Internal()
	}
	err := s.repo.Create(ctx, category)
	if err != nil {
		appErr := common.FromDBError(err, "video category")
		logger.Report(ctx, "failed to create video category", appErr, "name", req.Name)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video category created", "video_category_id", category.ID)
	return category, nil
}

func (s *videoCategoryService) Update(ctx context.Context, id string, req requests.UpdateVideoCategory) (*models.VideoCategory, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "video category")
		logger.Report(ctx, "failed to find video category for update", appErr, "id", id)
		return nil, appErr
	}
	if err := utils.Map(category, req); err != nil {
		logger.Report(ctx, "map update video category failed", err, "id", id)
		return nil, common.Internal()
	}
	if err := s.repo.Update(ctx, category); err != nil {
		appErr := common.FromDBError(err, "video category")
		logger.Report(ctx, "failed to update video category", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video category updated", "video_category_id", category.ID)
	return category, nil
}

func (s *videoCategoryService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)

	if err != nil {
		appErr := common.FromDBError(err, "video category")
		logger.Report(ctx, "failed to delete video category", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "video category deleted", "id", id)
	return nil
}

func (s *videoCategoryService) GetByID(ctx context.Context, id string) (*models.VideoCategory, error) {
	category, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "video category")
		logger.Report(ctx, "failed to get video category", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video category found", "video_category_id", category.ID)
	return category, nil
}

func (s *videoCategoryService) Count(ctx context.Context) (int64, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "video category")
		logger.Report(ctx, "failed to count video category", appErr)
		return 0, appErr
	}
	slog.InfoContext(ctx, "video category count", "count", count)
	return count, nil
}

func (s *videoCategoryService) List(ctx context.Context, limit, offset int) ([]*models.VideoCategory, int64, error) {
	categories, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "video category")
		logger.Report(ctx, "failed to list video categories", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.Count(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "video category")
		logger.Report(ctx, "failed to count video categories", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "video categories listed", "count", len(categories), "total", total)
	return categories, total, nil
}

func (s *videoCategoryService) ListWithPreview(ctx context.Context, previewK int) ([]*models.VideoCategory, error) {
	if previewK < 1 {
		previewK = common.VideoPreviewDefault
	}
	if previewK > common.VideoPreviewMax {
		previewK = common.VideoPreviewMax
	}
	categories, err := s.repo.ListWithPreview(ctx, previewK)
	if err != nil {
		appErr := common.FromDBError(err, "video category")
		logger.Report(ctx, "failed to list video categories with preview", appErr, "preview_k", previewK)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video categories with preview listed", "count", len(categories), "preview_k", previewK)
	return categories, nil
}
