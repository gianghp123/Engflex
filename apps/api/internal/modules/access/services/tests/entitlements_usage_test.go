// GET /premium/usage assembly (spec §10): plan plus keyed counter views,
// the costs table, and the settled-charge history tail (committed +
// reserved, newest first, paginated). Action names derive from the charged
// entity; holds explain missing credits while released/expired rows never
// appear (the repo query pins the status filter, asserted here by the
// expected call args).
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

	"engflex-api/config"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/access/dtos/responses"
	accessservices "engflex-api/internal/modules/access/services"
)

func requireAppError(t *testing.T, err error, expectedStatus int) *common.AppError {
	t.Helper()
	require.Error(t, err)
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, expectedStatus, appErr.Status)
	return appErr
}

func usageCatalog() []*models.PlanFeatureLimit {
	return []*models.PlanFeatureLimit{
		{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 15, Period: enums.UsagePeriodMonth},
		{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 3, Period: enums.UsagePeriodMonth},
		{PlanKey: enums.PlanKeyTrial, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 30, Period: enums.UsagePeriodTotal},
		{PlanKey: enums.PlanKeyTrial, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 7, Period: enums.UsagePeriodTotal},
		{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 100, Period: enums.UsagePeriodMonth},
		{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 60, Period: enums.UsagePeriodMonth},
	}
}

func newUsageSvc(t *testing.T, subs accessservices.SubscriptionPlanSource, trials *repomocks.MockUserTrialRepository, counters *repomocks.MockUsageCounterRepository, limits *repomocks.MockPlanFeatureLimitRepository, events *repomocks.MockUsageEventRepository) *accessservices.EntitlementsService {
	t.Helper()
	costs := repomocks.NewMockActionCreditCostRepository(t)
	costs.EXPECT().List(mock.Anything).Return(defaultCostRows(), nil).Maybe()
	settings := repomocks.NewMockTrialSettingRepository(t)
	settings.EXPECT().Get(mock.Anything).Return(&models.TrialSetting{ID: true, DurationDays: 7}, nil).Maybe()
	return accessservices.NewEntitlementsService(
		accessservices.NewPolicyService(subs, trials, counters, limits, costs, settings, config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"}),
		events)
}

func freeUsageMocks(t *testing.T) (*repomocks.MockUserTrialRepository, *repomocks.MockUsageCounterRepository, *repomocks.MockPlanFeatureLimitRepository, *repomocks.MockUsageEventRepository) {
	t.Helper()
	trials := repomocks.NewMockUserTrialRepository(t)
	counters := repomocks.NewMockUsageCounterRepository(t)
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Maybe()
	trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Maybe()
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{}, nil).Maybe()
	return trials, counters, limits, events
}

func TestMyUsageHistoryRendersEventAction(t *testing.T) {
	trials, counters, limits, events := freeUsageMocks(t)
	created := time.Now().UTC().Truncate(time.Second)
	events.EXPECT().ListByUserStatus(mock.Anything, "user_1",
		[]enums.UsageEventStatus{enums.UsageEventStatusCommitted, enums.UsageEventStatusReserved},
		20, 0).Return([]*models.UsageEvent{
		{ID: "evt-1", UserID: "user_1", Action: enums.AccessActionExamReview, Amount: 2,
			Status: enums.UsageEventStatusCommitted, RefType: enums.UsageRefTypeExamAttempt, RefID: "ref-1", CreatedAt: created},
		{ID: "evt-2", UserID: "user_1", Action: enums.AccessActionVoiceStart, Amount: 1,
			Status: enums.UsageEventStatusReserved, RefType: enums.UsageRefTypeConversation, RefID: "ref-2", CreatedAt: created},
	}, nil).Once()
	events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(2), nil).Once()

	svc := newUsageSvc(t, fakeSubscriptions{}, trials, counters, limits, events)
	got, err := svc.MyUsage(context.Background(), "user_1", false, 1, 20)
	require.NoError(t, err)
	require.Len(t, got.History.Items, 2)
	assert.Equal(t, string(enums.AccessActionExamReview), got.History.Items[0].Action, "history renders the ledger action, never a re-derived name")
	assert.Equal(t, 2, got.History.Items[0].Credits, "the row is amount-misleading on purpose: old amount-sniffing would render study_plan")
	assert.Equal(t, string(enums.AccessActionVoiceStart), got.History.Items[1].Action)
	assert.Equal(t, enums.UsageEventStatusReserved, got.History.Items[1].Status)
}

