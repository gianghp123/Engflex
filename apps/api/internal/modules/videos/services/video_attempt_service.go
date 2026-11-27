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
	"engflex-api/internal/utils"
	"engflex-api/internal/voice/pronunciation"
)

type VideoAttemptService interface {
	StartAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType, req requests.StartVideoAttempt, isAdmin bool) (*models.Attempt, error)
	GetOpenAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType) (*models.Attempt, error)
	SaveProgress(ctx context.Context, userID, videoExerciseID, attemptID string, req requests.SaveVideoProgress) (*models.Attempt, error)
	CompleteAttempt(ctx context.Context, userID, videoExerciseID, attemptID string, req requests.CompleteVideoAttempt) (*models.Attempt, error)
	AttachPracticeStatus(ctx context.Context, userID string, videos []*responses.VideoExerciseResponse) error
	Pronounce(ctx context.Context, userID, videoExerciseID, attemptID, transcriptID string, audio []byte, mime string) (*responses.VideoPronunciationResult, error)
}

type videoAttemptService struct {
	attempts     repositories.AttemptRepository
	exercises    repositories.VideoExerciseRepository
	transcripts  repositories.VideoTranscriptRepository
	transactions repositories.TransactionManager
	scorer       pronunciation.Scorer
	premium      common.PremiumChecker
}

func NewVideoAttemptService(attempts repositories.AttemptRepository, exercises repositories.VideoExerciseRepository, transcripts repositories.VideoTranscriptRepository, transactions repositories.TransactionManager, scorer pronunciation.Scorer, premium common.PremiumChecker) VideoAttemptService {
	return &videoAttemptService{attempts: attempts, exercises: exercises, transcripts: transcripts, transactions: transactions, scorer: scorer, premium: premium}
}

