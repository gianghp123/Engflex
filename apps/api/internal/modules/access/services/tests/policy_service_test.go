// PolicyService contract: plan resolution order (paid > active trial >
// free), snapshot fallback to defaults when the catalog is unreachable,
// fail-closed Decide for paid AI when the usage store is down, and period
// math in the configured timezone.
package tests

import (
	"bytes"
	"context"
	"log/slog"
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
	"engflex-api/internal/logger"
	accessservices "engflex-api/internal/modules/access/services"
)

type fakeSubscriptions struct {
	sub *models.Subscription
	err error
}

func (f fakeSubscriptions) GetActiveByUserID(ctx context.Context, userID string) (*models.Subscription, error) {
	return f.sub, f.err
}

func paidSub(plan enums.SubscriptionPlan) *models.Subscription {
	ends := time.Now().AddDate(0, 1, 0)
	return &models.Subscription{Plan: plan, Status: enums.SubscriptionStatusActive, ExpiresAt: &ends}
}

func trialRow(daysLeft int) *models.UserTrial {
	return &models.UserTrial{UserID: "user_1", StartedAt: time.Now().AddDate(0, 0, -7),
		EndsAt: time.Now().Add(time.Duration(daysLeft) * 24 * time.Hour), CreditsGranted: 30}
}

func defaultCostRows() []*models.ActionCreditCost {
	return []*models.ActionCreditCost{
		{Action: enums.AccessActionVoiceStart, Credits: 1},
		{Action: enums.AccessActionExamReview, Credits: 1},
		{Action: enums.AccessActionStudyPlan, Credits: 2},
	}
}

func newPolicy(t *testing.T, subs accessservices.SubscriptionPlanSource, trials *repomocks.MockUserTrialRepository, counters *repomocks.MockUsageCounterRepository, limits *repomocks.MockPlanFeatureLimitRepository) *accessservices.PolicyService {
	t.Helper()
	costs := repomocks.NewMockActionCreditCostRepository(t)
	costs.EXPECT().List(mock.Anything).Return(defaultCostRows(), nil).Maybe()
	settings := repomocks.NewMockTrialSettingRepository(t)
	settings.EXPECT().Get(mock.Anything).Return(&models.TrialSetting{ID: true, DurationDays: 7}, nil).Maybe()
	return accessservices.NewPolicyService(subs, trials, counters, limits, costs, settings, config.AccessConfig{
		Enforce: true, ContentGating: false, Timezone: "Asia/Ho_Chi_Minh",
	})
}

func TestPrincipalResolutionPlanOrder(t *testing.T) {
	tests := []struct {
		name     string
		sub      *models.Subscription
		subErr   error
		trial    *models.UserTrial
		wantPlan enums.PlanKey
	}{
		{name: "paid subscription wins over trial", sub: paidSub(enums.SubscriptionPlanOneMonth), trial: trialRow(3), wantPlan: enums.PlanKey1M},
		{name: "active trial applies without a subscription", trial: trialRow(3), wantPlan: enums.PlanKeyTrial},
		{name: "expired trial falls back to free", trial: trialRow(-1), wantPlan: enums.PlanKeyFree},
		{name: "no subscription and no trial is free", wantPlan: enums.PlanKeyFree},
		{name: "subscription lookup error still honors an active trial", subErr: gorm.ErrInvalidDB, trial: trialRow(3), wantPlan: enums.PlanKeyTrial},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trials := repomocks.NewMockUserTrialRepository(t)
			counters := repomocks.NewMockUsageCounterRepository(t)
			limits := repomocks.NewMockPlanFeatureLimitRepository(t)
			// Principal consults the trial store unless a paid subscription
			// wins first — including when the subscription lookup errors, so
			// a billing-store flap never downgrades a trial user to free.
			if tt.trial != nil {
				trials.EXPECT().Get(mock.Anything, "user_1").Return(tt.trial, nil).Maybe()
			} else {
				trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Maybe()
			}
			svc := newPolicy(t, fakeSubscriptions{sub: tt.sub, err: tt.subErr}, trials, counters, limits)
			p, trial, err := svc.Principal(context.Background(), "user_1", false)
			require.NoError(t, err)
			assert.Equal(t, tt.wantPlan, p.Plan)
			if tt.wantPlan == enums.PlanKeyTrial {
				assert.True(t, trial.Active)
			}
		})
	}
}

