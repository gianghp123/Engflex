// /premium/me assembly (spec §10): plan, credit and voice views with
// remaining + resetAt, trial reported separately, costs table for the
// pricing/upgrade surfaces.
package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/config"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/access/dtos/responses"
	accessservices "engflex-api/internal/modules/access/services"
)

func TestMyEntitlementsShape(t *testing.T) {
	tests := []struct {
		name        string
		sub         *models.Subscription
		trial       *models.UserTrial
		counters    []*models.UsageCounter
		wantPremium bool
		wantPlan    enums.PlanKey
		wantCredits responses.CounterView
	}{
		{
			name: "premium user with usage",
			sub:  paidSub(enums.SubscriptionPlanOneMonth),
			counters: []*models.UsageCounter{
				{UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "2026-11", Used: 12},
				{UserID: "user_1", CounterKey: enums.CounterKindVoiceSession, PeriodKey: "2026-11", Used: 3},
			},
			wantPremium: true, wantPlan: enums.PlanKey1M,
			wantCredits: responses.CounterView{Limit: 100, Used: 12, Reserved: 0, Remaining: 88},
		},
		{
			name:        "free user with no usage",
			counters:    []*models.UsageCounter{},
			wantPlan:    enums.PlanKeyFree,
			wantCredits: responses.CounterView{Limit: 15, Used: 0, Reserved: 0, Remaining: 15},
		},
		{
			name:  "active trial reported separately",
			trial: &models.UserTrial{UserID: "user_1", StartedAt: time.Now(), EndsAt: time.Now().Add(6 * 24 * time.Hour), CreditsGranted: 30},
			counters: []*models.UsageCounter{
				{UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "all", Used: 4},
			},
			wantPlan:    enums.PlanKeyTrial,
			wantCredits: responses.CounterView{Limit: 30, Used: 4, Reserved: 0, Remaining: 26},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trials := repomocks.NewMockUserTrialRepository(t)
			counters := repomocks.NewMockUsageCounterRepository(t)
			limits := repomocks.NewMockPlanFeatureLimitRepository(t)
			limits.EXPECT().ListAll(mock.Anything).Return([]*models.PlanFeatureLimit{
				{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 15, Period: enums.UsagePeriodMonth},
				{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 3, Period: enums.UsagePeriodMonth},
				{PlanKey: enums.PlanKeyTrial, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 30, Period: enums.UsagePeriodTotal},
				{PlanKey: enums.PlanKeyTrial, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 7, Period: enums.UsagePeriodTotal},
				{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 100, Period: enums.UsagePeriodMonth},
				{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 60, Period: enums.UsagePeriodMonth},
			}, nil).Once()
			if tt.trial != nil {
				trials.EXPECT().Get(mock.Anything, "user_1").Return(tt.trial, nil).Once()
			} else if tt.sub == nil {
				trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Once()
			}
			counters.EXPECT().ListByUser(mock.Anything, "user_1").Return(tt.counters, nil).Once()

			svc := accessservices.NewEntitlementsService(
				newPolicyFor(t, fakeSubscriptions{sub: tt.sub}, trials, counters, limits))

			got, err := svc.MyEntitlements(context.Background(), "user_1", false)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPremium, got.IsPremium)
			assert.Equal(t, tt.wantPlan, got.Plan)
			// ResetAt is clock-derived, so it is asserted separately from
			// the count fields: present for rolling periods, absent for the
			// trial's never-resetting total period.
			assert.Equal(t, tt.wantCredits.Limit, got.Credits.Limit)
			assert.Equal(t, tt.wantCredits.Used, got.Credits.Used)
			assert.Equal(t, tt.wantCredits.Reserved, got.Credits.Reserved)
			assert.Equal(t, tt.wantCredits.Remaining, got.Credits.Remaining)
			if tt.wantPlan == enums.PlanKeyTrial {
				assert.Nil(t, got.Credits.ResetAt, "total-period limits never reset")
				assert.Nil(t, got.Voice.ResetAt, "total-period limits never reset")
				require.NotNil(t, got.Trial, "active trial is reported separately")
				assert.True(t, got.Trial.Active)
			} else {
				require.NotNil(t, got.Credits.ResetAt, "resetAt must be present for period limits")
				require.NotNil(t, got.Voice.ResetAt, "resetAt must be present for period limits")
				assert.Nil(t, got.Trial)
			}
			assert.NotEmpty(t, got.Costs, "costs table for the pricing surfaces")
		})
	}
}

func newPolicyFor(t *testing.T, subs accessservices.SubscriptionPlanSource, trials *repomocks.MockUserTrialRepository, counters *repomocks.MockUsageCounterRepository, limits *repomocks.MockPlanFeatureLimitRepository) *accessservices.PolicyService {
	t.Helper()
	return accessservices.NewPolicyService(subs, trials, counters, limits, config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
}
