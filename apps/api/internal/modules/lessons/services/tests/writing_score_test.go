package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	llmmocks "engflex-api/internal/llm/mocks"
	"engflex-api/internal/modules/lessons/schemas"
	"engflex-api/internal/modules/lessons/services"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
)

const (
	writingUserID = "user_123"
	writingActID  = "77777777-7777-7777-7777-777777777777"
	writingOpenID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	writingText   = "Hello everyone, my name is Kenji and I is a sales assistant."
)

func writingActivityFixture() *models.LessonActivity {
	return &models.LessonActivity{
		ID:     writingActID,
		Type:   enums.ActivityTypeWriting,
		Config: datatypes.JSON(`{"task":"email","instructions":"i","stimulus":"s","minWords":10,"maxWords":100,"modelAnswer":"m","checklist":["greet the team","give name and role"]}`),
	}
}

func writingCanned() schemas.WritingFeedback {
	return schemas.WritingFeedback{
		Task: schemas.TaskCompletion{
			Verdicts: []schemas.TaskVerdict{
				{Item: "greet the team", Status: "covered", Evidence: "Hello everyone"},
				{Item: "give name and role", Status: "partial", Evidence: "my name is Kenji"},
			},
			AnswersPrompt: true,
			AnswersNote:   "Answers the prompt with a short self-introduction.",
		},
		Grammar: []schemas.Span{
			{Text: "I is", Occurrence: 1, Correction: "I am", Reason: "Be verb agrees with I as am"},
		},
		Phrasing: []schemas.Span{
			{Text: "sales assistant", Occurrence: 1, Correction: "sales associate", Reason: "More natural job title"},
		},
		Expressions: []string{"Hello everyone", "my name is"},
		Suggested:   "Hello everyone, my name is Kenji and I am a sales associate.",
		Tip:         "Remember the be verb: I am, not I is.",
	}
}

func writingOpenAttempt() *models.Attempt {
	lessonID := "11111111-1111-1111-1111-111111111111"
	return &models.Attempt{
		ID:       writingOpenID,
		UserID:   writingUserID,
		LessonID: &lessonID,
		Type:     enums.AttemptTypeLesson,
		Status:   enums.AttemptStatusInProgress,
		Result:   []byte(`{}`),
	}
}