func TestDecideServesSnapshotAndFailsClosed(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return([]*models.PlanFeatureLimit{
		{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 100, Period: enums.UsagePeriodMonth},
		{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 60, Period: enums.UsagePeriodMonth},
	}, nil).Once()

	counters := repomocks.NewMockUsageCounterRepository(t)
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{}, nil).Once()
	svc := newPolicy(t, fakeSubscriptions{sub: paidSub(enums.SubscriptionPlanOneMonth)},
		repomocks.NewMockUserTrialRepository(t), counters, limits)

	d, err := svc.Decide(context.Background(), "user_1", false, accessservices.ActionExamReview)
	require.NoError(t, err)
	require.True(t, d.Allow)
	assert.Equal(t, enums.PlanKey1M, d.Plan)
}

func TestDecidePolicyStoreUnavailable(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	// Decide resolves limits before reading counters; the catalog read must
	// be stubbed so the counters outage is what fails the decision.
	limits.EXPECT().ListAll(mock.Anything).Return(nil, gorm.ErrInvalidDB).Maybe()
	counters := repomocks.NewMockUsageCounterRepository(t)
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return(nil, gorm.ErrInvalidDB).Once()
	svc := newPolicy(t, fakeSubscriptions{sub: paidSub(enums.SubscriptionPlanOneMonth)},
		repomocks.NewMockUserTrialRepository(t), counters, limits)

	_, err := svc.Decide(context.Background(), "user_1", false, accessservices.ActionExamReview)
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 503, appErr.Status, "paid AI fails closed when usage is unreadable (spec §12)")
	require.Equal(t, common.CodePolicyUnavailable, appErr.Code)

	// Completion-critical AI never consults the store: the same outage must
	// still allow a writing score.
	d, err := svc.Decide(context.Background(), "user_1", false, accessservices.ActionWritingScore)
	require.NoError(t, err)
	assert.True(t, d.Allow)
}

