package tests

import (
	"context"
	"net/http"
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
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
)

func checkReadingFixture() *models.LessonActivity {
	return &models.LessonActivity{
		ID:     "33333333-3333-3333-3333-333333333333",
		Type:   enums.ActivityTypeReading,
		Config: datatypes.JSON(`{"text":"t","questions":[{"stem":"s","instruction":"i","options":[],"answerKey":"B","explanation":"why"}]}`),
	}
}

func TestCheckAnswer(t *testing.T) {
	reading := checkReadingFixture()
	writing := &models.LessonActivity{
		ID:     "44444444-4444-4444-4444-444444444444",
		Type:   enums.ActivityTypeWriting,
		Config: datatypes.JSON(`{"task":"email","instructions":"i","stimulus":"s","minWords":10,"maxWords":100,"modelAnswer":"m","checklist":[]}`),
	}
	broken := &models.LessonActivity{
		ID:     "55555555-5555-5555-5555-555555555555",
		Type:   enums.ActivityTypeReading,
		Config: datatypes.JSON(`{broken`),
	}

	tests := []struct {
		name          string
		activity      *models.LessonActivity
		repoErr       error
		questionIndex int
		key           string
		wantStatus    int
		wantCorrect   *bool
		wantKey       string
		wantExpl      string
	}{
		{name: "correct exact", activity: reading, questionIndex: 0, key: "B", wantCorrect: boolPtr(true), wantKey: "B", wantExpl: "why"},
		{name: "correct case and whitespace insensitive", activity: reading, questionIndex: 0, key: " b ", wantCorrect: boolPtr(true), wantKey: "B", wantExpl: "why"},
		{name: "incorrect", activity: reading, questionIndex: 0, key: "A", wantCorrect: boolPtr(false), wantKey: "B", wantExpl: "why"},
		{name: "index negative", activity: reading, questionIndex: -1, key: "B", wantStatus: http.StatusBadRequest},
		{name: "index out of range", activity: reading, questionIndex: 1, key: "B", wantStatus: http.StatusBadRequest},
		{name: "writing not checkable", activity: writing, questionIndex: 0, key: "B", wantStatus: http.StatusUnprocessableEntity},
		{name: "not found", activity: nil, repoErr: gorm.ErrRecordNotFound, questionIndex: 0, key: "B", wantStatus: http.StatusNotFound},
		{name: "broken config", activity: broken, questionIndex: 0, key: "B", wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockLessonActivityRepository(t)
			id := "33333333-3333-3333-3333-333333333333"
			if tt.activity != nil {
				id = tt.activity.ID
			}
			if tt.name == "not found" {
				id = "99999999-9999-9999-9999-999999999999"
			}
			repo.EXPECT().GetByID(mock.Anything, id).Return(tt.activity, tt.repoErr).Once()
			attempts := mocks.NewMockAttemptRepository(t)
			const userID = "user-1"
			const attemptID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
			if tt.wantStatus == 0 {
				lessonID := "11111111-1111-1111-1111-111111111111"
				attempts.EXPECT().GetForUpdate(mock.Anything, attemptID).Return(&models.Attempt{
					ID: attemptID, UserID: userID, LessonID: &lessonID,
					Type: enums.AttemptTypeLesson, Status: enums.AttemptStatusInProgress,
					Result: []byte(`{}`),
				}, nil).Once()
				repo.EXPECT().ListByLesson(mock.Anything, lessonID).Return([]*models.LessonActivity{{ID: id}}, nil).Once()
				attempts.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			}
			svc := newActivityService(t, repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llmmocks.NewMockLLMClient(t), attempts)
			got, err := svc.CheckAnswer(context.Background(), userID, id, attemptID, tt.questionIndex, tt.key)
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			require.Equal(t, *tt.wantCorrect, got.Correct)
			assert.Equal(t, tt.wantKey, got.CorrectKey)
			assert.Equal(t, tt.wantExpl, got.Explanation)
		})
	}
}

func boolPtr(b bool) *bool { return &b }