func TestMyUsageHistoryPagination(t *testing.T) {
	tests := []struct {
		name           string
		page, pageSize int
		wantLimit      int
		wantOffset     int
		wantPage       int
		wantPageSize   int
	}{
		{name: "first page has no offset", page: 1, pageSize: 20, wantLimit: 20, wantOffset: 0, wantPage: 1, wantPageSize: 20},
		{name: "later pages offset by limit", page: 3, pageSize: 7, wantLimit: 7, wantOffset: 14, wantPage: 3, wantPageSize: 7},
		{name: "pageSize capped at 100", page: 1, pageSize: 500, wantLimit: 100, wantOffset: 0, wantPage: 1, wantPageSize: 100},
		{name: "page below 1 clamps to 1", page: 0, pageSize: 20, wantLimit: 20, wantOffset: 0, wantPage: 1, wantPageSize: 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trials, counters, limits, events := freeUsageMocks(t)
			events.EXPECT().ListByUserStatus(mock.Anything, "user_1",
				[]enums.UsageEventStatus{enums.UsageEventStatusCommitted, enums.UsageEventStatusReserved},
				tt.wantLimit, tt.wantOffset).Return([]*models.UsageEvent{}, nil).Once()
			events.EXPECT().CountByUserStatus(mock.Anything, "user_1",
				[]enums.UsageEventStatus{enums.UsageEventStatusCommitted, enums.UsageEventStatusReserved}).
				Return(int64(42), nil).Once()

			svc := newUsageSvc(t, fakeSubscriptions{}, trials, counters, limits, events)
			got, err := svc.MyUsage(context.Background(), "user_1", false, tt.page, tt.pageSize)
			require.NoError(t, err)
			assert.Equal(t, enums.PlanKeyFree, got.Plan)
			assert.Equal(t, tt.wantPage, got.History.Page)
			assert.Equal(t, tt.wantPageSize, got.History.PageSize)
			assert.Equal(t, int64(42), got.History.Total)
			assert.NotNil(t, got.History.Items, "history items serialize as [] never null")
			assert.NotEmpty(t, got.Costs, "costs table rides along on the usage view")
		})
	}
}