func TestCatalogSnapshotDrivesChargesAndTrialTerms(t *testing.T) {
	ctx := context.Background()
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	costs := repomocks.NewMockActionCreditCostRepository(t)
	settings := repomocks.NewMockTrialSettingRepository(t)
	trials := repomocks.NewMockUserTrialRepository(t)
	counters := repomocks.NewMockUsageCounterRepository(t)

	limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Once()
	costs.EXPECT().List(mock.Anything).Return([]*models.ActionCreditCost{
		{Action: enums.AccessActionVoiceStart, Credits: 3},
		{Action: enums.AccessActionExamReview, Credits: 0},
		{Action: enums.AccessActionStudyPlan, Credits: 5},
	}, nil).Once()
	settings.EXPECT().Get(mock.Anything).Return(&models.TrialSetting{ID: true, DurationDays: 10}, nil).Once()

	svc := accessservices.NewPolicyService(fakeSubscriptions{}, trials, counters, limits, costs, settings,
		config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
	require.NoError(t, svc.SnapshotCatalog(ctx))

	assert.Equal(t, []accessservices.Charge{
		{Counter: enums.CounterKindAICredit, Amount: 3},
		{Counter: enums.CounterKindVoiceSession, Amount: 1},
	}, svc.Charges(accessservices.ActionVoiceStart))
	assert.Empty(t, svc.Charges(accessservices.ActionExamReview), "zero credits = free for an exam review")
	assert.Equal(t, 5, svc.CreditCost(accessservices.ActionStudyPlan))
	assert.Equal(t, []accessservices.CostEntry{
		{Action: accessservices.ActionVoiceStart, Credits: 3},
		{Action: accessservices.ActionExamReview, Credits: 0},
		{Action: accessservices.ActionStudyPlan, Credits: 5},
	}, svc.ActionCosts())
	assert.Equal(t, accessservices.TrialTerms{Days: 10, Credits: 30, VoiceSessions: 7}, svc.TrialTerms())
}

func TestDecideZeroCostSkipsUsageReads(t *testing.T) {
	ctx := context.Background()
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	costs := repomocks.NewMockActionCreditCostRepository(t)
	settings := repomocks.NewMockTrialSettingRepository(t)
	trials := repomocks.NewMockUserTrialRepository(t)
	counters := repomocks.NewMockUsageCounterRepository(t)
	subs := fakeSubscriptions{}

	limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Once()
	costs.EXPECT().List(mock.Anything).Return([]*models.ActionCreditCost{
		{Action: enums.AccessActionExamReview, Credits: 0},
		{Action: enums.AccessActionVoiceStart, Credits: 1},
		{Action: enums.AccessActionStudyPlan, Credits: 2},
	}, nil).Once()
	settings.EXPECT().Get(mock.Anything).Return(&models.TrialSetting{ID: true, DurationDays: 7}, nil).Once()
	trials.EXPECT().Get(mock.Anything, "user_1").Return(nil, gorm.ErrRecordNotFound).Once()

	svc := accessservices.NewPolicyService(subs, trials, counters, limits, costs, settings,
		config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
	require.NoError(t, svc.SnapshotCatalog(ctx))
	d, err := svc.Decide(ctx, "user_1", false, accessservices.ActionExamReview)
	require.NoError(t, err)
	assert.True(t, d.Allow)
	assert.Empty(t, d.Charges)
	// counters.ListByUser has no expectation: it must not be called.
}

func TestSnapshotFallsBackToDefaults(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return(nil, gorm.ErrInvalidDB).Once()
	svc := newPolicy(t, fakeSubscriptions{},
		repomocks.NewMockUserTrialRepository(t), repomocks.NewMockUsageCounterRepository(t), limits)
	require.NoError(t, svc.SnapshotCatalog(context.Background()), "a catalog outage must never fail the process (spec §5.6)")
	got, err := svc.LimitsForPlan(context.Background(), enums.PlanKeyFree)
	require.NoError(t, err)
	assert.Equal(t, 15, got[enums.CounterKindAICredit].Value)
}

func TestSnapshotKeepsLastGoodCatalogOnReloadFailure(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return([]*models.PlanFeatureLimit{
		{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 42, Period: enums.UsagePeriodMonth},
	}, nil).Once()
	limits.EXPECT().ListAll(mock.Anything).Return(nil, gorm.ErrInvalidDB).Once()
	svc := newPolicy(t, fakeSubscriptions{},
		repomocks.NewMockUserTrialRepository(t), repomocks.NewMockUsageCounterRepository(t), limits)
	ctx := context.Background()
	require.NoError(t, svc.SnapshotCatalog(ctx))
	require.NoError(t, svc.SnapshotCatalog(ctx), "a failed reload must not fail the process")
	got, err := svc.LimitsForPlan(ctx, enums.PlanKeyFree)
	require.NoError(t, err)
	assert.Equal(t, 42, got[enums.CounterKindAICredit].Value, "the last-good catalog survives a reload failure")
}

func TestDecideTrialPlanWithinLimits(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return([]*models.PlanFeatureLimit{
		{PlanKey: enums.PlanKeyTrial, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 30, Period: enums.UsagePeriodTotal},
		{PlanKey: enums.PlanKeyTrial, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 7, Period: enums.UsagePeriodTotal},
	}, nil).Once()
	trials := repomocks.NewMockUserTrialRepository(t)
	trials.EXPECT().Get(mock.Anything, "user_1").Return(trialRow(30), nil).Once()
	counters := repomocks.NewMockUsageCounterRepository(t)
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{}, nil).Once()
	svc := newPolicy(t, fakeSubscriptions{}, trials, counters, limits)

	d, err := svc.Decide(context.Background(), "user_1", false, accessservices.ActionVoiceStart)
	require.NoError(t, err)
	require.True(t, d.Allow)
	assert.Equal(t, enums.PlanKeyTrial, d.Plan)
}

func TestSnapshotWarnsOnMissingRowsAndFallsSoft(t *testing.T) {
	tests := []struct {
		name       string
		setting    *models.TrialSetting
		settingErr error
		wantLog    string
	}{
		{name: "missing trial setting warns", setting: nil, settingErr: gorm.ErrRecordNotFound, wantLog: "trial setting missing; using fallback"},
		{name: "non-positive duration warns", setting: &models.TrialSetting{ID: true, DurationDays: 0}, settingErr: nil, wantLog: "trial setting invalid; using fallback"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			prev := slog.Default()
			slog.SetDefault(logger.NewWithWriter(&buf, "warn", "json"))
			t.Cleanup(func() { slog.SetDefault(prev) })

			ctx := context.Background()
			limits := repomocks.NewMockPlanFeatureLimitRepository(t)
			costs := repomocks.NewMockActionCreditCostRepository(t)
			settings := repomocks.NewMockTrialSettingRepository(t)

			limits.EXPECT().ListAll(mock.Anything).Return(usageCatalog(), nil).Once()
			// The study-plan cost row is absent: the snapshot must warn and
			// fall back to the default instead of silently serving it.
			costs.EXPECT().List(mock.Anything).Return([]*models.ActionCreditCost{
				{Action: enums.AccessActionVoiceStart, Credits: 3},
				{Action: enums.AccessActionExamReview, Credits: 0},
			}, nil).Once()
			settings.EXPECT().Get(mock.Anything).Return(tt.setting, tt.settingErr).Once()

			svc := accessservices.NewPolicyService(fakeSubscriptions{},
				repomocks.NewMockUserTrialRepository(t), repomocks.NewMockUsageCounterRepository(t),
				limits, costs, settings,
				config.AccessConfig{Enforce: true, Timezone: "Asia/Ho_Chi_Minh"})
			require.NoError(t, svc.SnapshotCatalog(ctx), "a missing row must never fail the snapshot")

			out := buf.String()
			require.Contains(t, out, "action credit cost missing; using fallback")
			require.Contains(t, out, "study_plan", "the warning names the absent action")
			require.Contains(t, out, tt.wantLog)
			assert.Equal(t, 2, svc.CreditCost(accessservices.ActionStudyPlan), "missing cost row falls back to the default")
			assert.Equal(t, accessservices.TrialTerms{Days: common.DefaultTrialDays, Credits: 30, VoiceSessions: 7}, svc.TrialTerms())
		})
	}
}

