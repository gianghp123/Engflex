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
	"strings"
)

type VideoExerciseService interface {
	List(ctx context.Context, req requests.ListVideoExercises) ([]*models.VideoExercise, int64, error)
	GetByID(ctx context.Context, id string) (*models.VideoExercise, error)
	Create(ctx context.Context, req requests.CreateVideoExercise) (*models.VideoExercise, error)
	Update(ctx context.Context, id string, req requests.UpdateVideoExercise) (*models.VideoExercise, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type videoExerciseService struct {
	repo repositories.VideoExerciseRepository
}

func NewVideoExerciseService(repo repositories.VideoExerciseRepository) VideoExerciseService {
	return &videoExerciseService{repo: repo}
}

// List pages the catalog under the request's filters. Page and the total
// come from the same filter set, so the page count the client derives is
// exact. A blank search (only whitespace) matches everything.
func (s *videoExerciseService) List(ctx context.Context, req requests.ListVideoExercises) ([]*models.VideoExercise, int64, error) {
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = common.DefaultPage
	}
	if pageSize < 1 {
		pageSize = common.DefaultPageSize
	}
	offset := utils.Offset(page, pageSize)
	search := strings.TrimSpace(req.Search)

	lessons, err := s.repo.List(ctx, req.CategoryID, req.CEFRLevel, search, pageSize, offset)
	if err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to list video exercises", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.Count(ctx, req.CategoryID, req.CEFRLevel, search)
	if err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to count video exercises", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "video exercises listed", "count", len(lessons), "total", total, "page", page)
	return lessons, total, nil
}

func (s *videoExerciseService) GetByID(ctx context.Context, id string) (*models.VideoExercise, error) {
	lesson, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to get video exercise", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video exercise found", "video_exercise_id", lesson.ID)
	return lesson, nil
}

func (s *videoExerciseService) Create(ctx context.Context, req requests.CreateVideoExercise) (*models.VideoExercise, error) {
	lesson := &models.VideoExercise{}
	if err := utils.Map(lesson, req); err != nil {
		logger.Report(ctx, "map create video exercise failed", err)
		return nil, common.Internal()
	}
	err := s.repo.Create(ctx, lesson)
	if err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to create video exercise", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video exercise created", "video_exercise_id", lesson.ID)
	return lesson, nil
}

func (s *videoExerciseService) Update(ctx context.Context, id string, req requests.UpdateVideoExercise) (*models.VideoExercise, error) {
	lesson, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to get video exercise for update", appErr, "id", id)
		return nil, appErr
	}
	if err := utils.Map(lesson, req); err != nil {
		logger.Report(ctx, "map update video exercise failed", err, "id", id)
		return nil, common.Internal()
	}
	if err := s.repo.Update(ctx, lesson); err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to update video exercise", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video exercise updated", "video_exercise_id", lesson.ID)
	return lesson, nil
}

func (s *videoExerciseService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to delete video exercise", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "video exercise deleted", "id", id)
	return nil
}

func (s *videoExerciseService) Count(ctx context.Context) (int64, error) {
	count, err := s.repo.Count(ctx, nil, nil, "")
	if err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to count video exercise", appErr)
		return 0, appErr
	}
	slog.InfoContext(ctx, "video exercise count", "count", count)
	return count, nil
}