func TestMyUsageCountersMirrorMe(t *testing.T) {
	t.Run("trial uses the all period key and omits resetAt", func(t *testing.T) {
		trials := repomocks.NewMockUserTrialRepository(t)
		counters := repomocks.NewMockUsageCounterRepository(t)
		limits := repomocks.NewMockPlanFeatureLimitRepository(t)
		events := repomocks.NewMockUsageEventRepository(t)
		limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Once()
		trials.EXPECT().Get(mock.Anything, "user_1").Return(trialRow(3), nil).Once()
		counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{
			{UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "all", Used: 4},
		}, nil).Once()
		events.EXPECT().ListByUserStatus(mock.Anything, "user_1", mock.Anything, mock.Anything, mock.Anything).
			Return([]*models.UsageEvent{}, nil).Once()
		events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(0), nil).Once()

		svc := newUsageSvc(t, fakeSubscriptions{}, trials, counters, limits, events)
		got, err := svc.MyUsage(context.Background(), "user_1", false, 1, 20)
		require.NoError(t, err)
		assert.Equal(t, enums.PlanKeyTrial, got.Plan)
		require.Len(t, got.Counters, 2)
		assert.Equal(t, enums.CounterKindAICredit, got.Counters[0].Counter)
		assert.Equal(t, 30, got.Counters[0].Limit)
		assert.Equal(t, 4, got.Counters[0].Used)
		assert.Equal(t, 26, got.Counters[0].Remaining)
		assert.Nil(t, got.Counters[0].ResetAt, "total-period limits never reset")
		assert.Nil(t, got.Counters[1].ResetAt, "total-period limits never reset")
		assert.Equal(t, enums.CounterKindVoiceSession, got.Counters[1].Counter)
	})

	t.Run("monthly plan matches the live period key", func(t *testing.T) {
		policy := accessservices.NewPolicyService(fakeSubscriptions{sub: paidSub(enums.SubscriptionPlanOneMonth)},
			repomocks.NewMockUserTrialRepository(t), repomocks.NewMockUsageCounterRepository(t),
			repomocks.NewMockPlanFeatureLimitRepository(t),
			repomocks.NewMockActionCreditCostRepository(t), repomocks.NewMockTrialSettingRepository(t),
			config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
		monthKey := policy.PeriodKey(enums.UsagePeriodMonth, time.Now().UTC())

		trials := repomocks.NewMockUserTrialRepository(t)
		counters := repomocks.NewMockUsageCounterRepository(t)
		limits := repomocks.NewMockPlanFeatureLimitRepository(t)
		events := repomocks.NewMockUsageEventRepository(t)
		limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Once()
		counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{
			{UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: monthKey, Used: 12},
			{UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "2000-01", Used: 999},
			{UserID: "user_1", CounterKey: enums.CounterKindVoiceSession, PeriodKey: monthKey, Used: 3, Reserved: 1},
		}, nil).Once()
		events.EXPECT().ListByUserStatus(mock.Anything, "user_1", mock.Anything, mock.Anything, mock.Anything).
			Return([]*models.UsageEvent{}, nil).Once()
		events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(0), nil).Once()

		svc := newUsageSvc(t, fakeSubscriptions{sub: paidSub(enums.SubscriptionPlanOneMonth)}, trials, counters, limits, events)
		got, err := svc.MyUsage(context.Background(), "user_1", false, 1, 20)
		require.NoError(t, err)
		assert.Equal(t, enums.PlanKey1M, got.Plan)
		assert.Equal(t, 100, got.Counters[0].Limit)
		assert.Equal(t, 12, got.Counters[0].Used, "stale-period rows are ignored")
		assert.Equal(t, 88, got.Counters[0].Remaining)
		require.NotNil(t, got.Counters[0].ResetAt)
		assert.Equal(t, 60, got.Counters[1].Limit)
		assert.Equal(t, 3, got.Counters[1].Used)
		assert.Equal(t, 1, got.Counters[1].Reserved)
		assert.Equal(t, 56, got.Counters[1].Remaining)
		require.NotNil(t, got.Counters[1].ResetAt)
	})
}

func TestMyUsageAdminMirrorsMe(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Once()
	costs := repomocks.NewMockActionCreditCostRepository(t)
	costs.EXPECT().List(mock.Anything).Return(defaultCostRows(), nil).Once()
	settings := repomocks.NewMockTrialSettingRepository(t)
	settings.EXPECT().Get(mock.Anything).Return(&models.TrialSetting{ID: true, DurationDays: 7}, nil).Once()
	// No counter read: admins bypass metering with zero views, like /premium/me.
	events.EXPECT().ListByUserStatus(mock.Anything, "user_1", mock.Anything, mock.Anything, mock.Anything).
		Return([]*models.UsageEvent{}, nil).Once()
	events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(0), nil).Once()

	svc := accessservices.NewEntitlementsService(
		accessservices.NewPolicyService(fakeSubscriptions{}, repomocks.NewMockUserTrialRepository(t),
			repomocks.NewMockUsageCounterRepository(t), limits, costs, settings,
			config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"}),
		events)
	got, err := svc.MyUsage(context.Background(), "user_1", true, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, enums.PlanKeyAdmin, got.Plan)
	for _, c := range got.Counters {
		assert.Zero(t, c.Limit, "admin plan carries zero limits like /premium/me")
		assert.Zero(t, c.Used)
		assert.Zero(t, c.Reserved)
		assert.Zero(t, c.Remaining)
		assert.Nil(t, c.ResetAt)
	}
}

