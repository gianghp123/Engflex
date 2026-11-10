package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/modules/videos/dtos/responses"
)

type VideoAttemptService interface {
	StartAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType, req requests.StartVideoAttempt) (*models.Attempt, error)
	GetOpenAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType) (*models.Attempt, error)
	SaveProgress(ctx context.Context, userID, videoExerciseID, attemptID string, req requests.SaveVideoProgress) (*models.Attempt, error)
	CompleteAttempt(ctx context.Context, userID, videoExerciseID, attemptID string, req requests.CompleteVideoAttempt) (*models.Attempt, error)
	AttachPracticeStatus(ctx context.Context, userID string, videos []*responses.VideoExerciseResponse) error
}

type videoAttemptService struct {
	attempts     repositories.AttemptRepository
	exercises    repositories.VideoExerciseRepository
	transcripts  repositories.VideoTranscriptRepository
	transactions repositories.TransactionManager
}

func NewVideoAttemptService(attempts repositories.AttemptRepository, exercises repositories.VideoExerciseRepository, transcripts repositories.VideoTranscriptRepository, transactions repositories.TransactionManager) VideoAttemptService {
	return &videoAttemptService{attempts: attempts, exercises: exercises, transcripts: transcripts, transactions: transactions}
}

func (s *videoAttemptService) StartAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType, req requests.StartVideoAttempt) (*models.Attempt, error) {
	if _, err := uuid.Parse(videoExerciseID); err != nil {
		return nil, common.BadRequest("invalid video exercise id")
	}
	if attemptType != enums.AttemptTypeDictation && attemptType != enums.AttemptTypeShadowing {
		return nil, common.BadRequest("invalid video practice type")
	}
	if req.Level != nil && (attemptType != enums.AttemptTypeDictation ||
		(*req.Level != "easy" && *req.Level != "normal" && *req.Level != "hard")) {
		return nil, common.BadRequest("invalid dictation level")
	}

	var attempt *models.Attempt
	err := s.transactions.Run(ctx, func(tx *gorm.DB) error {
		// Lock the parent before looking for an open attempt, so simultaneous starts reuse it.
		if _, err := s.exercises.WithTx(tx).GetByIDForUpdate(ctx, videoExerciseID); err != nil {
			return common.FromDBError(err, "video exercise")
		}
		attempts := s.attempts.WithTx(tx)
		open, err := attempts.GetVideoAttempt(ctx, userID, videoExerciseID, attemptType)
		if err == nil {
			attempt = open
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return common.FromDBError(err, "attempt")
		}
		attempt = &models.Attempt{
			UserID:          userID,
			VideoExerciseID: &videoExerciseID,
			Type:            attemptType,
			Status:          enums.AttemptStatusInProgress,
			Result:          []byte(`{}`),
		}
		if req.Level != nil {
			attempt.Result, err = json.Marshal(requests.SaveVideoProgress{Level: req.Level, DictationVersion: 1})
			if err != nil {
				return err
			}
		}
		return attempts.Create(ctx, attempt)
	})
	if err != nil {
		return nil, videoAttemptError(ctx, "start video attempt failed", videoExerciseID, err)
	}
	slog.InfoContext(ctx, "video attempt opened", "attemptID", attempt.ID, "type", attemptType)
	return attempt, nil
}

