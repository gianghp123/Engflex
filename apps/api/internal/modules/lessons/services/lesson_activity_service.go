package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/llm"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/lessons/activityconfig"
	"engflex-api/internal/modules/lessons/dtos/requests"
	"engflex-api/internal/modules/lessons/dtos/responses"
	"engflex-api/internal/modules/lessons/prompts"
	"engflex-api/internal/modules/lessons/schemas"
	"engflex-api/internal/utils"
	"engflex-api/internal/voice/pronunciation"
)

// LessonActivityService owns ordered lesson parts and lesson attempts
// (1 attempt = 1 lesson; per-part results attach to the open run).
type LessonActivityService interface {
	ListByLesson(ctx context.Context, lessonID string) ([]*models.LessonActivity, error)
	GetByID(ctx context.Context, id string) (*models.LessonActivity, error)
	CheckAnswer(ctx context.Context, userID, id, attemptID string, questionIndex int, key string) (*responses.CheckResult, error)
	Pronounce(ctx context.Context, userID, id, attemptID string, itemIndex int, audio []byte, mime string) (*responses.PronunciationResult, error)
	StartAttempt(ctx context.Context, userID, lessonID string) (*models.Attempt, error)
	GetOpenProgress(ctx context.Context, userID, lessonID string) (*responses.AttemptProgress, error)
	AttachPartResult(ctx context.Context, userID, attemptID, activityID string, entry PartEntry) (*models.Attempt, error)
	ScoreWriting(ctx context.Context, userID, id, attemptID, text string) (*responses.WritingScoreResponse, error)
	Create(ctx context.Context, req requests.CreateLessonActivity) (*models.LessonActivity, error)
	Update(ctx context.Context, id string, req requests.UpdateLessonActivity) (*models.LessonActivity, error)
	Delete(ctx context.Context, id string) error
}

type lessonActivityService struct {
	repo     repositories.LessonActivityRepository
	scorer   pronunciation.Scorer
	llm      llm.LLMClient
	attempts repositories.AttemptRepository
}

// NewLessonActivityService builds the service over repository and client
// interfaces (never concretes, never nil).
func NewLessonActivityService(repo repositories.LessonActivityRepository, scorer pronunciation.Scorer, llm llm.LLMClient, attempts repositories.AttemptRepository) LessonActivityService {
	return &lessonActivityService{repo: repo, scorer: scorer, llm: llm, attempts: attempts}
}

func (s *lessonActivityService) ListByLesson(ctx context.Context, lessonID string) ([]*models.LessonActivity, error) {
	items, err := s.repo.ListByLesson(ctx, lessonID)
	if err != nil {
		appErr := common.FromDBError(err, "lesson activity")
		logger.Report(ctx, "list lesson activities failed", appErr, "lessonID", lessonID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "lesson activities listed", "lessonID", lessonID, "count", len(items))
	return items, nil
}

func (s *lessonActivityService) GetByID(ctx context.Context, id string) (*models.LessonActivity, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "lesson activity")
		logger.Report(ctx, "get lesson activity failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "lesson activity found", "id", m.ID)
	return m, nil
}

// loadActivity fetches one activity, mapping DB errors to the API vocabulary.
func (s *lessonActivityService) loadActivity(ctx context.Context, id string) (*models.LessonActivity, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "lesson activity")
		logger.Report(ctx, "load lesson activity failed", appErr, "id", id)
		return nil, appErr
	}
	return m, nil
}

// StartAttempt opens one in_progress lesson run, or returns the existing open
// run (idempotent start: refresh/StrictMode retries never fork rows). Lesson
// existence rides the lessons FK (23503 -> 404 via FromDBError); malformed
// ids are 400 here.
func (s *lessonActivityService) StartAttempt(ctx context.Context, userID, lessonID string) (*models.Attempt, error) {
	if _, err := uuid.Parse(lessonID); err != nil {
		return nil, common.BadRequest("invalid lesson id")
	}
	if open, err := s.attempts.GetOpenAttempt(ctx, userID, lessonID); err == nil {
		return open, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "find open attempt failed", appErr, "lessonID", lessonID)
		return nil, appErr
	}
	m := &models.Attempt{
		UserID:   userID,
		LessonID: &lessonID,
		Type:     enums.AttemptTypeLesson,
		Status:   enums.AttemptStatusInProgress,
		Result:   []byte(`{}`),
	}
	if err := s.attempts.Create(ctx, m); err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "start attempt failed", appErr, "lessonID", lessonID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "attempt started", "lessonID", lessonID)
	return m, nil
}

