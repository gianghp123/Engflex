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
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/profiles/dtos/requests"
	"engflex-api/internal/modules/profiles/services"
)

const profileUserID = "user_profile_1"

func requireAppError(t *testing.T, err error, status int) {
	t.Helper()
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, status, appErr.Status)
}

func TestCompleteOnboarding(t *testing.T) {
	completed := time.Date(2026, 5, 18, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		existing        *models.UserProfile
		loadErr         error
		upsertErr       error
		req             requests.UpdateProfile
		wantErrStatus   int
		wantPreferences models.Preferences
	}{
		{
			name:    "new profile completes onboarding",
			loadErr: gorm.ErrRecordNotFound,
			req: requests.UpdateProfile{
				Level:              enums.LevelIntermediate,
				Goals:              []enums.Goal{enums.GoalWork, enums.GoalExams},
				DailyCommitmentMin: 15,
				Topics:             []enums.InterestTopic{enums.InterestTopicBusiness, enums.InterestTopicNews},
			},
			wantPreferences: models.Preferences{
				Level:              enums.LevelIntermediate,
				Goals:              []enums.Goal{enums.GoalWork, enums.GoalExams},
				DailyCommitmentMin: 15,
				Topics:             []enums.InterestTopic{enums.InterestTopicBusiness, enums.InterestTopicNews},
			},
		},
		{
			name: "partial body merges into existing preferences",
			existing: &models.UserProfile{UserID: profileUserID, Preferences: models.Preferences{
				Level:              enums.LevelBeginner,
				Goals:              []enums.Goal{enums.GoalTravel},
				DailyCommitmentMin: 30,
				Topics:             []enums.InterestTopic{enums.InterestTopicFood, enums.InterestTopicMusic},
			}},
			req: requests.UpdateProfile{Goals: []enums.Goal{enums.GoalWork}},
			wantPreferences: models.Preferences{
				Level:              enums.LevelBeginner,
				Goals:              []enums.Goal{enums.GoalWork},
				DailyCommitmentMin: 30,
				Topics:             []enums.InterestTopic{enums.InterestTopicFood, enums.InterestTopicMusic},
			},
		},
		{
			name:            "already-completed timestamp is not reset",
			existing:        &models.UserProfile{UserID: profileUserID, OnboardingCompletedAt: &completed},
			req:             requests.UpdateProfile{Level: enums.LevelAdvanced},
			wantPreferences: models.Preferences{Level: enums.LevelAdvanced},
		},
		{
			name:          "upsert error maps to 500",
			loadErr:       gorm.ErrRecordNotFound,
			upsertErr:     errors.New("boom"),
			req:           requests.UpdateProfile{Level: enums.LevelBeginner},
			wantErrStatus: http.StatusInternalServerError,
		},
		{
			name:          "load error maps to 500",
			loadErr:       errors.New("boom"),
			wantErrStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomocks.NewMockUserProfileRepository(t)
			repo.On("GetByUserID", mock.Anything, profileUserID).Return(tt.existing, tt.loadErr)
			repo.On("Upsert", mock.Anything, mock.Anything).Return(tt.upsertErr).Maybe()

			svc := services.NewProfileService(repo)
			got, err := svc.CompleteOnboarding(context.Background(), profileUserID, tt.req)

			if tt.wantErrStatus != 0 {
				requireAppError(t, err, tt.wantErrStatus)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tt.wantPreferences, got.Preferences)
			require.NotNil(t, got.OnboardingCompletedAt)
			if tt.name == "already-completed timestamp is not reset" {
				assert.Equal(t, completed, *got.OnboardingCompletedAt)
			}
		})
	}
}

func TestGetProfileNotFound(t *testing.T) {
	repo := repomocks.NewMockUserProfileRepository(t)
	repo.On("GetByUserID", mock.Anything, profileUserID).Return(nil, gorm.ErrRecordNotFound)

	svc := services.NewProfileService(repo)
	got, err := svc.GetProfile(context.Background(), profileUserID)

	require.Nil(t, got)
	requireAppError(t, err, http.StatusNotFound)
}
