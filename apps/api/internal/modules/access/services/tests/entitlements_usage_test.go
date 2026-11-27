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
	return accessservices.NewEntitlementsService(
		accessservices.NewPolicyService(subs, trials, counters, limits, config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"}),
		events)
}

func freeUsageMocks(t *testing.T) (*repomocks.MockUserTrialRepository, *repomocks.MockUsageCounterRepository, *repomocks.MockPlanFeatureLimitRepository, *repomocks.MockUsageEventRepository) {
	t.Helper()
	trials := repomocks.NewMockUserTrialRepository(t)
	counters := repomocks.NewMockUsageCounterRepository(t)
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Once()
	trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Once()
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{}, nil).Once()
	return trials, counters, limits, events
}

func TestUsageActionDerivation(t *testing.T) {
	tests := []struct {
		name    string
		refType enums.UsageRefType
		amount  int
		want    string
	}{
		{name: "conversation is a voice session start", refType: enums.UsageRefTypeConversation, amount: 1, want: "voice_session_start"},
		{name: "exam attempt amount 1 is an AI review", refType: enums.UsageRefTypeExamAttempt, amount: 1, want: "exam_ai_review"},
		{name: "exam attempt amount 2 is a study plan", refType: enums.UsageRefTypeExamAttempt, amount: 2, want: "study_plan"},
		{name: "writing activity is a writing score", refType: enums.UsageRefTypeWritingActivity, amount: 1, want: "writing_score"},
		{name: "study plan ref is a study plan", refType: enums.UsageRefTypeStudyPlan, amount: 2, want: "study_plan"},
		{name: "unknown ref type renders as other", refType: enums.UsageRefType("lesson"), amount: 1, want: "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trials, counters, limits, events := freeUsageMocks(t)
			created := time.Now().UTC().Truncate(time.Second)
			events.EXPECT().ListByUserStatus(mock.Anything, "user_1",
				[]enums.UsageEventStatus{enums.UsageEventStatusCommitted, enums.UsageEventStatusReserved},
				20, 0).Return([]*models.UsageEvent{
				{ID: "evt-1", UserID: "user_1", Amount: tt.amount, Status: enums.UsageEventStatusCommitted,
					RefType: tt.refType, RefID: "ref-1", CreatedAt: created},
			}, nil).Once()
			events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(1), nil).Once()

			svc := newUsageSvc(t, fakeSubscriptions{}, trials, counters, limits, events)
			got, err := svc.MyUsage(context.Background(), "user_1", false, 1, 20)
			require.NoError(t, err)
			require.Len(t, got.History.Items, 1)
			item := got.History.Items[0]
			assert.Equal(t, tt.want, item.Action)
			assert.Equal(t, tt.amount, item.Credits, "credits are the event amount")
			assert.Equal(t, enums.UsageEventStatusCommitted, item.Status)
			assert.Equal(t, created, item.CreatedAt)
			assert.Equal(t, tt.refType, item.RefType)
			assert.Equal(t, "ref-1", item.RefID)
			assert.Equal(t, "evt-1", item.ID)
		})
	}
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
	// No counter read: admins bypass metering with zero views, like /premium/me.
	events.EXPECT().ListByUserStatus(mock.Anything, "user_1", mock.Anything, mock.Anything, mock.Anything).
		Return([]*models.UsageEvent{}, nil).Once()
	events.EXPECT().CountByUserStatus(mock.Anything, "user_1", mock.Anything).Return(int64(0), nil).Once()

	svc := accessservices.NewEntitlementsService(
		accessservices.NewPolicyService(fakeSubscriptions{}, repomocks.NewMockUserTrialRepository(t),
			repomocks.NewMockUsageCounterRepository(t), limits,
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