// GetOpenProgress returns the open run with graded questions rehydrated so
// the client can resume mid-lesson. 404 when no run is open (the caller then
// starts one). Non-checkable entries are omitted; outcomes pointing past the
// current config are skipped as stale.
func (s *lessonActivityService) GetOpenProgress(ctx context.Context, userID, lessonID string) (*responses.AttemptProgress, error) {
	if _, err := uuid.Parse(lessonID); err != nil {
		return nil, common.BadRequest("invalid lesson id")
	}
	m, err := s.attempts.GetOpenAttempt(ctx, userID, lessonID)
	if err != nil {
		appErr := common.FromDBError(err, "attempt")
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Report(ctx, "find open attempt failed", appErr, "lessonID", lessonID)
		}
		return nil, appErr
	}
	checks := map[string][]responses.CheckedQuestion{}
	for activityID, entry := range decodePartMap(m) {
		am, err := s.repo.GetByID(ctx, activityID)
		if err != nil {
			logger.Report(ctx, "progress activity missing", common.FromDBError(err, "lesson activity"), "activityID", activityID)
			continue
		}
		questions, err := activityconfig.QuestionsFor(am.Type, am.Config)
		if errors.Is(err, activityconfig.ErrNotCheckable) {
			continue
		}
		if err != nil {
			logger.Report(ctx, "progress config failed", common.Internal(), "activityID", activityID)
			continue
		}
		var detail CheckPartDetail
		if raw, merr := json.Marshal(entry.Detail); merr != nil || json.Unmarshal(raw, &detail) != nil {
			continue
		}
		var out []responses.CheckedQuestion
		for _, o := range detail.Outcomes {
			if o.Index < 0 || o.Index >= len(questions) {
				continue
			}
			q := questions[o.Index]
			out = append(out, responses.CheckedQuestion{Index: o.Index, Key: o.Key, Correct: o.Correct, CorrectKey: q.AnswerKey, Explanation: q.Explanation})
		}
		if len(out) > 0 {
			checks[activityID] = out
		}
	}
	var summary responses.AttemptSummary
	if err := utils.Map(&summary, m); err != nil {
		logger.Report(ctx, "progress map failed", common.Internal(), "attemptID", m.ID)
		return nil, common.Internal()
	}
	slog.InfoContext(ctx, "attempt progress loaded", "attemptID", m.ID, "activities", len(checks))
	return &responses.AttemptProgress{Attempt: summary, Checks: checks}, nil
}

