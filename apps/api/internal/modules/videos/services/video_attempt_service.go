package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/videos/dtos/requests"
)

type VideoAttemptService interface {
	StartAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType, req requests.StartVideoAttempt) (*models.Attempt, error)
	GetOpenAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType) (*models.Attempt, error)
	SaveProgress(ctx context.Context, userID, videoExerciseID, attemptID string, req requests.SaveVideoProgress) (*models.Attempt, error)
	CompleteAttempt(ctx context.Context, userID, videoExerciseID, attemptID string, req requests.CompleteVideoAttempt) (*models.Attempt, error)
}

type VideoAccessPolicy interface {
	CheckAccess(ctx context.Context, userID, id string) error
}

type videoAttemptService struct {
	access    VideoAccessPolicy
	attempts  repositories.AttemptRepository
	exercises repositories.VideoExerciseRepository
}

func NewVideoAttemptService(attempts repositories.AttemptRepository, exercises repositories.VideoExerciseRepository, access VideoAccessPolicy) VideoAttemptService {
	return &videoAttemptService{attempts: attempts, exercises: exercises, access: access}
}

func (s *videoAttemptService) StartAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType, req requests.StartVideoAttempt) (*models.Attempt, error) {
	if err := s.access.CheckAccess(ctx, userID, videoExerciseID); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(videoExerciseID); err != nil {
		return nil, common.BadRequest("invalid video exercise id")
	}
	if _, err := s.exercises.GetByID(ctx, videoExerciseID); err != nil {
		appErr := common.FromDBError(err, "video exercise")
		logger.Report(ctx, "find video exercise failed", appErr, "videoExerciseID", videoExerciseID)
		return nil, appErr
	}
	if attemptType != enums.AttemptTypeDictation && attemptType != enums.AttemptTypeShadowing {
		return nil, common.BadRequest("invalid video practice type")
	}

	if open, err := s.attempts.GetVideoAttempt(ctx, userID, videoExerciseID, attemptType); err == nil {
		slog.InfoContext(ctx, "open video attempt reused", "attemptID", open.ID, "type", attemptType)
		return open, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "find open video attempt failed", appErr, "videoExerciseID", videoExerciseID)
		return nil, appErr
	}

	m := &models.Attempt{
		UserID:          userID,
		VideoExerciseID: &videoExerciseID,
		Type:            attemptType,
		Status:          enums.AttemptStatusInProgress,
		Result:          []byte(`{}`),
	}
	if err := s.attempts.CreateOpen(ctx, m); err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "start video attempt failed", appErr, "videoExerciseID", videoExerciseID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video attempt started", "attemptID", m.ID, "type", attemptType)
	return m, nil
}

func (s *videoAttemptService) GetOpenAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType) (*models.Attempt, error) {
	if err := s.access.CheckAccess(ctx, userID, videoExerciseID); err != nil {
		return nil, err
	}
	if _, err := uuid.Parse(videoExerciseID); err != nil {
		return nil, common.BadRequest("invalid video exercise id")
	}
	if attemptType != enums.AttemptTypeDictation && attemptType != enums.AttemptTypeShadowing {
		return nil, common.BadRequest("invalid video practice type")
	}
	m, err := s.attempts.GetVideoAttempt(ctx, userID, videoExerciseID, attemptType)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Report(ctx, "find open video attempt failed", appErr, "videoExerciseID", videoExerciseID)
		}
		return nil, appErr
	}
	slog.InfoContext(ctx, "open video attempt loaded", "attemptID", m.ID, "type", attemptType)
	return m, nil
}

func (s *videoAttemptService) loadOwnedAttempt(ctx context.Context, userID, attemptID string) (*models.Attempt, error) {
	if _, err := uuid.Parse(attemptID); err != nil {
		return nil, common.BadRequest("invalid attempt id")
	}
	m, err := s.attempts.GetByID(ctx, attemptID)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "load owned attempt failed", appErr, "attemptID", attemptID)
		return nil, appErr
	}
	if m.UserID != userID {
		return nil, common.NotFound("attempt not found")
	}
	if m.Status != enums.AttemptStatusInProgress {
		return nil, common.Conflict("attempt is closed")
	}
	return m, nil
}

