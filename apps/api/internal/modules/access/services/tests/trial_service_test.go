// Trial contract (spec §8.4): one row per user (idempotent start), 7 days
// from now, granted amounts recorded, admin grant uses the same single
// path, and a used/expired trial can never restart.
package tests

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	accessservices "engflex-api/internal/modules/access/services"
)

func TestStartTrialOncePerUser(t *testing.T) {
	trials := repomocks.NewMockUserTrialRepository(t)
	trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Once()
	trials.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, m *models.UserTrial) error {
			assert.Equal(t, common.DefaultTrialCredits, m.CreditsGranted)
			assert.WithinDuration(t, time.Now().AddDate(0, 0, common.DefaultTrialDays), m.EndsAt, time.Minute)
			return nil
		}).Once()
	svc := accessservices.NewTrialService(trials, staticTrialTerms{accessservices.TrialTerms{Days: common.DefaultTrialDays, Credits: common.DefaultTrialCredits, VoiceSessions: common.DefaultTrialVoiceSessions}})
	started, err := svc.Start(context.Background(), "user_1")
	require.NoError(t, err)
	assert.True(t, started)
}

func TestStartTrialIdempotentWhenAlreadyUsed(t *testing.T) {
	trials := repomocks.NewMockUserTrialRepository(t)
	used := &models.UserTrial{UserID: "user_1", StartedAt: time.Now().AddDate(0, 0, -10), EndsAt: time.Now().AddDate(0, 0, -3), CreditsGranted: common.DefaultTrialCredits}
	trials.EXPECT().Get(mock.Anything, "user_1").Return(used, nil).Once()
	svc := accessservices.NewTrialService(trials, staticTrialTerms{accessservices.TrialTerms{Days: common.DefaultTrialDays, Credits: common.DefaultTrialCredits, VoiceSessions: common.DefaultTrialVoiceSessions}})
	started, err := svc.Start(context.Background(), "user_1")
	require.NoError(t, err)
	assert.False(t, started, "an expired trial never restarts")
}

func TestStartTrialActiveReportsAlreadyStarted(t *testing.T) {
	trials := repomocks.NewMockUserTrialRepository(t)
	active := &models.UserTrial{UserID: "user_1", StartedAt: time.Now(), EndsAt: time.Now().Add(24 * time.Hour), CreditsGranted: common.DefaultTrialCredits}
	trials.EXPECT().Get(mock.Anything, "user_1").Return(active, nil).Once()
	svc := accessservices.NewTrialService(trials, staticTrialTerms{accessservices.TrialTerms{Days: common.DefaultTrialDays, Credits: common.DefaultTrialCredits, VoiceSessions: common.DefaultTrialVoiceSessions}})
	started, err := svc.Start(context.Background(), "user_1")
	require.NoError(t, err)
	assert.True(t, started, "an active trial is already started — idempotent success")
}

func TestStartTrialRaceOnCreateReportsStarted(t *testing.T) {
	trials := repomocks.NewMockUserTrialRepository(t)
	trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Once()
	trials.EXPECT().Create(mock.Anything, mock.Anything).
		Return(&pgconn.PgError{Code: "23505", Message: `duplicate key value violates unique constraint "user_trials_pkey"`}).Once()
	svc := accessservices.NewTrialService(trials, staticTrialTerms{accessservices.TrialTerms{Days: common.DefaultTrialDays, Credits: common.DefaultTrialCredits, VoiceSessions: common.DefaultTrialVoiceSessions}})
	started, err := svc.Start(context.Background(), "user_1")
	require.NoError(t, err)
	assert.True(t, started, "a concurrent start won the race — idempotent success")
}

func TestStartTrialLookupFailure(t *testing.T) {
	trials := repomocks.NewMockUserTrialRepository(t)
	trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrInvalidDB).Once()
	svc := accessservices.NewTrialService(trials, staticTrialTerms{accessservices.TrialTerms{Days: common.DefaultTrialDays, Credits: common.DefaultTrialCredits, VoiceSessions: common.DefaultTrialVoiceSessions}})
	_, err := svc.Start(context.Background(), "user_1")
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 500, appErr.Status)
}

type staticTrialTerms struct{ terms accessservices.TrialTerms }

func (s staticTrialTerms) TrialTerms() accessservices.TrialTerms { return s.terms }

func TestStartTrialUsesSnapshotTerms(t *testing.T) {
	trials := repomocks.NewMockUserTrialRepository(t)
	trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Once()
	trials.EXPECT().Create(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, m *models.UserTrial) error {
			assert.Equal(t, 40, m.CreditsGranted)
			assert.WithinDuration(t, time.Now().AddDate(0, 0, 10), m.EndsAt, time.Minute)
			return nil
		}).Once()
	svc := accessservices.NewTrialService(trials, staticTrialTerms{accessservices.TrialTerms{Days: 10, Credits: 40, VoiceSessions: 10}})
	started, err := svc.Start(context.Background(), "user_1")
	require.NoError(t, err)
	assert.True(t, started)
}
