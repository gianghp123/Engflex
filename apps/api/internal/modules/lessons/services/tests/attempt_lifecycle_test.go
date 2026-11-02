package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	llmmocks "engflex-api/internal/llm/mocks"
	"engflex-api/internal/modules/lessons/services"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
)

func TestAttemptLifecycle(t *testing.T) {
	const (
		userID   = "user_123"
		otherID  = "user_999"
		lessonID = "11111111-1111-1111-1111-111111111111"
		act1     = "33333333-3333-3333-3333-333333333333"
		act2     = "44444444-4444-4444-4444-444444444444"
		openID   = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	)

	newSvc := func(t *testing.T, repo *mocks.MockLessonActivityRepository, attempts *mocks.MockAttemptRepository) services.LessonActivityService {
		t.Helper()
		return newActivityService(t, repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llmmocks.NewMockLLMClient(t), attempts)
	}

	openAttempt := func() *models.Attempt {
		return &models.Attempt{
			ID:       openID,
			UserID:   userID,
			LessonID: &[]string{lessonID}[0],
			Type:     enums.AttemptTypeLesson,
			Status:   enums.AttemptStatusInProgress,
			Result:   []byte(`{}`),
		}
	}
	twoActivities := []*models.LessonActivity{{ID: act1, Type: enums.ActivityTypeWriting}, {ID: act2, Type: enums.ActivityTypeSpeaking, Config: []byte(`{"items":[{"text":"Hello","modelAudioKey":"hello.mp3"}]}`)}}

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "start ok creates in_progress lesson attempt",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				attempts.EXPECT().GetOpenAttempt(mock.Anything, userID, lessonID).Return(nil, gorm.ErrRecordNotFound).Once()
				var created *models.Attempt
				attempts.EXPECT().CreateOpen(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, m *models.Attempt) { created = m }).
					Return(nil).Once()
				svc := newSvc(t, repo, attempts)

				got, err := svc.StartAttempt(context.Background(), userID, lessonID)

				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, userID, got.UserID)
				require.NotNil(t, got.LessonID)
				assert.Equal(t, lessonID, *got.LessonID)
				assert.Equal(t, enums.AttemptTypeLesson, got.Type)
				assert.Equal(t, enums.AttemptStatusInProgress, got.Status)
				assert.JSONEq(t, `{}`, string(got.Result))
				require.NotNil(t, created)
				assert.Same(t, created, got)
			},
		},
		{
			name: "start bad lesson uuid",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				svc := newSvc(t, repo, attempts)

				got, err := svc.StartAttempt(context.Background(), userID, "not-a-uuid")

				requireAppError(t, err, http.StatusBadRequest)
				assert.Nil(t, got)
				attempts.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			},
		},
		{
			name: "start returns existing open attempt",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				open := openAttempt()
				attempts.EXPECT().GetOpenAttempt(mock.Anything, userID, lessonID).Return(open, nil).Once()
				svc := newSvc(t, repo, attempts)

				got, err := svc.StartAttempt(context.Background(), userID, lessonID)

				require.NoError(t, err)
				assert.Same(t, open, got)
				attempts.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			},
		},
		{
			name: "attach missing attempt",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				attempts.EXPECT().GetForUpdate(mock.Anything, openID).Return(nil, gorm.ErrRecordNotFound).Once()
				svc := newSvc(t, repo, attempts)

				got, err := svc.AttachPartResult(context.Background(), userID, openID, act1, services.PartEntry{Type: enums.AttemptTypeWriting, Score: 80})

				requireAppError(t, err, http.StatusNotFound)
				assert.Nil(t, got)
			},
		},
		{
			name: "attach malformed attempt id",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				svc := newSvc(t, repo, attempts)

				got, err := svc.AttachPartResult(context.Background(), userID, "nope", act1, services.PartEntry{Type: enums.AttemptTypeWriting, Score: 80})

				requireAppError(t, err, http.StatusBadRequest)
				assert.Nil(t, got)
				attempts.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
			},
		},
		{
			name: "attach foreign attempt",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				foreign := openAttempt()
				foreign.UserID = otherID
				attempts.EXPECT().GetForUpdate(mock.Anything, openID).Return(foreign, nil).Once()
				svc := newSvc(t, repo, attempts)

				got, err := svc.AttachPartResult(context.Background(), userID, openID, act1, services.PartEntry{Type: enums.AttemptTypeWriting, Score: 80})

				requireAppError(t, err, http.StatusNotFound)
				assert.Nil(t, got)
			},
		},
		{
			name: "attach closed attempt",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				closed := openAttempt()
				closed.Status = enums.AttemptStatusCompleted
				attempts.EXPECT().GetForUpdate(mock.Anything, openID).Return(closed, nil).Once()
				svc := newSvc(t, repo, attempts)

				got, err := svc.AttachPartResult(context.Background(), userID, openID, act1, services.PartEntry{Type: enums.AttemptTypeWriting, Score: 80})

				requireAppError(t, err, http.StatusConflict)
				assert.Nil(t, got)
			},
		},
		{
			name: "attach merges and stays open",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				attempts.EXPECT().GetForUpdate(mock.Anything, openID).Return(openAttempt(), nil).Once()
				repo.EXPECT().ListByLesson(mock.Anything, lessonID).Return(twoActivities, nil).Once()
				var saved *models.Attempt
				attempts.EXPECT().Update(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, m *models.Attempt) { saved = m }).
					Return(nil).Once()
				svc := newSvc(t, repo, attempts)

				got, err := svc.AttachPartResult(context.Background(), userID, openID, act1, services.PartEntry{Type: enums.AttemptTypeWriting, Score: 80, Detail: map[string]string{"suggested": "Improved answer", "tip": "x"}})

				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, enums.AttemptStatusInProgress, got.Status)
				assert.Nil(t, got.Score)
				require.NotNil(t, saved)
				var parts map[string]services.PartEntry
				require.NoError(t, json.Unmarshal(saved.Result, &parts))
				require.Contains(t, parts, act1)
				assert.Equal(t, 80.0, parts[act1].Score)
				assert.NotContains(t, parts, act2)
			},
		},
		{
			name: "attach completes with mean score",
			run: func(t *testing.T) {
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				seeded, err := json.Marshal(map[string]services.PartEntry{
					act1: {Type: enums.AttemptTypeWriting, Score: 80, Detail: map[string]string{"suggested": "Improved answer", "tip": "x"}},
				})
				require.NoError(t, err)
				open := openAttempt()
				open.Result = seeded
				attempts.EXPECT().GetForUpdate(mock.Anything, openID).Return(open, nil).Once()
				repo.EXPECT().ListByLesson(mock.Anything, lessonID).Return(twoActivities, nil).Once()
				var saved *models.Attempt
				attempts.EXPECT().Update(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, m *models.Attempt) { saved = m }).
					Return(nil).Once()
				svc := newSvc(t, repo, attempts)

				got, err := svc.AttachPartResult(context.Background(), userID, openID, act2, services.PartEntry{Type: enums.AttemptTypeSpeaking, Score: 60, Detail: map[string]any{"items": map[string]float64{"0": 60}}})

				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, enums.AttemptStatusCompleted, got.Status)
				require.NotNil(t, got.Score)
				assert.Equal(t, 70.0, *got.Score)
				require.NotNil(t, saved)
				var parts map[string]services.PartEntry
				require.NoError(t, json.Unmarshal(saved.Result, &parts))
				assert.Len(t, parts, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