func (s *videoAttemptService) StartAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType, req requests.StartVideoAttempt, isAdmin bool) (*models.Attempt, error) {
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
	if common.ContentGatingEnabled() {
		// Phase 0 gate (spec §5.4): premium-flagged video exercises block
		// attempt starts for free users when content gating is on. Inert at
		// launch — the switch ships false — but the check must exist so a
		// future flag-on is airtight: the attempt route must not be the side
		// door around the exercise detail gate.
		if ex, err := s.exercises.GetByID(ctx, videoExerciseID); err == nil {
			if appErr := common.RequirePremiumAccess(ctx, s.premium, userID, false, ex.IsPremium); appErr != nil {
				return nil, appErr
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, videoAttemptError(ctx, "gate video attempt failed", videoExerciseID, err)
		}
		// A video exercise that does not exist still falls through to the
		// locked read, which maps it to 404 exactly as before.
	}

	var attempt *models.Attempt
	err := s.transactions.Run(ctx, func(tx *gorm.DB) error {
		// Lock the parent before looking for an open attempt, so simultaneous starts reuse it.
		// The same row answers the premium gate, so GetByID's rule (admin
		// preview, otherwise a live subscription) applies here too — closing
		// the gap where a free user could practice a premium video just by
		// never opening its detail page.
		exercise, err := s.exercises.WithTx(tx).GetByIDForUpdate(ctx, videoExerciseID)
		if err != nil {
			return common.FromDBError(err, "video exercise")
		}
		if err := common.RequirePremiumAccess(ctx, s.premium, userID, isAdmin, exercise.IsPremium); err != nil {
			return err
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
		var previous requests.SaveVideoProgress
		if err := json.Unmarshal(attempt.Result, &previous); err != nil {
			return err
		}
		if attempt.Type == enums.AttemptTypeDictation && (previous.DictationVersion == 1 || req.DictationVersion == 1) {
			if err := mergeDictationProgress(&req, previous, transcripts); err != nil {
				return err
			}
			attempt.Result, err = json.Marshal(req)
			return err
		}
		// Every other path (plain dictation, shadowing) merges forward: a
		// per-line entry already on the attempt — dictation's draft, or a
		// shadowing score Pronounce persisted — survives a checkpoint save
		// that doesn't touch that line, instead of being wiped by the empty
		// map the shadowing frontend always sends here.
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

// Pronounce scores one shadowing take against its transcript line, reusing
// the same engine/scorer as lesson speaking — only the reference text
// source differs (a video transcript line here, a speaking item there).
// The score+feedback is best-effort persisted onto the attempt row (see
// persistShadowingScore) so it survives a reload; persistence failure never
// fails the request since the engine call already paid the real cost and
// the learner is waiting on the result, not the save.
func (s *videoAttemptService) Pronounce(ctx context.Context, userID, videoExerciseID, attemptID, transcriptID string, audio []byte, mime string) (*responses.VideoPronunciationResult, error) {
	if _, err := uuid.Parse(videoExerciseID); err != nil {
		return nil, common.BadRequest("invalid video exercise id")
	}
	if _, err := uuid.Parse(attemptID); err != nil {
		return nil, common.BadRequest("invalid attempt id")
	}
	transcript, err := s.transcripts.GetByID(ctx, transcriptID)
	if err != nil {
		appErr := common.FromDBError(err, "video transcript")
		logger.Report(ctx, "pronounce transcript lookup failed", appErr, "transcriptID", transcriptID)
		return nil, appErr
	}
	if transcript.VideoExerciseID != videoExerciseID {
		return nil, common.BadRequest("transcript does not belong to this video")
	}
	attempt, err := s.attempts.GetByID(ctx, attemptID)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "pronounce attempt lookup failed", appErr, "attemptID", attemptID)
		return nil, appErr
	}
	if attempt.UserID != userID || attempt.VideoExerciseID == nil || *attempt.VideoExerciseID != videoExerciseID {
		return nil, common.NotFound("attempt not found")
	}
	if attempt.Type != enums.AttemptTypeShadowing {
		return nil, common.UnprocessableEntity("attempt is not a shadowing exercise")
	}
	if attempt.Status != enums.AttemptStatusInProgress {
		return nil, common.Conflict("attempt is closed")
	}
	out, err := s.scorer.Score(ctx, transcript.Content, audio, mime)
	if err != nil {
		return nil, err
	}
	var dto responses.VideoPronunciationResult
	if utils.Map(&dto, out) != nil {
		logger.Report(ctx, "pronounce map failed", common.Internal(), "transcriptID", transcriptID)
		return nil, common.Internal()
	}
	if err := s.persistShadowingScore(ctx, userID, videoExerciseID, attemptID, transcriptID, &dto); err != nil {
		logger.Report(ctx, "pronounce persist failed", common.Internal(), "attemptID", attemptID, "transcriptID", transcriptID)
	}
	slog.InfoContext(ctx, "video shadowing pronunciation scored",
		"videoExerciseID", videoExerciseID, "transcriptID", transcriptID, "score", dto.Score)
	return &dto, nil
}

// persistShadowingScore writes one line's score+feedback into
// Result.Transcripts[transcriptID], keyed the same way dictation keys its
// DictationLineResult entries, so a reload can show what the learner already
// scored on a line without re-recording.
func (s *videoAttemptService) persistShadowingScore(ctx context.Context, userID, videoExerciseID, attemptID, transcriptID string, dto *responses.VideoPronunciationResult) error {
	_, err := s.changeAttempt(ctx, userID, videoExerciseID, attemptID, "persist shadowing score", func(_ *gorm.DB, attempt *models.Attempt) error {
		var payload responses.VideoProgressPayload
		if len(attempt.Result) > 0 {
			if err := json.Unmarshal(attempt.Result, &payload); err != nil {
				return err
			}
		}
		if payload.Transcripts == nil {
			payload.Transcripts = map[string]any{}
		}
		payload.Transcripts[transcriptID] = responses.ShadowingLineResult{
			Score:         dto.Score,
			Feedback:      dto.Feedback,
			Transcription: dto.Transcription,
			AttemptedAt:   time.Now(),
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		attempt.Result = encoded
		return nil
	})
	return err
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