func expectLLMCanned(llm *llmmocks.MockLLMClient, canned schemas.WritingFeedback) {
	llm.EXPECT().CompleteStructured(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Run(func(ctx context.Context, prompt string, output any, schemaName string, schema map[string]any) {
			*output.(*schemas.WritingFeedback) = canned
		}).Return(nil).Once()
}

func TestScoreWriting(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "happy path returns summary task spans and attempt echo",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				llm := llmmocks.NewMockLLMClient(t)
				repo.EXPECT().GetByID(mock.Anything, writingActID).Return(writingActivityFixture(), nil).Once()
				expectLLMCanned(llm, writingCanned())
				attempts.EXPECT().GetByID(mock.Anything, writingOpenID).Return(writingOpenAttempt(), nil).Twice()
				repo.EXPECT().ListByLesson(mock.Anything, mock.Anything).Return([]*models.LessonActivity{{ID: writingActID}}, nil).Once()
				var saved *models.Attempt
				attempts.EXPECT().Update(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, m *models.Attempt) { saved = m }).
					Return(nil).Once()
				svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llm, attempts)

				got, err := svc.ScoreWriting(context.Background(), writingUserID, writingActID, writingOpenID, writingText)

				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, 1, got.Summary.PointsCovered)
				assert.Equal(t, 2, got.Summary.PointsTotal)
				assert.Equal(t, 75.0, got.Summary.Score)
				assert.Equal(t, 12, got.Summary.WordCount)
				assert.Equal(t, 10, got.Summary.MinWords)
				assert.Equal(t, 100, got.Summary.MaxWords)
				require.Len(t, got.Task.Verdicts, 2)
				assert.Equal(t, "covered", got.Task.Verdicts[0].Status)
				assert.Equal(t, "partial", got.Task.Verdicts[1].Status)
				assert.True(t, got.Task.AnswersPrompt)
				require.Len(t, got.Grammar, 1)
				assert.Equal(t, "I is", got.Grammar[0].Text)
				assert.Equal(t, "I am", got.Grammar[0].Correction)
				require.Len(t, got.Phrasing, 1)
				assert.Equal(t, "sales assistant", got.Phrasing[0].Text)
				assert.Equal(t, []string{"Hello everyone", "my name is"}, got.Expressions)
				assert.Equal(t, "Hello everyone, my name is Kenji and I am a sales associate.", got.Suggested)
				assert.Equal(t, "Remember the be verb: I am, not I is.", got.Tip)
				assert.Equal(t, writingOpenID, got.Attempt.ID)
				assert.Equal(t, string(enums.AttemptStatusCompleted), got.Attempt.Status)
				require.NotNil(t, got.Attempt.Score)
				assert.Equal(t, 75.0, *got.Attempt.Score)
				require.NotNil(t, saved)
				assert.Equal(t, enums.AttemptStatusCompleted, saved.Status)
				require.NotNil(t, saved.Score)
				assert.Equal(t, 75.0, *saved.Score)
				var parts map[string]services.PartEntry
				require.NoError(t, json.Unmarshal(saved.Result, &parts))
				require.Contains(t, parts, writingActID)
				assert.Equal(t, 75.0, parts[writingActID].Score)
			},
		},
		{
			name: "llm error maps to 503 and writes nothing",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				llm := llmmocks.NewMockLLMClient(t)
				repo.EXPECT().GetByID(mock.Anything, writingActID).Return(writingActivityFixture(), nil).Once()
				attempts.EXPECT().GetByID(mock.Anything, writingOpenID).Return(writingOpenAttempt(), nil).Once()
				llm.EXPECT().CompleteStructured(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("boom")).Once()
				svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llm, attempts)

				got, err := svc.ScoreWriting(context.Background(), writingUserID, writingActID, writingOpenID, writingText)

				requireAppError(t, err, http.StatusServiceUnavailable)
				assert.Nil(t, got)
				attempts.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			},
		},
		{
			name: "span not substring maps to 503",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				llm := llmmocks.NewMockLLMClient(t)
				repo.EXPECT().GetByID(mock.Anything, writingActID).Return(writingActivityFixture(), nil).Once()
				attempts.EXPECT().GetByID(mock.Anything, writingOpenID).Return(writingOpenAttempt(), nil).Once()
				bad := writingCanned()
				bad.Grammar[0].Text = "never written words"
				expectLLMCanned(llm, bad)
				svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llm, attempts)

				got, err := svc.ScoreWriting(context.Background(), writingUserID, writingActID, writingOpenID, writingText)

				requireAppError(t, err, http.StatusServiceUnavailable)
				assert.Nil(t, got)
				attempts.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			},
		},
		{
			name: "verdict count mismatch maps to 503",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				llm := llmmocks.NewMockLLMClient(t)
				repo.EXPECT().GetByID(mock.Anything, writingActID).Return(writingActivityFixture(), nil).Once()
				attempts.EXPECT().GetByID(mock.Anything, writingOpenID).Return(writingOpenAttempt(), nil).Once()
				bad := writingCanned()
				bad.Task.Verdicts = bad.Task.Verdicts[:1]
				expectLLMCanned(llm, bad)
				svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llm, attempts)

				got, err := svc.ScoreWriting(context.Background(), writingUserID, writingActID, writingOpenID, writingText)

				requireAppError(t, err, http.StatusServiceUnavailable)
				assert.Nil(t, got)
				attempts.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			},
		},
		{
			name: "non-writing activity maps to 422",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				llm := llmmocks.NewMockLLMClient(t)
				reading := checkReadingFixture()
				repo.EXPECT().GetByID(mock.Anything, reading.ID).Return(reading, nil).Once()
				svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llm, attempts)

				got, err := svc.ScoreWriting(context.Background(), writingUserID, reading.ID, writingOpenID, writingText)

				requireAppError(t, err, http.StatusUnprocessableEntity)
				assert.Nil(t, got)
				llm.AssertNotCalled(t, "CompleteStructured", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				attempts.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			},
		},
		{
			name: "empty text maps to 400 without llm",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				llm := llmmocks.NewMockLLMClient(t)
				svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llm, attempts)

				got, err := svc.ScoreWriting(context.Background(), writingUserID, writingActID, writingOpenID, "   ")

				requireAppError(t, err, http.StatusBadRequest)
				assert.Nil(t, got)
				llm.AssertNotCalled(t, "CompleteStructured", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				attempts.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			},
		},
		{
			name: "over-cap text maps to 413 without llm",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				llm := llmmocks.NewMockLLMClient(t)
				svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llm, attempts)

				got, err := svc.ScoreWriting(context.Background(), writingUserID, writingActID, writingOpenID, strings.Repeat("word ", 501))

				requireAppError(t, err, http.StatusRequestEntityTooLarge)
				assert.Nil(t, got)
				llm.AssertNotCalled(t, "CompleteStructured", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				attempts.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			},
		},
		{
			name: "foreign attempt maps to 404",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				llm := llmmocks.NewMockLLMClient(t)
				repo.EXPECT().GetByID(mock.Anything, writingActID).Return(writingActivityFixture(), nil).Once()
				foreign := writingOpenAttempt()
				foreign.UserID = "user_999"
				attempts.EXPECT().GetByID(mock.Anything, writingOpenID).Return(foreign, nil).Once()
				svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llm, attempts)

				got, err := svc.ScoreWriting(context.Background(), writingUserID, writingActID, writingOpenID, writingText)

				requireAppError(t, err, http.StatusNotFound)
				assert.Nil(t, got)
				llm.AssertNotCalled(t, "CompleteStructured", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				attempts.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			},
		},
		{
			name: "missing activity maps to 404",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				llm := llmmocks.NewMockLLMClient(t)
				repo.EXPECT().GetByID(mock.Anything, writingActID).Return(nil, gorm.ErrRecordNotFound).Once()
				svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llm, attempts)

				got, err := svc.ScoreWriting(context.Background(), writingUserID, writingActID, writingOpenID, writingText)

				requireAppError(t, err, http.StatusNotFound)
				assert.Nil(t, got)
				llm.AssertNotCalled(t, "CompleteStructured", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