func (s *videoAttemptService) SaveProgress(ctx context.Context, userID, videoExerciseID, attemptID string, req requests.SaveVideoProgress) (*models.Attempt, error) {
	if err := s.access.CheckAccess(ctx, userID, videoExerciseID); err != nil {
		return nil, err
	}
	m, err := s.loadOwnedAttempt(ctx, userID, attemptID)
	if err != nil {
		return nil, err
	}
	if m.VideoExerciseID == nil || *m.VideoExerciseID != videoExerciseID {
		return nil, common.BadRequest("attempt does not belong to this video exercise")
	}

	encoded, err := json.Marshal(req)
	if err != nil {
		logger.Report(ctx, "encode video progress failed", common.Internal(), "attemptID", m.ID)
		return nil, common.Internal()
	}
	if req.Revision != m.Revision {
		return nil, common.Conflict("attempt changed; reload before saving")
	}
	m.Result = encoded
	if err := s.attempts.SaveVideo(ctx, m, req.Revision); err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "save video progress failed", appErr, "attemptID", m.ID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video progress saved", "attemptID", m.ID)
	return m, nil
}

func (s *videoAttemptService) CompleteAttempt(ctx context.Context, userID, videoExerciseID, attemptID string, req requests.CompleteVideoAttempt) (*models.Attempt, error) {
	if err := s.access.CheckAccess(ctx, userID, videoExerciseID); err != nil {
		return nil, err
	}
	m, err := s.attempts.GetByID(ctx, attemptID)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "load video attempt for completion failed", appErr, "attemptID", attemptID)
		return nil, appErr
	}
	if m.VideoExerciseID == nil || *m.VideoExerciseID != videoExerciseID {
		return nil, common.BadRequest("attempt does not belong to this video exercise")
	}

	if m.UserID != userID {
		return nil, common.NotFound("attempt not found")
	}
	if m.Status == enums.AttemptStatusCompleted {
		if sameVideoCompletion(m, req) {
			slog.InfoContext(ctx, "video completion replayed", "attemptID", m.ID)
			return m, nil
		}
		return nil, common.Conflict("attempt is completed")
	}
	if m.Status != enums.AttemptStatusInProgress {
		return nil, common.Conflict("attempt is closed")
	}
	m.Status = enums.AttemptStatusCompleted
	now := time.Now().UTC()
	m.SubmittedAt = &now
	m.Score = req.Score
	m.DurationSec = req.DurationSec
	if err := s.attempts.SaveVideo(ctx, m, m.Revision); err != nil {
		appErr := common.FromDBError(err, "attempt")
		if appErr.Status == http.StatusConflict {
			latest, reloadErr := s.attempts.GetByID(ctx, attemptID)
			if reloadErr != nil {
				mapped := common.FromDBError(reloadErr, "attempt")
				logger.Report(ctx, "reload concurrent video completion failed", mapped)
				return nil, mapped
			}
			if latest.UserID == userID && latest.VideoExerciseID != nil && *latest.VideoExerciseID == videoExerciseID && latest.Status == enums.AttemptStatusCompleted && sameVideoCompletion(latest, req) {
				slog.InfoContext(ctx, "concurrent video completion replayed", "attemptID", latest.ID)
				return latest, nil
			}
		}
		logger.Report(ctx, "complete video attempt failed", appErr, "attemptID", m.ID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "video attempt completed", "attemptID", m.ID, "score", m.Score)
	return m, nil
}

func sameVideoCompletion(m *models.Attempt, req requests.CompleteVideoAttempt) bool {
	sameScore := m.Score == nil && req.Score == nil || m.Score != nil && req.Score != nil && *m.Score == *req.Score
	sameDuration := m.DurationSec == nil && req.DurationSec == nil || m.DurationSec != nil && req.DurationSec != nil && *m.DurationSec == *req.DurationSec
	return sameScore && sameDuration
}