func TestDecideAdminBypassesUsageStore(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return([]*models.PlanFeatureLimit{
		{PlanKey: enums.PlanKeyAdmin, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 1, Period: enums.UsagePeriodMonth},
	}, nil).Once()
	counters := repomocks.NewMockUsageCounterRepository(t)
	svc := newPolicy(t, fakeSubscriptions{},
		repomocks.NewMockUserTrialRepository(t), counters, limits)

	d, err := svc.Decide(context.Background(), "user_1", true, accessservices.ActionVoiceStart)
	require.NoError(t, err)
	require.True(t, d.Allow)
	assert.Equal(t, "admin_bypass", d.Code)
	assert.True(t, counters.AssertNotCalled(t, "ListByUser", mock.Anything, "user_1"),
		"admin decisions never read the usage store")
}

func TestDecideDayPeriodFiltering(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return([]*models.PlanFeatureLimit{
		{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 1, Period: enums.UsagePeriodDay},
	}, nil).Once()
	counters := repomocks.NewMockUsageCounterRepository(t)
	svc := newPolicy(t, fakeSubscriptions{sub: paidSub(enums.SubscriptionPlanOneMonth)},
		repomocks.NewMockUserTrialRepository(t), counters, limits)
	dayKey := svc.PeriodKey(enums.UsagePeriodDay, time.Now().UTC())
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{
		{UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: dayKey, Used: 1, Reserved: 0},
	}, nil).Once()

	d, err := svc.Decide(context.Background(), "user_1", false, accessservices.ActionExamReview)
	require.NoError(t, err)
	require.False(t, d.Allow, "a day-keyed row at the day limit must deny")
	assert.Equal(t, common.CodeInsufficientCredits, d.Code)
}