// loadOwnedAttempt fetches one attempt and proves ownership: foreign ids
// read as 404, never leaking existence. Malformed ids are 400 here (not via
// the repo) so every attach path shares one validation point.
func (s *lessonActivityService) loadOwnedAttempt(ctx context.Context, userID, attemptID string) (*models.Attempt, error) {
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

// AttachPartResult merges one part entry into the open attempt and finalizes
// inline when the map covers every lesson activity.
func (s *lessonActivityService) AttachPartResult(ctx context.Context, userID, attemptID, activityID string, entry PartEntry) (*models.Attempt, error) {
	m, err := s.loadOwnedAttempt(ctx, userID, attemptID)
	if err != nil {
		return nil, err
	}
	parts := decodePartMap(m)
	parts[activityID] = entry
	return s.attachWithMap(ctx, m, parts)
}

// attachWithMap persists one merged part entry, finalizing inline when the
// map covers every lesson activity.
func (s *lessonActivityService) attachWithMap(ctx context.Context, m *models.Attempt, parts map[string]PartEntry) (*models.Attempt, error) {
	encoded, err := encodePartMap(parts)
	if err != nil {
		logger.Report(ctx, "attach encode failed", common.Internal(), "attemptID", m.ID)
		return nil, common.Internal()
	}
	m.Result = encoded
	return s.maybeFinalize(ctx, m, parts)
}

// maybeFinalize completes the attempt when every lesson activity is present;
// otherwise it persists the merge. Score is the mean of part scores.
func (s *lessonActivityService) maybeFinalize(ctx context.Context, m *models.Attempt, parts map[string]PartEntry) (*models.Attempt, error) {
	if m.LessonID == nil {
		logger.Report(ctx, "attempt missing lesson", common.Internal(), "attemptID", m.ID)
		return nil, common.Internal()
	}
	activities, err := s.repo.ListByLesson(ctx, *m.LessonID)
	if err != nil {
		appErr := common.FromDBError(err, "lesson activity")
		logger.Report(ctx, "finalize list failed", appErr, "lessonID", *m.LessonID)
		return nil, appErr
	}
	complete := len(activities) > 0
	sum := 0.0
	for _, a := range activities {
		p, ok := parts[a.ID]
		if !ok {
			complete = false
			break
		}
		sum += p.Score
	}
	if !complete {
		if err := s.attempts.Update(ctx, m); err != nil {
			appErr := common.FromDBError(err, "attempt")
			logger.Report(ctx, "attach save failed", appErr, "attemptID", m.ID)
			return nil, appErr
		}
		slog.InfoContext(ctx, "part attached", "attemptID", m.ID)
		return m, nil
	}
	mean := sum / float64(len(activities))
	m.Score = &mean
	m.Status = enums.AttemptStatusCompleted
	if err := s.attempts.Update(ctx, m); err != nil {
		appErr := common.FromDBError(err, "attempt")
		logger.Report(ctx, "finalize save failed", appErr, "attemptID", m.ID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "attempt completed", "attemptID", m.ID, "score", mean)
	return m, nil
}

// CheckAnswer grades one question of a reading/listening activity. Answer keys
// live only in the config blob; nothing is returned until the caller commits.
func (s *lessonActivityService) CheckAnswer(ctx context.Context, userID, id, attemptID string, questionIndex int, key string) (*responses.CheckResult, error) {
	m, err := s.loadActivity(ctx, id)
	if err != nil {
		return nil, err
	}
	questions, err := activityconfig.QuestionsFor(m.Type, m.Config)
	if errors.Is(err, activityconfig.ErrNotCheckable) {
		return nil, common.UnprocessableEntity("activity is not checkable")
	}
	if err != nil {
		logger.Report(ctx, "check config failed", common.Internal(), "id", id)
		return nil, common.Internal()
	}
	if questionIndex < 0 || questionIndex >= len(questions) {
		return nil, common.BadRequest("question index out of range")
	}
	q := questions[questionIndex]
	correct := strings.EqualFold(strings.TrimSpace(key), strings.TrimSpace(q.AnswerKey))
	outcome := QuestionOutcome{Index: questionIndex, Key: key, Correct: correct}
	if _, err := s.AttachCheckOutcome(ctx, userID, attemptID, id, enums.AttemptType(m.Type), outcome); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "activity checked", "id", id, "questionIndex", questionIndex, "correct", correct)
	return &responses.CheckResult{Correct: correct, CorrectKey: q.AnswerKey, Explanation: q.Explanation}, nil
}

// AttachCheckOutcome merges one graded question into the part entry.
func (s *lessonActivityService) AttachCheckOutcome(ctx context.Context, userID, attemptID, activityID string, partType enums.AttemptType, outcome QuestionOutcome) (*models.Attempt, error) {
	m, err := s.loadOwnedAttempt(ctx, userID, attemptID)
	if err != nil {
		return nil, err
	}
	parts := decodePartMap(m)
	entry := parts[activityID]
	var detail CheckPartDetail
	if entry.Detail != nil {
		if raw, merr := json.Marshal(entry.Detail); merr == nil {
			_ = json.Unmarshal(raw, &detail)
		}
	}
	merged := false
	for i, o := range detail.Outcomes {
		if o.Index == outcome.Index {
			detail.Outcomes[i] = outcome
			merged = true
		}
	}
	if !merged {
		detail.Outcomes = append(detail.Outcomes, outcome)
	}
	correct := 0
	for _, o := range detail.Outcomes {
		if o.Correct {
			correct++
		}
	}
	score := 0.0
	if len(detail.Outcomes) > 0 {
		score = float64(correct) / float64(len(detail.Outcomes)) * 100
	}
	parts[activityID] = PartEntry{Type: partType, Score: score, Detail: detail}
	return s.attachWithMap(ctx, m, parts)
}

// AttachPronounceItem merges one item score (latest wins) into the part entry.
func (s *lessonActivityService) AttachPronounceItem(ctx context.Context, userID, attemptID, activityID string, itemIndex int, score float64) (*models.Attempt, error) {
	m, err := s.loadOwnedAttempt(ctx, userID, attemptID)
	if err != nil {
		return nil, err
	}
	parts := decodePartMap(m)
	entry := parts[activityID]
	var detail PronouncePartDetail
	if entry.Detail != nil {
		if raw, merr := json.Marshal(entry.Detail); merr == nil {
			_ = json.Unmarshal(raw, &detail)
		}
	}
	if detail.Items == nil {
		detail.Items = map[string]float64{}
	}
	detail.Items[strconv.Itoa(itemIndex)] = score
	sum := 0.0
	for _, v := range detail.Items {
		sum += v
	}
	mean := sum / float64(len(detail.Items))
	parts[activityID] = PartEntry{Type: enums.AttemptTypeSpeaking, Score: mean, Detail: detail}
	return s.attachWithMap(ctx, m, parts)
}

// Pronounce scores one read-aloud attempt. It resolves the reference text
// from the speaking config and delegates scoring to the shared scorer.
func (s *lessonActivityService) Pronounce(ctx context.Context, userID, id, attemptID string, itemIndex int, audio []byte, mime string) (*responses.PronunciationResult, error) {
	m, err := s.loadActivity(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.Type != enums.ActivityTypeSpeaking {
		return nil, common.UnprocessableEntity("activity is not a speaking exercise")
	}
	cfg, perr := activityconfig.ParseSpeakingConfig(m.Config)
	if perr != nil {
		logger.Report(ctx, "pronounce config failed", common.Internal(), "id", id)
		return nil, common.Internal()
	}
	if itemIndex < 0 || itemIndex >= len(cfg.Items) {
		return nil, common.BadRequest("item index out of range")
	}
	if _, err := s.loadOwnedAttempt(ctx, userID, attemptID); err != nil {
		return nil, err
	}
	out, err := s.scorer.Score(ctx, cfg.Items[itemIndex].Text, audio, mime)
	if err != nil {
		return nil, err
	}
	var dto responses.PronunciationResult
	if utils.Map(&dto, out) != nil {
		logger.Report(ctx, "pronounce map failed", common.Internal(), "id", id)
		return nil, common.Internal()
	}
	if _, err := s.AttachPronounceItem(ctx, userID, attemptID, id, itemIndex, dto.Score); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "pronunciation scored", "activityID", id, "score", dto.Score)
	return &dto, nil
}

// ScoreWriting grades one writing part with a stateless LLM call and attaches
// the result to the open lesson attempt. Any LLM or shape failure returns an
// error and writes nothing.
func (s *lessonActivityService) ScoreWriting(ctx context.Context, userID, id, attemptID, text string) (*responses.WritingScoreResponse, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, common.BadRequest("text is required")
	}
	words := countWords(trimmed)
	if words > maxWritingWords {
		return nil, common.New(http.StatusRequestEntityTooLarge, "text too long")
	}
	m, err := s.loadActivity(ctx, id)
	if err != nil {
		return nil, err
	}
	if m.Type != enums.ActivityTypeWriting {
		return nil, common.UnprocessableEntity("activity is not a writing exercise")
	}
	cfg, perr := activityconfig.ParseWritingConfig(m.Config)
	if perr != nil {
		logger.Report(ctx, "writing config failed", common.Internal(), "id", id)
		return nil, common.Internal()
	}
	if _, err := s.loadOwnedAttempt(ctx, userID, attemptID); err != nil {
		return nil, err
	}
	prompt := prompts.BuildWritingScorePrompt(cfg, trimmed, words)
	var out schemas.WritingFeedback
	if err := s.llm.CompleteStructured(ctx, prompt, &out, schemas.WritingFeedbackSchemaName, schemas.WritingFeedbackSchema); err != nil {
		logger.Report(ctx, "writing score engine call failed", common.ServiceUnavailable("scoring unavailable"), "activityID", id)
		return nil, common.ServiceUnavailable("scoring unavailable")
	}
	scored, verr := schemas.NormalizeWritingFeedback(out, trimmed, len(cfg.Checklist))
	if verr != nil {
		logger.Report(ctx, "writing score rejected", common.ServiceUnavailable("scoring unavailable"), "activityID", id, "reason", verr.Error())
		return nil, common.ServiceUnavailable("scoring unavailable")
	}
	score := coverageScore(scored.Task.Verdicts)
	attached, err := s.AttachPartResult(ctx, userID, attemptID, id, PartEntry{Type: enums.AttemptTypeWriting, Score: score, Detail: scored})
	if err != nil {
		return nil, err
	}
	var dto responses.WritingScoreResponse
	if err := utils.Map(&dto, scored); err != nil {
		logger.Report(ctx, "writing map failed", common.Internal(), "id", id)
		return nil, common.Internal()
	}
	dto.Summary = responses.ScoreSummary{
		PointsCovered: countCovered(scored.Task.Verdicts),
		PointsTotal:   len(scored.Task.Verdicts),
		Score:         score,
		WordCount:     words,
		MinWords:      cfg.MinWords,
		MaxWords:      cfg.MaxWords,
	}
	dto.Attempt = responses.AttemptSummary{ID: attached.ID, Status: string(attached.Status), Score: attached.Score}
	slog.InfoContext(ctx, "writing scored", "activityID", id, "score", score)
	return &dto, nil
}

