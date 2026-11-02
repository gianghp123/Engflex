package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/utils"
)

type VideoTranscriptService interface {
	List(ctx context.Context, limit, offset int) ([]*models.VideoTranscript, int64, error)
	GetByID(ctx context.Context, id string) (*models.VideoTranscript, error)
	GetByVideoExerciseID(ctx context.Context, videoExerciseID string) ([]*models.VideoTranscript, error)
	Create(ctx context.Context, req requests.CreateVideoTranscript) (*models.VideoTranscript, error)
	Update(ctx context.Context, id string, req requests.UpdateVideoTranscript) (*models.VideoTranscript, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type videoTranscriptService struct {
	access VideoAccessPolicy
	repo   repositories.VideoTranscriptRepository
}

func NewVideoTranscriptService(repo repositories.VideoTranscriptRepository, access VideoAccessPolicy) VideoTranscriptService {
	return &videoTranscriptService{repo: repo, access: access}
}

func (s *videoTranscriptService) Create(ctx context.Context, req requests.CreateVideoTranscript) (*models.VideoTranscript, error) {
	transcript := &models.VideoTranscript{}
	if err := utils.Map(transcript, req); err != nil {
		logger.Report(ctx, "map create video transcript failed", err)
		return nil, common.Internal()
	}
	err := s.repo.Create(ctx, transcript)
	if err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "failed to create video transcript", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video transcript created", "id", transcript.ID)
	return transcript, nil
}

func (s *videoTranscriptService) Update(ctx context.Context, id string, req requests.UpdateVideoTranscript) (*models.VideoTranscript, error) {
	transcript, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "failed to get video transcript for update", appErr, "id", id)
		return nil, appErr
	}
	// UpdateVideoTranscript fields are pointers; copier skips nil sources,
	// so one Map call is the partial update — no manual nil checks.
	if err := utils.Map(transcript, req); err != nil {
		logger.Report(ctx, "map update video transcript failed", err, "id", id)
		return nil, common.Internal()
	}
	if err = s.repo.Update(ctx, transcript); err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "failed to update video transcript", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video transcript updated", "id", transcript.ID)
	return transcript, nil
}

func (s *videoTranscriptService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "failed to delete video transcript", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "video transcript deleted", "id", id)
	return nil
}

func (s *videoTranscriptService) GetByID(ctx context.Context, id string) (*models.VideoTranscript, error) {
	transcript, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "failed to get video transcript", appErr, "id", id)
		return nil, appErr
	}
	if err := s.access.CheckAccess(ctx, common.ActorFromContext(ctx).UserID, transcript.VideoExerciseID); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "video transcript found", "id", transcript.ID)
	return transcript, nil
}

func (s *videoTranscriptService) List(ctx context.Context, limit, offset int) ([]*models.VideoTranscript, int64, error) {
	if !common.ActorFromContext(ctx).Admin {
		return nil, 0, common.Forbidden("admin required")
	}
	transcripts, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "failed to list video transcripts", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.Count(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "failed to count video transcripts", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "video transcripts listed", "count", len(transcripts), "total", total)
	return transcripts, total, nil
}

func (s *videoTranscriptService) GetByVideoExerciseID(ctx context.Context, videoExerciseID string) ([]*models.VideoTranscript, error) {
	if err := s.access.CheckAccess(ctx, common.ActorFromContext(ctx).UserID, videoExerciseID); err != nil {
		return nil, err
	}
	transcripts, err := s.repo.GetByVideoExerciseID(ctx, videoExerciseID)
	if err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "failed to get video transcripts by video exercise id", appErr, "video_exercise_id", videoExerciseID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video transcripts found by video exercise id", "video_exercise_id", videoExerciseID)
	return transcripts, nil
}

func (s *videoTranscriptService) Count(ctx context.Context) (int64, error) {
	count, err := s.repo.Count(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "failed to count video transcripts", appErr)
		return 0, appErr
	}
	slog.InfoContext(ctx, "video transcripts count", "count", count)
	return count, nil
}