func TestMyUsageStoreFailures(t *testing.T) {
	tests := []struct {
		name    string
		counter error
		history error
		count   error
	}{
		{name: "counters down", counter: errors.New("db down")},
		{name: "history list down", history: errors.New("db down")},
		{name: "history count down", count: errors.New("db down")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trials := repomocks.NewMockUserTrialRepository(t)
			counters := repomocks.NewMockUsageCounterRepository(t)
			limits := repomocks.NewMockPlanFeatureLimitRepository(t)
			events := repomocks.NewMockUsageEventRepository(t)
			limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Once()
			trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Once()
			if tt.counter != nil {
				counters.EXPECT().ListByUser(mock.Anything, "user_1").Return(nil, tt.counter).Once()
			} else {
				counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{}, nil).Once()
			}
			if tt.counter == nil {
				if tt.history != nil {
					events.EXPECT().ListByUserStatus(mock.Anything, "user_1", mock.Anything, mock.Anything, mock.Anything).
						Return(nil, tt.history).Once()
				} else {
					events.EXPECT().ListByUserStatus(mock.Anything, "user_1", mock.Anything, mock.Anything, mock.Anything).
						Return([]*models.UsageEvent{}, nil).Once()
				}
			}
			if tt.counter == nil && tt.history == nil {
				if tt.count != nil {
					events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(0), tt.count).Once()
				} else {
					events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(0), nil).Once()
				}
			}

			svc := newUsageSvc(t, fakeSubscriptions{}, trials, counters, limits, events)
			_, err := svc.MyUsage(context.Background(), "user_1", false, 1, 20)
			appErr := requireAppError(t, err, http.StatusServiceUnavailable)
			assert.Equal(t, common.CodePolicyUnavailable, appErr.Code)
		})
	}
}

// TestUsageCostsComeFromPolicy pins the Task 3 consumer contract: both
// Costs tables render the policy catalog snapshot (zero included), never
// hardcoded literals.
func TestUsageCostsComeFromPolicy(t *testing.T) {
	trials := repomocks.NewMockUserTrialRepository(t)
	counters := repomocks.NewMockUsageCounterRepository(t)
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Maybe()
	costs := repomocks.NewMockActionCreditCostRepository(t)
	costs.EXPECT().List(mock.Anything).Return([]*models.ActionCreditCost{
		{Action: enums.AccessActionVoiceStart, Credits: 4},
		{Action: enums.AccessActionExamReview, Credits: 0},
		{Action: enums.AccessActionStudyPlan, Credits: 6},
	}, nil).Maybe()
	settings := repomocks.NewMockTrialSettingRepository(t)
	settings.EXPECT().Get(mock.Anything).Return(&models.TrialSetting{ID: true, DurationDays: 7}, nil).Maybe()
	trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Maybe()
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{}, nil).Maybe()
	events.EXPECT().ListByUserStatus(mock.Anything, "user_1", mock.Anything, mock.Anything, mock.Anything).
		Return([]*models.UsageEvent{}, nil).Maybe()
	events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(0), nil).Maybe()

	policy := accessservices.NewPolicyService(fakeSubscriptions{}, trials, counters, limits, costs, settings,
		config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
	svc := accessservices.NewEntitlementsService(policy, events)

	got, err := svc.MyUsage(context.Background(), "user_1", false, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, []responses.ActionCost{
		{Action: enums.AccessActionVoiceStart, Credits: 4},
		{Action: enums.AccessActionExamReview, Credits: 0},
		{Action: enums.AccessActionStudyPlan, Credits: 6},
	}, got.Costs)

	me, err := svc.MyEntitlements(context.Background(), "user_1", false)
	require.NoError(t, err)
	assert.Equal(t, got.Costs, me.Costs, "both Costs tables render the same snapshot")
}
