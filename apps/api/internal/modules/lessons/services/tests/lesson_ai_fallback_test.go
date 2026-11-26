// Completion-critical AI contract (spec §3.3, §8.2): an AI outage must
// never block lesson completion. The fallback part attaches, the attempt
// finalizes, the error is retryable, and a retry overwrites the part.
package tests

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/lessons/schemas"
	"engflex-api/internal/modules/lessons/services"
	"engflex-api/internal/voice/pronunciation"
)

// fakeScorer satisfies pronunciation.Scorer with a canned failure. The
// writing paths never reach the scorer; it exists for constructor
// completeness (the house rule bans nil into constructors).
type fakeScorer struct{ err error }

func (f fakeScorer) Score(ctx context.Context, text string, audio []byte, mime string) (*pronunciation.Result, error) {
	return nil, f.err
}

// fakeLLM satisfies llm.LLMClient. err fails the call; otherwise out is
// copied into the output so success paths can normalize a valid reply.
// (A zero reply can never normalize: Suggested/Expressions are required.)
type fakeLLM struct {
	out schemas.WritingFeedback
	err error
}

func (f fakeLLM) CompleteText(ctx context.Context, prompt string) (string, error) { return "", f.err }
func (f fakeLLM) CompleteStructured(ctx context.Context, prompt string, output any, schemaName string, schema map[string]any) error {
	if f.err != nil {
		return f.err
	}
	if o, ok := output.(*schemas.WritingFeedback); ok {
		*o = f.out
	}
	return nil
}

// retryFeedback is a valid reply for the one-checklist writing activity
// below: two verdicts (covered + partial = 75), no spans to substring-check.
func retryFeedback() schemas.WritingFeedback {
	return schemas.WritingFeedback{
		Task: schemas.TaskCompletion{
			Verdicts: []schemas.TaskVerdict{
				{Item: "state an opinion", Status: "covered", Evidence: "I think it is good"},
				{Item: "give a reason", Status: "partial", Evidence: "because it is cheap"},
			},
			AnswersPrompt: true,
			AnswersNote:   "States an opinion with a short reason.",
		},
		Expressions: []string{"I think"},
		Suggested:   "I think it is good because it is cheap and reliable.",
		Tip:         "Add one more reason with because.",
	}
}

func baseAttempt() *models.Attempt {
	return &models.Attempt{
		ID:       "00000000-0000-0000-0000-0000000000a1",
		UserID:   "user_1",
		LessonID: strPtr("00000000-0000-0000-0000-0000000000b1"),
		Status:   enums.AttemptStatusInProgress,
		Result:   []byte(`{}`),
	}
}

func strPtr(s string) *string { return &s }

func writingActivity() *models.LessonActivity {
	return &models.LessonActivity{
		ID:     "00000000-0000-0000-0000-0000000000c1",
		Type:   enums.ActivityTypeWriting,
		Config: []byte(`{"checklist":["state an opinion","give a reason"],"minWords":1,"maxWords":500}`),
	}
}

func TestDecodePartMapDefaultsLegacyStatus(t *testing.T) {
	m := &models.Attempt{Result: []byte(`{"00000000-0000-0000-0000-0000000000c1":{"type":"writing","score":80,"detail":{"x":1}}}`)}
	parts := services.DecodePartMapForTest(m)
	require.Len(t, parts, 1)
	entry := parts["00000000-0000-0000-0000-0000000000c1"]
	assert.Equal(t, services.PartStatusScored, entry.Status, "rows written before Status existed decode as scored")
	assert.Equal(t, 80.0, entry.Score)
}

func TestScoreWritingStillAttachesPartWhenLLMDown(t *testing.T) {
	// One writing activity only: attaching the fallback part must finalize
	// the attempt (mean score 0), and the error must be a retryable 503
	// carrying CodeAIUnavailable.
	activities := repomocks.NewMockLessonActivityRepository(t)
	activities.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-0000000000c1").Return(writingActivity(), nil).Once()
	activities.EXPECT().ListByLesson(mock.Anything, "00000000-0000-0000-0000-0000000000b1").
		Return([]*models.LessonActivity{writingActivity()}, nil).Maybe()
	attempts := repomocks.NewMockAttemptRepository(t)
	attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-0000000000a1").Return(baseAttempt(), nil).Once()
	saved := make([]*models.Attempt, 0)
	attempts.EXPECT().Update(mock.Anything, mock.Anything).Run(func(ctx context.Context, m *models.Attempt) {
		saved = append(saved, m)
	}).Return(nil).Once()

	svc := services.NewLessonActivityService(activities, fakeScorer{}, fakeLLM{err: errors.New("llm down")}, attempts,
		repomocks.NewMockLessonRepository(t), fakeNotPremium{})
	dto, err := svc.ScoreWriting(context.Background(), "user_1", "00000000-0000-0000-0000-0000000000c1", "00000000-0000-0000-0000-0000000000a1", "I think it is good")
	require.Nil(t, dto)
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 503, appErr.Status)
	require.Equal(t, common.CodeAIUnavailable, appErr.Code)

	require.Len(t, saved, 1)
	m := saved[0]
	require.Equal(t, enums.AttemptStatusCompleted, m.Status, "the lesson must still finalize")
	var parts map[string]services.PartEntry
	require.NoError(t, json.Unmarshal(m.Result, &parts))
	entry := parts["00000000-0000-0000-0000-0000000000c1"]
	require.Equal(t, services.PartStatusAIUnavailable, entry.Status)
	require.Equal(t, enums.AttemptTypeWriting, entry.Type)
}