func (s *videoAttemptService) GetOpenAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType) (*models.Attempt, error) {
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

// changeAttempt serializes checkpoint writes and completion on the same attempt row.
func (s *videoAttemptService) changeAttempt(
	ctx context.Context, userID, videoExerciseID, attemptID, operation string,
	change func(*gorm.DB, *models.Attempt) error,
) (*models.Attempt, error) {
	if _, err := uuid.Parse(attemptID); err != nil {
		return nil, common.BadRequest("invalid attempt id")
	}
	if _, err := uuid.Parse(videoExerciseID); err != nil {
		return nil, common.BadRequest("invalid video exercise id")
	}
	var attempt *models.Attempt
	err := s.transactions.Run(ctx, func(tx *gorm.DB) error {
		attempts := s.attempts.WithTx(tx)
		var err error
		attempt, err = attempts.GetVideoAttemptForUpdate(ctx, userID, videoExerciseID, attemptID)
		if err != nil {
			return err
		}
		if attempt.Type == enums.AttemptTypeDictation && attempt.Status == enums.AttemptStatusCompleted {
			return nil
		}
		if attempt.Status != enums.AttemptStatusInProgress {
			return common.Conflict("attempt is closed")
		}
		if err := change(tx, attempt); err != nil {
			return err
		}
		return attempts.Update(ctx, attempt)
	})
	if err != nil {
		return nil, videoAttemptError(ctx, operation+" failed", attemptID, err)
	}
	slog.InfoContext(ctx, operation, "attemptID", attemptID)
	return attempt, nil
}

func videoAttemptError(ctx context.Context, operation, id string, err error) error {
	var appErr *common.AppError
	if !errors.As(err, &appErr) {
		appErr = common.FromDBError(err, "attempt")
	}
	logger.Report(ctx, operation, appErr, "id", id)
	return appErr
}

func (s *videoAttemptService) SaveProgress(ctx context.Context, userID, videoExerciseID, attemptID string, req requests.SaveVideoProgress) (*models.Attempt, error) {
	return s.changeAttempt(ctx, userID, videoExerciseID, attemptID, "save video progress", func(tx *gorm.DB, attempt *models.Attempt) error {
		transcripts, err := s.transcripts.WithTx(tx).GetByVideoExerciseID(ctx, videoExerciseID)
		if err != nil {
			return err
		}
		valid := make(map[string]bool, len(transcripts))
		for _, transcript := range transcripts {
			valid[transcript.ID] = true
		}
		if req.ActiveTranscriptID != nil && !valid[*req.ActiveTranscriptID] {
			return common.BadRequest("active transcript does not belong to this video")
		}
		for id := range req.Transcripts {
			if !valid[id] {
				return common.BadRequest("transcript does not belong to this video")
			}
		}
		completed := make(map[string]bool, len(req.CompletedTranscriptIDs))
		for _, id := range req.CompletedTranscriptIDs {
			if !valid[id] {
				return common.BadRequest("completed transcript does not belong to this video")
			}
			completed[id] = true
		}
		if attempt.Type == enums.AttemptTypeDictation {
			var previous requests.SaveVideoProgress
			if err := json.Unmarshal(attempt.Result, &previous); err != nil {
				return err
			}
			if previous.DictationVersion == 1 || req.DictationVersion == 1 {
				if err := mergeDictationProgress(&req, previous, transcripts); err != nil {
					return err
				}
				attempt.Result, err = json.Marshal(req)
				return err
			}
			if req.Transcripts == nil {
				req.Transcripts = make(map[string]any)
			}
			for id, state := range previous.Transcripts {
				if _, changed := req.Transcripts[id]; valid[id] && !changed {
					req.Transcripts[id] = state
				}
			}
			for _, id := range previous.CompletedTranscriptIDs {
				if valid[id] {
					completed[id] = true
				}
			}
		}
		req.CompletedTranscriptIDs = make([]string, 0, len(completed))
		for _, transcript := range transcripts {
			if completed[transcript.ID] {
				req.CompletedTranscriptIDs = append(req.CompletedTranscriptIDs, transcript.ID)
			}
		}
		encoded, err := json.Marshal(req)
		if err != nil {
			return err
		}
		attempt.Result = encoded
		return nil
	})
}

func (s *videoAttemptService) CompleteAttempt(ctx context.Context, userID, videoExerciseID, attemptID string, req requests.CompleteVideoAttempt) (*models.Attempt, error) {
	return s.changeAttempt(ctx, userID, videoExerciseID, attemptID, "complete video attempt", func(tx *gorm.DB, attempt *models.Attempt) error {
		score := req.Score
		if attempt.Type == enums.AttemptTypeDictation {
			var progress requests.SaveVideoProgress
			if err := json.Unmarshal(attempt.Result, &progress); err != nil {
				return err
			}
			transcripts, err := s.transcripts.WithTx(tx).GetByVideoExerciseID(ctx, videoExerciseID)
			if err != nil {
				return err
			}
			if len(transcripts) == 0 {
				return common.Conflict("video has no transcripts")
			}
			score, err = dictationScore(progress, transcripts)
			if err != nil {
				return err
			}
		}
		now := time.Now()
		attempt.Status = enums.AttemptStatusCompleted
		attempt.Score = score
		attempt.DurationSec = req.DurationSec
		attempt.SubmittedAt = &now
		return nil
	})
}

func (s *videoAttemptService) AttachPracticeStatus(ctx context.Context, userID string, videos []*responses.VideoExerciseResponse) error {
	if len(videos) == 0 || userID == "" {
		return nil
	}
	ids := make([]string, len(videos))
	for i, v := range videos {
		ids[i] = v.ID
	}
	attempts, err := s.attempts.ListCompletedVideo(ctx, userID, ids)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "list completed video attempts failed", appErr, "userID", userID)
		return appErr
	}

	doneMap := make(map[string]map[enums.AttemptType]bool)
	for _, a := range attempts {
		if a.VideoExerciseID == nil {
			continue
		}
		vid := *a.VideoExerciseID
		if doneMap[vid] == nil {
			doneMap[vid] = make(map[enums.AttemptType]bool)
		}
		doneMap[vid][a.Type] = true
	}

	for _, v := range videos {
		dm := doneMap[v.ID]
		v.Practice = &responses.VideoPracticeStatus{
			Dictation: dm[enums.AttemptTypeDictation],
			Shadowing: dm[enums.AttemptTypeShadowing],
		}
	}
	return nil
}
