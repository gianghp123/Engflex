package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/attempts/dtos/responses"
	"engflex-api/internal/modules/attempts/services"
)

const progressUserID = "user_progress_1"

// day returns the UTC calendar date (YYYY-MM-DD) `offset` days from today,
// matching the repository's to_char(..., 'YYYY-MM-DD') output.
func day(offset int) string {
	return time.Now().UTC().AddDate(0, 0, offset).Format("2006-01-02")
}

func requireAppError(t *testing.T, err error, status int) {
	t.Helper()
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, status, appErr.Status)
}

func TestGetProgress(t *testing.T) {
	profile := func(minutes int) *models.UserProfile {
		return &models.UserProfile{
			UserID:      progressUserID,
			Preferences: models.Preferences{DailyCommitmentMin: minutes},
		}
	}

	tests := []struct {
		name        string
		attemptDays []string
		endedDays   []string
		lessons     int64
		seconds     int64
		profile     *models.UserProfile
		profileErr  error
		daysErr     error
		want        *responses.Progress
		wantStatus  int
	}{
		{
			name:        "consecutive days including today",
			attemptDays: []string{day(0), day(-1), day(-2)},
			lessons:     5,
			seconds:     1800,
			profile:     profile(30),
			want:        &responses.Progress{StreakDays: 3, VoiceMinutesToday: 30, DailyGoalMinutes: 30, LessonsCompleted: 5},
		},
		{
			name:        "only yesterday keeps the streak alive",
			attemptDays: []string{day(-1)},
			profile:     profile(10),
			want:        &responses.Progress{StreakDays: 1, DailyGoalMinutes: 10},
		},
		{
			name:        "gap yesterday breaks the streak",
			attemptDays: []string{day(-2), day(-3)},
			profile:     profile(15),
			want:        &responses.Progress{StreakDays: 0, DailyGoalMinutes: 15},
		},
		{
			name:        "attempts and conversations union the days",
			attemptDays: []string{day(0), day(-2)},
			endedDays:   []string{day(-1)},
			profile:     profile(15),
			want:        &responses.Progress{StreakDays: 3, DailyGoalMinutes: 15},
		},
		{
			name:        "missing profile falls back to the default goal",
			attemptDays: []string{day(0)},
			profileErr:  gorm.ErrRecordNotFound,
			want:        &responses.Progress{StreakDays: 1, DailyGoalMinutes: 15},
		},
		{
			name:       "repo error maps to 500",
			daysErr:    errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attempts := repomocks.NewMockAttemptRepository(t)
			convs := repomocks.NewMockConversationRepository(t)
			profiles := repomocks.NewMockUserProfileRepository(t)
			// Maybe: a repo error short-circuits before the later calls, so
			// not every expectation fires on every case.
			attempts.On("ListCompletedDays", mock.Anything, progressUserID, mock.Anything).
				Return(tt.attemptDays, tt.daysErr).Maybe()
			convs.On("ListEndedDays", mock.Anything, progressUserID, mock.Anything).
				Return(tt.endedDays, nil).Maybe()
			attempts.On("CountCompleted", mock.Anything, progressUserID).
				Return(tt.lessons, nil).Maybe()
			convs.On("SumDurationSince", mock.Anything, progressUserID, mock.Anything).
				Return(tt.seconds, nil).Maybe()
			profiles.On("GetByUserID", mock.Anything, progressUserID).
				Return(tt.profile, tt.profileErr).Maybe()

			svc := services.NewProgressService(attempts, convs, profiles)
			got, err := svc.GetProgress(context.Background(), progressUserID)

			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.want, got)
		})
	}
}
