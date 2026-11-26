package services

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/utils"
)

type VideoTranscriptService interface {
	List(ctx context.Context, limit, offset int) ([]*models.VideoTranscript, int64, error)
	GetByID(ctx context.Context, id, userID string) (*models.VideoTranscript, error)
	GetByVideoExerciseID(ctx context.Context, videoExerciseID, userID string) ([]*models.VideoTranscript, error)
	Create(ctx context.Context, req requests.CreateVideoTranscript) (*models.VideoTranscript, error)
	Update(ctx context.Context, id string, req requests.UpdateVideoTranscript) (*models.VideoTranscript, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type videoTranscriptService struct {
	repo      repositories.VideoTranscriptRepository
	exercises repositories.VideoExerciseRepository
	premium   common.PremiumChecker
}

func NewVideoTranscriptService(repo repositories.VideoTranscriptRepository, exercises repositories.VideoExerciseRepository, premium common.PremiumChecker) VideoTranscriptService {
	return &videoTranscriptService{repo: repo, exercises: exercises, premium: premium}
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

func (s *videoTranscriptService) GetByID(ctx context.Context, id, userID string) (*models.VideoTranscript, error) {
	transcript, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "failed to get video transcript", appErr, "id", id)
		return nil, appErr
	}
	if common.ContentGatingEnabled() && transcript.VideoExerciseID != "" {
		// Phase 0 gate (spec §5.4): captions are the exercise's content, so a
		// premium exercise blocks transcript reads for free users. Inert at
		// launch.
		if ex, err := s.exercises.GetByID(ctx, transcript.VideoExerciseID); err == nil {
			if appErr := common.RequirePremiumAccess(ctx, s.premium, userID, false, ex.IsPremium); appErr != nil {
				return nil, appErr
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := common.FromDBError(err, "video exercise")
			logger.Report(ctx, "gate video transcript failed", appErr, "id", id)
			return nil, appErr
		}
	}
	slog.InfoContext(ctx, "video transcript found", "id", transcript.ID)
	return transcript, nil
}

func (s *videoTranscriptService) List(ctx context.Context, limit, offset int) ([]*models.VideoTranscript, int64, error) {
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

func (s *videoTranscriptService) GetByVideoExerciseID(ctx context.Context, videoExerciseID, userID string) ([]*models.VideoTranscript, error) {
	if common.ContentGatingEnabled() {
		// Phase 0 gate (spec §5.4): same premium check as the single-transcript
		// read above. A missing exercise falls through to the transcript query
		// (empty list, exactly as before).
		if ex, err := s.exercises.GetByID(ctx, videoExerciseID); err == nil {
			if appErr := common.RequirePremiumAccess(ctx, s.premium, userID, false, ex.IsPremium); appErr != nil {
				return nil, appErr
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			appErr := common.FromDBError(err, "video exercise")
			logger.Report(ctx, "gate video transcripts failed", appErr, "video_exercise_id", videoExerciseID)
			return nil, appErr
		}
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
