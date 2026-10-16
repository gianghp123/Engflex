package tests

import (
	"context"
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
	"engflex-api/internal/modules/lessons/dtos/responses"
	"engflex-api/internal/modules/lessons/services"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
)

func TestListProgress(t *testing.T) {
	const (
		userID = "user-1"
		l1     = "11111111-1111-1111-1111-111111111111"
		l2     = "22222222-2222-2222-2222-222222222222"
		a1     = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
		a2     = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
		a3     = "cccccccc-cccc-cccc-cccc-cccccccccccc"
	)
	attempt := func(id, lessonID string, status enums.AttemptStatus, result string) *models.Attempt {
		lesson := lessonID
		return &models.Attempt{
			ID: id, UserID: userID, LessonID: &lesson, Type: enums.AttemptTypeLesson,
			Status: status, Result: []byte(result),
		}
	}
	one := `{"` + a1 + `":{"type":"reading","score":100,"detail":{}}}`
	three := `{"` + a1 + `":{"type":"reading","score":100,"detail":{}},"` + a2 + `":{"type":"writing","score":80,"detail":{}},"` + a3 + `":{"type":"speaking","score":90,"detail":{}}}`

	tests := []struct {
		name       string
		setup      func(attempts *mocks.MockAttemptRepository)
		wantStatus int
		want       []*responses.LessonProgressSummary
	}{
		{
			name: "no attempts",
			setup: func(attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().ListLessonAttempts(mock.Anything, userID).
					Return([]*models.Attempt{}, nil).Once()
			},
			want: []*responses.LessonProgressSummary{},
		},
		{
			name: "open run with parts sorted by id",
			setup: func(attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().ListLessonAttempts(mock.Anything, userID).
					Return([]*models.Attempt{
						attempt("dddddddd-dddd-dddd-dddd-dddddddddddd", l1, enums.AttemptStatusInProgress, three),
					}, nil).Once()
			},
			want: []*responses.LessonProgressSummary{
				{LessonID: l1, Status: enums.AttemptStatusInProgress, CompletedActivityIDs: []string{a1, a2, a3}},
			},
		},
		{
			name: "completed beats newer open",
			setup: func(attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().ListLessonAttempts(mock.Anything, userID).
					Return([]*models.Attempt{
						attempt("dddddddd-dddd-dddd-dddd-dddddddddddd", l1, enums.AttemptStatusInProgress, one),
						attempt("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee", l1, enums.AttemptStatusCompleted, three),
					}, nil).Once()
			},
			want: []*responses.LessonProgressSummary{
				{LessonID: l1, Status: enums.AttemptStatusCompleted, CompletedActivityIDs: []string{a1, a2, a3}},
			},
		},
		{
			name: "newest completed wins",
			setup: func(attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().ListLessonAttempts(mock.Anything, userID).
					Return([]*models.Attempt{
						attempt("dddddddd-dddd-dddd-dddd-dddddddddddd", l1, enums.AttemptStatusCompleted, three),
						attempt("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee", l1, enums.AttemptStatusCompleted, one),
					}, nil).Once()
			},
			want: []*responses.LessonProgressSummary{
				{LessonID: l1, Status: enums.AttemptStatusCompleted, CompletedActivityIDs: []string{a1, a2, a3}},
			},
		},
		{
			name: "two lessons sorted by lesson id",
			setup: func(attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().ListLessonAttempts(mock.Anything, userID).
					Return([]*models.Attempt{
						attempt("dddddddd-dddd-dddd-dddd-dddddddddddd", l2, enums.AttemptStatusInProgress, one),
						attempt("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee", l1, enums.AttemptStatusInProgress, three),
					}, nil).Once()
			},
			want: []*responses.LessonProgressSummary{
				{LessonID: l1, Status: enums.AttemptStatusInProgress, CompletedActivityIDs: []string{a1, a2, a3}},
				{LessonID: l2, Status: enums.AttemptStatusInProgress, CompletedActivityIDs: []string{a1}},
			},
		},
		{
			name: "repo error",
			setup: func(attempts *mocks.MockAttemptRepository) {
				attempts.EXPECT().ListLessonAttempts(mock.Anything, userID).
					Return(nil, gorm.ErrInvalidDB).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockLessonActivityRepository(t)
			attempts := mocks.NewMockAttemptRepository(t)
			tt.setup(attempts)
			svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llmmocks.NewMockLLMClient(t), attempts)
			got, err := svc.ListProgress(context.Background(), userID)
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