// validateActivityConfig parses the raw config against the shape its type
// demands, so a malformed admin-authored document never reaches the
// column (and only ever fails loudly, not silently, at check/pronounce/
// score time downstream).
func validateActivityConfig(t enums.ActivityType, raw []byte) error {
	cfg := datatypes.JSON(raw)
	var err error
	switch t {
	case enums.ActivityTypeReading:
		_, err = activityconfig.ParseReadingConfig(cfg)
	case enums.ActivityTypeListening:
		_, err = activityconfig.ParseListeningConfig(cfg)
	case enums.ActivityTypeWriting:
		_, err = activityconfig.ParseWritingConfig(cfg)
	case enums.ActivityTypeSpeaking:
		_, err = activityconfig.ParseSpeakingConfig(cfg)
	default:
		return fmt.Errorf("unknown activity type %q", t)
	}
	if err != nil {
		return fmt.Errorf("invalid config for type %q: %w", t, err)
	}
	return nil
}

// Create adds one ordered part to a lesson (admin-only, route-gated).
func (s *lessonActivityService) Create(ctx context.Context, req requests.CreateLessonActivity) (*models.LessonActivity, error) {
	if err := validateActivityConfig(req.Type, req.Config); err != nil {
		return nil, common.BadRequest(err.Error())
	}
	m := &models.LessonActivity{}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map create lesson activity failed", err)
		return nil, common.Internal()
	}
	// Config is datatypes.JSON (itself a Scanner/Valuer), so it gets the
	// same explicit []byte conversion as every other jsonb-backed field
	// instead of riding the blanket utils.Map above.
	m.Config = datatypes.JSON(req.Config)
	if err := s.repo.Create(ctx, m); err != nil {
		appErr := common.FromDBError(err, "lesson activity")
		logger.Report(ctx, "create lesson activity failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "lesson activity created", "id", m.ID)
	return m, nil
}

func (s *lessonActivityService) Update(ctx context.Context, id string, req requests.UpdateLessonActivity) (*models.LessonActivity, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "lesson activity")
		logger.Report(ctx, "get lesson activity for update failed", appErr, "id", id)
		return nil, appErr
	}
	effectiveType := m.Type
	if req.Type != nil {
		effectiveType = *req.Type
	}
	if req.Config != nil {
		if err := validateActivityConfig(effectiveType, req.Config); err != nil {
			return nil, common.BadRequest(err.Error())
		}
	}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map update lesson activity failed", err, "id", id)
		return nil, common.Internal()
	}
	if req.Config != nil {
		m.Config = datatypes.JSON(req.Config)
	}
	if err := s.repo.Update(ctx, m); err != nil {
		appErr := common.FromDBError(err, "lesson activity")
		logger.Report(ctx, "update lesson activity failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "lesson activity updated", "id", id)
	return m, nil
}

func (s *lessonActivityService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "lesson activity")
		logger.Report(ctx, "delete lesson activity failed", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "lesson activity deleted", "id", id)
	return nil
}