func TestDecideIgnoresUsageForCountersAbsentFromLimits(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	limits.EXPECT().ListAll(mock.Anything).Return([]*models.PlanFeatureLimit{
		{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 100, Period: enums.UsagePeriodMonth},
	}, nil).Once()
	counters := repomocks.NewMockUsageCounterRepository(t)
	svc := newPolicy(t, fakeSubscriptions{sub: paidSub(enums.SubscriptionPlanOneMonth)},
		repomocks.NewMockUserTrialRepository(t), counters, limits)
	monthKey := svc.PeriodKey(enums.UsagePeriodMonth, time.Now().UTC())
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{
		{UserID: "user_1", CounterKey: enums.CounterKindVoiceSession, PeriodKey: monthKey, Used: 9999, Reserved: 0},
	}, nil).Once()

	d, err := svc.Decide(context.Background(), "user_1", false, accessservices.ActionExamReview)
	require.NoError(t, err)
	assert.True(t, d.Allow, "a usage row for a counter absent from the plan snapshot must not influence the decision")
}

func TestPeriodKeysCutInConfiguredTimezone(t *testing.T) {
	svc := newPolicy(t, fakeSubscriptions{},
		repomocks.NewMockUserTrialRepository(t), repomocks.NewMockUsageCounterRepository(t),
		repomocks.NewMockPlanFeatureLimitRepository(t))
	// 2026-11-26 23:30 UTC is already 2026-11-27 in Ho_Chi_Minh (+07).
	now := time.Date(2026, 11, 26, 23, 30, 0, 0, time.UTC)
	assert.Equal(t, "2026-11", svc.PeriodKey(enums.UsagePeriodMonth, now))
	assert.Equal(t, "2026-11-27", svc.PeriodKey(enums.UsagePeriodDay, now), "the day key follows APP_TIMEZONE, not UTC")
	assert.Equal(t, "all", svc.PeriodKey(enums.UsagePeriodTotal, now))

	reset := svc.ResetAt(enums.UsagePeriodMonth, now)
	assert.Equal(t, "2026-12-01 00:00:00 +0700 +0700", reset.Format("2006-01-02 15:04:05 -0700 Z0700"))
}

func TestAppErrorMapsDenials(t *testing.T) {
	deny := func(code string) accessservices.Decision {
		return accessservices.Decision{
			Allow: false, Code: code, Plan: enums.PlanKeyFree,
			Charges:         []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}},
			LimitingCounter: enums.CounterKindAICredit, Limit: 15, Used: 15, Remaining: 0,
			RequiredPlan: enums.PlanKey1M,
		}
	}
	tests := []struct {
		name       string
		decision   accessservices.Decision
		wantStatus int
		wantCode   string
	}{
		{name: "allow maps to nil", decision: accessservices.Decision{Allow: true}},
		{name: "insufficient credits is 429", decision: deny(common.CodeInsufficientCredits), wantStatus: 429, wantCode: common.CodeInsufficientCredits},
		{name: "quota exceeded is 429", decision: deny(common.CodeQuotaExceeded), wantStatus: 429, wantCode: common.CodeQuotaExceeded},
		{name: "feature not in plan is 403", decision: deny(common.CodeFeatureNotInPlan), wantStatus: 403, wantCode: common.CodeFeatureNotInPlan},
		{name: "premium required is 403", decision: deny(common.CodePremiumRequired), wantStatus: 403, wantCode: common.CodePremiumRequired},
		{name: "trial expired is 403", decision: deny(common.CodeTrialExpired), wantStatus: 403, wantCode: common.CodeTrialExpired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := accessservices.AppError(tt.decision)
			if tt.wantCode == "" {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.wantStatus, got.Status)
			assert.Equal(t, tt.wantCode, got.Code)
			details, ok := got.Details.(map[string]any)
			require.True(t, ok, "denials carry a Details payload")
			for _, key := range []string{"code", "feature", "required", "limit", "used", "remaining", "resetAt", "requiredPlan", "upgradeUrl"} {
				assert.Contains(t, details, key)
			}
			assert.Equal(t, tt.wantCode, details["code"])
			assert.Equal(t, 1, details["required"], "the limiting charge amount is reported")
		})
	}
}