func TestScoreWritingRetryOverwritesFallbackPart(t *testing.T) {
	// A retry after the fallback must update the same part to scored and
	// re-finalize; the AI paths prove ownership without an open-status
	// guard, so retrying a completed attempt is legal and is pinned here.
	activities := repomocks.NewMockLessonActivityRepository(t)
	activities.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-0000000000c1").Return(writingActivity(), nil).Once()
	activities.EXPECT().ListByLesson(mock.Anything, "00000000-0000-0000-0000-0000000000b1").
		Return([]*models.LessonActivity{writingActivity()}, nil).Once()
	attempts := repomocks.NewMockAttemptRepository(t)
	fallback := baseAttempt()
	fallback.Status = enums.AttemptStatusCompleted
	fallback.Result = []byte(`{"00000000-0000-0000-0000-0000000000c1":{"type":"writing","status":"ai_unavailable","score":0,"detail":null}}`)
	attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-0000000000a1").Return(fallback, nil).Once()
	saved := make([]*models.Attempt, 0)
	attempts.EXPECT().Update(mock.Anything, mock.Anything).Run(func(ctx context.Context, m *models.Attempt) {
		saved = append(saved, m)
	}).Return(nil).Once()

	svc := services.NewLessonActivityService(activities, fakeScorer{}, fakeLLM{out: retryFeedback()}, attempts,
		repomocks.NewMockLessonRepository(t), fakeNotPremium{})
	dto, err := svc.ScoreWriting(context.Background(), "user_1", "00000000-0000-0000-0000-0000000000c1", "00000000-0000-0000-0000-0000000000a1", "I think it is good because it is cheap")
	require.NoError(t, err)
	require.NotNil(t, dto)
	require.Equal(t, enums.AttemptStatusCompleted, saved[0].Status)
	var parts map[string]services.PartEntry
	require.NoError(t, json.Unmarshal(saved[0].Result, &parts))
	require.Equal(t, services.PartStatusScored, parts["00000000-0000-0000-0000-0000000000c1"].Status)
	assert.Greater(t, parts["00000000-0000-0000-0000-0000000000c1"].Score, 0.0)
}

func TestPronounceAttachesFallbackAndLimitsInFlight(t *testing.T) {
	speaking := &models.LessonActivity{
		ID:     "00000000-0000-0000-0000-0000000000c2",
		Type:   enums.ActivityTypeSpeaking,
		Config: []byte(`{"items":[{"text":"hello"}]}`),
	}
	activities := repomocks.NewMockLessonActivityRepository(t)
	activities.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-0000000000c2").Return(speaking, nil).Twice()
	activities.EXPECT().ListByLesson(mock.Anything, "00000000-0000-0000-0000-0000000000b1").
		Return([]*models.LessonActivity{speaking}, nil).Once()
	attempts := repomocks.NewMockAttemptRepository(t)
	attempts.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-0000000000a1").Return(baseAttempt(), nil).Twice()
	attempts.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()

	release := make(chan struct{})
	firstStarted := make(chan struct{})
	scorer := blockingScorer{started: firstStarted, release: release}
	svc := services.NewLessonActivityService(activities, scorer, fakeLLM{}, attempts,
		repomocks.NewMockLessonRepository(t), fakeNotPremium{})

	done := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := svc.Pronounce(context.Background(), "user_1", "00000000-0000-0000-0000-0000000000c2", "00000000-0000-0000-0000-0000000000a1", 0, []byte("audio"), "audio/wav")
		done <- err
	}()
	<-firstStarted
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := svc.Pronounce(context.Background(), "user_1", "00000000-0000-0000-0000-0000000000c2", "00000000-0000-0000-0000-0000000000a1", 0, []byte("audio2"), "audio/wav")
		done <- err
	}()
	select {
	case err := <-done:
		appErr := &common.AppError{}
		require.ErrorAs(t, err, &appErr)
		require.Equal(t, 429, appErr.Status, "the second in-flight pronunciation must be refused")
		require.Equal(t, common.CodeRateLimited, appErr.Code)
	case <-time.After(2 * time.Second):
		t.Fatal("second call never returned")
	}
	close(release)
	wg.Wait()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("first call never returned")
	}
}

type blockingScorer struct {
	started chan struct{}
	release chan struct{}
}

func (b blockingScorer) Score(ctx context.Context, text string, audio []byte, mime string) (*pronunciation.Result, error) {
	close(b.started) // may only be called once: the second call must never reach the scorer
	<-b.release
	return &pronunciation.Result{Score: 90}, nil
}
