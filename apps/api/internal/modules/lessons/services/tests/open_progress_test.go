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
	"engflex-api/internal/modules/lessons/dtos/responses"
	"engflex-api/internal/modules/lessons/services"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
)

func TestGetOpenProgress(t *testing.T) {
	const (
		userID   = "user-1"
		lessonID = "11111111-1111-1111-1111-111111111111"
		openID   = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
		readID   = "33333333-3333-3333-3333-333333333333"
		writeID  = "44444444-4444-4444-4444-444444444444"
		goneID   = "55555555-5555-5555-5555-555555555555"
	)
	reading := &models.LessonActivity{
		ID:     readID,
		Type:   enums.ActivityTypeReading,
		Config: datatypes.JSON(`{"text":"t","questions":[{"stem":"s0","instruction":"i","options":[],"answerKey":"A","explanation":"why0"},{"stem":"s1","instruction":"i","options":[],"answerKey":"B","explanation":"why1"}]}`),
	}
	writing := &models.LessonActivity{
		ID:     writeID,
		Type:   enums.ActivityTypeWriting,
		Config: datatypes.JSON(`{"task":"email","instructions":"i","stimulus":"s","minWords":10,"maxWords":100,"modelAnswer":"m","checklist":[]}`),
	}
	openWith := func(result string) *models.Attempt {
		return &models.Attempt{
			ID: openID, UserID: userID, LessonID: &[]string{lessonID}[0],
			Type: enums.AttemptTypeLesson, Status: enums.AttemptStatusInProgress,
			Result: []byte(result),
		}
	}
	twoOutcomes := `{"` + readID + `":{"type":"reading","score":50,"detail":{"outcomes":[{"index":0,"key":"A","correct":true},{"index":1,"key":"A","correct":false}]}}}`
	newSvc := func(repo *mocks.MockLessonActivityRepository, attempts *mocks.MockAttemptRepository) services.LessonActivityService {
		return services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llmmocks.NewMockLLMClient(t), attempts)
	}

	tests := []struct {
		name       string
		lessonID   string
		setup      func(repo *mocks.MockLessonActivityRepository, attempts *mocks.MockAttemptRepository)
		wantStatus int
		want       *responses.AttemptProgress
	}{
		{
			name:       "invalid lesson id",
			lessonID:   "not-a-uuid",
			setup:      func(repo *mocks.MockLessonActivityRepository, attempts *mocks.MockAttemptRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:     "no open attempt",
			lessonID: lessonID,
			setup: func(repo *mocks.MockLessonActivityRepository, attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().GetOpenAttempt(mock.Anything, userID, lessonID).Return(nil, gorm.ErrRecordNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "repo error",
			lessonID: lessonID,
			setup: func(repo *mocks.MockLessonActivityRepository, attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().GetOpenAttempt(mock.Anything, userID, lessonID).Return(nil, gorm.ErrInvalidDB).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:     "enriched outcomes with attempt summary",
			lessonID: lessonID,
			setup: func(repo *mocks.MockLessonActivityRepository, attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().GetOpenAttempt(mock.Anything, userID, lessonID).Return(openWith(twoOutcomes), nil).Once()
				repo.EXPECT().GetByID(mock.Anything, readID).Return(reading, nil).Once()
			},
			want: &responses.AttemptProgress{
				Attempt: responses.AttemptSummary{ID: openID, Status: string(enums.AttemptStatusInProgress)},
				Checks: map[string][]responses.CheckedQuestion{
					readID: {
						{Index: 0, Key: "A", Correct: true, CorrectKey: "A", Explanation: "why0"},
						{Index: 1, Key: "A", Correct: false, CorrectKey: "B", Explanation: "why1"},
					},
				},
			},
		},
		{
			name:     "non-checkable entry omitted",
			lessonID: lessonID,
			setup: func(repo *mocks.MockLessonActivityRepository, attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().GetOpenAttempt(mock.Anything, userID, lessonID).Return(openWith(`{"`+writeID+`":{"type":"writing","score":80,"detail":{}}}`), nil).Once()
				repo.EXPECT().GetByID(mock.Anything, writeID).Return(writing, nil).Once()
			},
			want: &responses.AttemptProgress{
				Attempt: responses.AttemptSummary{ID: openID, Status: string(enums.AttemptStatusInProgress)},
				Checks:  map[string][]responses.CheckedQuestion{},
			},
		},
		{
			name:     "stale index skipped",
			lessonID: lessonID,
			setup: func(repo *mocks.MockLessonActivityRepository, attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().GetOpenAttempt(mock.Anything, userID, lessonID).Return(openWith(`{"`+readID+`":{"type":"reading","score":100,"detail":{"outcomes":[{"index":0,"key":"A","correct":true},{"index":9,"key":"C","correct":false}]}}}`), nil).Once()
				repo.EXPECT().GetByID(mock.Anything, readID).Return(reading, nil).Once()
			},
			want: &responses.AttemptProgress{
				Attempt: responses.AttemptSummary{ID: openID, Status: string(enums.AttemptStatusInProgress)},
				Checks: map[string][]responses.CheckedQuestion{
					readID: {
						{Index: 0, Key: "A", Correct: true, CorrectKey: "A", Explanation: "why0"},
					},
				},
			},
		},
		{
			name:     "missing activity skipped",
			lessonID: lessonID,
			setup: func(repo *mocks.MockLessonActivityRepository, attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().GetOpenAttempt(mock.Anything, userID, lessonID).Return(openWith(`{"`+readID+`":{"type":"reading","score":100,"detail":{"outcomes":[{"index":1,"key":"B","correct":true}]}},"`+goneID+`":{"type":"reading","score":0,"detail":{"outcomes":[{"index":0,"key":"A","correct":false}]}}}`), nil).Once()
				repo.EXPECT().GetByID(mock.Anything, readID).Return(reading, nil).Once()
				repo.EXPECT().GetByID(mock.Anything, goneID).Return(nil, gorm.ErrRecordNotFound).Once()
			},
			want: &responses.AttemptProgress{
				Attempt: responses.AttemptSummary{ID: openID, Status: string(enums.AttemptStatusInProgress)},
				Checks: map[string][]responses.CheckedQuestion{
					readID: {
						{Index: 1, Key: "B", Correct: true, CorrectKey: "B", Explanation: "why1"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockLessonActivityRepository(t)
			attempts := mocks.NewMockAttemptRepository(t)
			tt.setup(repo, attempts)
			svc := newSvc(repo, attempts)
			got, err := svc.GetOpenProgress(context.Background(), userID, tt.lessonID)
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.want.Attempt.ID, got.Attempt.ID)
			assert.Equal(t, tt.want.Attempt.Status, got.Attempt.Status)
			assert.Equal(t, tt.want.Checks, got.Checks)
		})
	}
}
