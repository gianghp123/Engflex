package services

import (
	"context"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/utils"
	"log/slog"
)

type VideoExerciseService interface {
	List(ctx context.Context, limit int, offset int, isAdmin bool) ([]*models.VideoExercise, int64, error)
	GetByID(ctx context.Context, id string, isAdmin bool) (*models.VideoExercise, error)
	Create(ctx context.Context, req requests.CreateVideoExercise) (*models.VideoExercise, error)
	Update(ctx context.Context, id string, req requests.UpdateVideoExercise) (*models.VideoExercise, error)
	Delete(ctx context.Context, id string) error
}

type videoExerciseService struct {
	repo repositories.VideoExerciseRepository
}

func NewVideoExerciseService(repo repositories.VideoExerciseRepository) VideoExerciseService {
	return &videoExerciseService{repo: repo}
}

func (s *videoExerciseService) List(ctx context.Context, limit, offset int, isAdmin bool) ([]*models.VideoExercise, int64, error) {
	lessons, err := s.repo.List(ctx, limit, offset, !isAdmin)
	if err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to list video exercises", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.Count(ctx, !isAdmin)
	if err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to count video exercises", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "video exercises listed", "count", len(lessons), "total", total)
	return lessons, total, nil
}

// GetByID: a draft exercise is invisible to non-admins (404, existence not
// disclosed), same rule as vocabulary items/decks.
func (s *videoExerciseService) GetByID(ctx context.Context, id string, isAdmin bool) (*models.VideoExercise, error) {
	lesson, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "failed to get video exercise", appErr, "id", id)
		return nil, appErr
	}
	if lesson.Status != enums.PublishStatusPublished && !isAdmin {
		return nil, common.NotFound("video exercise not found")
	}
	slog.InfoContext(ctx, "video exercise found", "video_exercise_id", lesson.ID)
	return lesson, nil
}

// Create is admin-only (route-gated) and always starts as draft — only an
// explicit publish makes it visible to learners.
func (s *videoExerciseService) Create(ctx context.Context, req requests.CreateVideoExercise) (*models.VideoExercise, error) {
	lesson := &models.VideoExercise{}
	if err := utils.Map(lesson, req); err != nil {
		logger.Report(ctx, "map create video exercise failed", err)
		return nil, common.Internal()
	}
	lesson.Status = enums.PublishStatusDraft
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
	if req.Status != nil {
		lesson.Status = *req.Status
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