// Decide stamps the limiting counter's reset instant: rolling periods get
// the next boundary in APP_TIMEZONE, total periods get nil (the UI shows
// "no reset"). The denial wire detail mirrors it RFC3339-or-nil.
func TestDecideResetAtFollowsLimitingPeriod(t *testing.T) {
	newSvc := func(t *testing.T, limits []*models.PlanFeatureLimit, counters []*models.UsageCounter) *accessservices.PolicyService {
		t.Helper()
		limitsRepo := repomocks.NewMockPlanFeatureLimitRepository(t)
		limitsRepo.EXPECT().ListAll(mock.Anything).Return(limits, nil).Once()
		countersRepo := repomocks.NewMockUsageCounterRepository(t)
		countersRepo.EXPECT().ListByUser(mock.Anything, "user_1").Return(counters, nil).Once()
		return newPolicy(t, fakeSubscriptions{sub: paidSub(enums.SubscriptionPlanOneMonth)},
			repomocks.NewMockUserTrialRepository(t), countersRepo, limitsRepo)
	}
	tests := []struct {
		name       string
		limits     []*models.PlanFeatureLimit
		counters   []*models.UsageCounter
		wantReset  bool
		wantDetail any
	}{
		{
			name: "month denial carries the next reset",
			limits: []*models.PlanFeatureLimit{
				{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 1, Period: enums.UsagePeriodMonth},
			},
			counters: []*models.UsageCounter{
				{UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "2026-11", Used: 1},
			},
			wantReset: true,
		},
		{
			name: "total denial never resets",
			limits: []*models.PlanFeatureLimit{
				{PlanKey: enums.PlanKey1M, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 1, Period: enums.UsagePeriodTotal},
			},
			counters: []*models.UsageCounter{
				{UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "all", Used: 1},
			},
			wantReset: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc(t, tt.limits, tt.counters)
			d, err := svc.Decide(context.Background(), "user_1", false, accessservices.ActionExamReview)
			require.NoError(t, err)
			require.False(t, d.Allow)
			if tt.wantReset {
				require.NotNil(t, d.ResetAt, "a month-limit denial names its reset")
			} else {
				require.Nil(t, d.ResetAt, "a total-limit denial never resets")
			}
			wire := accessservices.AppError(d)
			require.NotNil(t, wire)
			details, ok := wire.Details.(map[string]any)
			require.True(t, ok)
			assert.Contains(t, details, "required")
			assert.Contains(t, details, "remaining")
			if tt.wantReset {
				resetAt, ok := details["resetAt"].(string)
				require.True(t, ok, "a month-limit denial renders resetAt RFC3339")
				parsed, perr := time.Parse(time.RFC3339, resetAt)
				require.NoError(t, perr)
				assert.Equal(t, d.ResetAt.Format(time.RFC3339), parsed.Format(time.RFC3339))
			} else {
				assert.Nil(t, details["resetAt"], "a total-limit denial renders a null resetAt")
			}
		})
	}
}
