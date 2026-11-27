// Public pricing catalog composition (spec §10): free limits, trial terms,
// per-plan monthly limits, and the per-action cost table render from the
// policy snapshot plus the active plans list. A plan read failure is a 503
// policy_unavailable (paid surfaces fail closed).
package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/config"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	accessservices "engflex-api/internal/modules/access/services"
)

func TestPricingCatalogComposition(t *testing.T) {
	ctx := context.Background()
	plans := repomocks.NewMockPlanRepository(t)
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	costs := repomocks.NewMockActionCreditCostRepository(t)
	settings := repomocks.NewMockTrialSettingRepository(t)
	trials := repomocks.NewMockUserTrialRepository(t)
	counters := repomocks.NewMockUsageCounterRepository(t)

	plans.EXPECT().ListActive(mock.Anything).Return([]*models.Plan{
		{Plan: enums.SubscriptionPlanOneMonth, Months: 1, Amount: 99000, Currency: "VND", IsActive: true},
	}, nil).Once()
	limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Once()
	costs.EXPECT().List(mock.Anything).Return([]*models.ActionCreditCost{
		{Action: enums.AccessActionVoiceStart, Credits: 4},
		{Action: enums.AccessActionExamReview, Credits: 1},
		{Action: enums.AccessActionStudyPlan, Credits: 2},
	}, nil).Once()
	settings.EXPECT().Get(mock.Anything).Return(&models.TrialSetting{ID: true, DurationDays: 9}, nil).Once()

	policy := accessservices.NewPolicyService(fakeSubscriptions{}, trials, counters, limits, costs, settings,
		config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
	require.NoError(t, policy.SnapshotCatalog(ctx))
	svc := accessservices.NewPricingService(plans, policy)
	out, err := svc.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, 15, out.Free.Credits)
	assert.Equal(t, 3, out.Free.VoiceSessions)
	assert.Equal(t, 9, out.Trial.Days)
	assert.Equal(t, 30, out.Trial.Credits)
	require.Len(t, out.Plans, 1)
	assert.Equal(t, 100, out.Plans[0].Credits)
	assert.Equal(t, 60, out.Plans[0].VoiceSessions)
	require.Len(t, out.Costs, 3)
	assert.Equal(t, 4, out.Costs[0].Credits)
}

func TestPricingCatalogPlanReadFailureIs503(t *testing.T) {
	plans := repomocks.NewMockPlanRepository(t)
	plans.EXPECT().ListActive(mock.Anything).Return(nil, errors.New("db down")).Once()
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	costs := repomocks.NewMockActionCreditCostRepository(t)
	settings := repomocks.NewMockTrialSettingRepository(t)
	trials := repomocks.NewMockUserTrialRepository(t)
	counters := repomocks.NewMockUsageCounterRepository(t)
	policy := accessservices.NewPolicyService(fakeSubscriptions{}, trials, counters, limits, costs, settings,
		config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
	svc := accessservices.NewPricingService(plans, policy)
	out, err := svc.Get(context.Background())
	require.Error(t, err)
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, http.StatusServiceUnavailable, appErr.Status)
	assert.Equal(t, common.CodePolicyUnavailable, appErr.Code)
	assert.Nil(t, out)
}
