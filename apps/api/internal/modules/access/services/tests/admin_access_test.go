// Admin limits + user usage (spec §5.6, §8.6): batch validation, catalog
// read, user usage read. Binding-tag cases live in the controller table
// (controllers/tests); this file pins the service rules: plan allowlist,
// duplicate counters, empty payload, and error mapping.
package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/access/dtos/requests"
	accessservices "engflex-api/internal/modules/access/services"
)

// stubRefresher is the SnapshotCatalog seam: records refresh calls without
// a policy snapshot. Repository mocks stay in repositories/mocks; this is a
// service-local interface, not a repository.
type stubRefresher struct {
	calls int
	err   error
}

func (s *stubRefresher) SnapshotCatalog(ctx context.Context) error {
	s.calls++
	return s.err
}

func validLimitReq() requests.UpdatePlanLimits {
	return requests.UpdatePlanLimits{Counters: []requests.PlanLimitInput{
		{CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 15, Period: enums.UsagePeriodMonth},
		{CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: 3, Period: enums.UsagePeriodMonth},
	}}
}

func TestUpdatePlanLimitsValidation(t *testing.T) {
	tests := []struct {
		name    string
		plan    enums.PlanKey
		mutate  func(*requests.UpdatePlanLimits)
		wantErr int // 0 = success
	}{
		{
			name: "valid batch upsert",
			plan: enums.PlanKeyFree,
		},
		{
			name:    "admin plan rejected",
			plan:    enums.PlanKeyAdmin,
			wantErr: 400,
		},
		{
			name:    "unknown plan rejected",
			plan:    enums.PlanKey("mega"),
			wantErr: 400,
		},
		{
			name: "duplicate counter rejected",
			plan: enums.PlanKeyFree,
			mutate: func(r *requests.UpdatePlanLimits) {
				r.Counters = append(r.Counters, r.Counters[0])
			},
			wantErr: 422,
		},
		{
			name: "empty counters rejected",
			plan: enums.PlanKeyFree,
			mutate: func(r *requests.UpdatePlanLimits) {
				r.Counters = nil
			},
			wantErr: 400,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := repomocks.NewMockPlanFeatureLimitRepository(t)
			policy := &stubRefresher{}
			req := validLimitReq()
			if tt.mutate != nil {
				tt.mutate(&req)
			}
			if tt.wantErr == 0 {
				limits.EXPECT().UpsertBatch(mock.Anything, mock.Anything).RunAndReturn(
					func(ctx context.Context, rows []*models.PlanFeatureLimit) error {
						require.Len(t, rows, 2)
						for _, r := range rows {
							assert.Equal(t, tt.plan, r.PlanKey)
							assert.True(t, r.IsEnabled, "is_enabled comes from the request")
						}
						return nil
					}).Once()
			}
			svc := accessservices.NewAdminAccessService(limits, nil, nil, nil, nil, nil, policy)
			err := svc.UpdatePlanLimits(context.Background(), tt.plan, req)
			if tt.wantErr == 0 {
				require.NoError(t, err)
				assert.Equal(t, 1, policy.calls, "a successful edit invalidates the snapshot")
				return
			}
			appErr := &common.AppError{}
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, tt.wantErr, appErr.Status)
			assert.Equal(t, 0, policy.calls, "a rejected edit never touches the snapshot")
		})
	}
}

func TestUpdatePlanLimitsSnapshotRefreshIsBestEffort(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	limits.EXPECT().UpsertBatch(mock.Anything, mock.Anything).Return(nil).Once()
	policy := &stubRefresher{err: errors.New("snapshot store down")}
	svc := accessservices.NewAdminAccessService(limits, nil, nil, nil, nil, nil, policy)
	require.NoError(t, svc.UpdatePlanLimits(context.Background(), enums.PlanKeyFree, validLimitReq()),
		"the write landed; snapshot refresh is best-effort (TTL converges)")
}

func TestListLimits(t *testing.T) {
	tests := []struct {
		name    string
		rows    []*models.PlanFeatureLimit
		listErr error
		wantErr int
	}{
		{
			name: "catalog listed",
			rows: []*models.PlanFeatureLimit{
				{PlanKey: enums.PlanKeyFree, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: 15, Period: enums.UsagePeriodMonth},
			},
		},
		{
			name:    "store failure maps to 500",
			listErr: errors.New("db down"),
			wantErr: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := repomocks.NewMockPlanFeatureLimitRepository(t)
			limits.EXPECT().ListAll(mock.Anything).Return(tt.rows, tt.listErr).Once()
			svc := accessservices.NewAdminAccessService(limits, nil, nil, nil, nil, nil, &stubRefresher{})
			got, err := svc.ListLimits(context.Background())
			if tt.wantErr != 0 {
				appErr := &common.AppError{}
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.wantErr, appErr.Status)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.rows, got)
		})
	}
}

func TestUserUsageRead(t *testing.T) {
	counters := []*models.UsageCounter{
		{UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "2026-11", Used: 4},
	}
	events := []*models.UsageEvent{
		{ID: "00000000-0000-0000-0000-000000000001", UserID: "user_1", CounterKey: enums.CounterKindAICredit, PeriodKey: "2026-11", Amount: 1},
	}
	tests := []struct {
		name       string
		counterErr error
		eventErr   error
		wantErr    int
		wantEvents int
	}{
		{name: "counters plus recent events", wantEvents: 1},
		{name: "counter failure maps to 500", counterErr: errors.New("db down"), wantErr: 500},
		{name: "event failure maps to 500", eventErr: errors.New("db down"), wantErr: 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counterRepo := repomocks.NewMockUsageCounterRepository(t)
			eventRepo := repomocks.NewMockUsageEventRepository(t)
			counterRepo.EXPECT().ListByUser(mock.Anything, "user_1").Return(counters, tt.counterErr).Once()
			if tt.counterErr == nil {
				eventRepo.EXPECT().ListByUser(mock.Anything, "user_1", 50).Return(events, tt.eventErr).Once()
			}
			svc := accessservices.NewAdminAccessService(nil, counterRepo, eventRepo, nil, nil, nil, &stubRefresher{})
			got, err := svc.UserUsage(context.Background(), "user_1")
			if tt.wantErr != 0 {
				appErr := &common.AppError{}
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.wantErr, appErr.Status)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, counters, got.Counters)
			assert.Len(t, got.Events, tt.wantEvents)
		})
	}
}

func validCostReq() requests.UpdateActionCosts {
	return requests.UpdateActionCosts{Costs: []requests.ActionCostInput{
		{Action: enums.AccessActionVoiceStart, Credits: 1},
		{Action: enums.AccessActionExamReview, Credits: 1},
		{Action: enums.AccessActionStudyPlan, Credits: 2},
	}}
}

func TestUpdateActionCostsValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*requests.UpdateActionCosts)
		wantErr int
	}{
		{name: "valid batch"},
		{name: "free action rejected", wantErr: 422, mutate: func(r *requests.UpdateActionCosts) {
			r.Costs[0].Action = enums.AccessActionWritingScore
		}},
		{name: "duplicate action rejected", wantErr: 422, mutate: func(r *requests.UpdateActionCosts) {
			r.Costs[1].Action = r.Costs[0].Action
		}},
		{name: "missing action rejected", wantErr: 400, mutate: func(r *requests.UpdateActionCosts) {
			r.Costs = r.Costs[:2]
		}},
		{name: "negative credits rejected", wantErr: 400, mutate: func(r *requests.UpdateActionCosts) {
			r.Costs[0].Credits = -1
		}},
		{name: "above-ceiling credits rejected", wantErr: 400, mutate: func(r *requests.UpdateActionCosts) {
			r.Costs[1].Credits = common.MaxPlanLimitValue + 1
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			costs := repomocks.NewMockActionCreditCostRepository(t)
			policy := &stubRefresher{}
			req := validCostReq()
			if tt.mutate != nil {
				tt.mutate(&req)
			}
			if tt.wantErr == 0 {
				costs.EXPECT().UpsertBatch(mock.Anything, mock.Anything).RunAndReturn(
					func(ctx context.Context, rows []*models.ActionCreditCost) error {
						require.Len(t, rows, 3)
						assert.Equal(t, "admin_1", rows[0].UpdatedBy)
						return nil
					}).Once()
			}
			svc := accessservices.NewAdminAccessService(nil, nil, nil, costs, nil, nil, policy)
			err := svc.UpdateActionCosts(context.Background(), "admin_1", req)
			if tt.wantErr == 0 {
				require.NoError(t, err)
				assert.Equal(t, 1, policy.calls)
				return
			}
			appErr := requireAppError(t, err, tt.wantErr)
			assert.NotEmpty(t, appErr.Message)
		})
	}
}

func TestUpdateTrialTermsWritesBothTablesInOneTx(t *testing.T) {
	limits := repomocks.NewMockPlanFeatureLimitRepository(t)
	settings := repomocks.NewMockTrialSettingRepository(t)
	tx := repomocks.NewMockTransactionManager(t)
	policy := &stubRefresher{}
	tx.EXPECT().Run(mock.Anything, mock.Anything).RunAndReturn(
		func(ctx context.Context, fn func(*gorm.DB) error) error { return fn(nil) }).Once()
	limits.EXPECT().WithTx(mock.Anything).Return(limits).Once()
	settings.EXPECT().WithTx(mock.Anything).Return(settings).Once()
	settings.EXPECT().Upsert(mock.Anything, mock.MatchedBy(func(m *models.TrialSetting) bool {
		return m.DurationDays == 10 && m.UpdatedBy == "admin_1"
	})).Return(nil).Once()
	limits.EXPECT().UpsertBatch(mock.Anything, mock.MatchedBy(func(rows []*models.PlanFeatureLimit) bool {
		return len(rows) == 2 && rows[0].PlanKey == enums.PlanKeyTrial &&
			rows[0].LimitValue == 40 && rows[1].LimitValue == 10
	})).Return(nil).Once()
	svc := accessservices.NewAdminAccessService(limits, nil, nil, nil, settings, tx, policy)
	require.NoError(t, svc.UpdateTrialTerms(context.Background(), "admin_1",
		requests.UpdateTrialTerms{Days: 10, Credits: 40, VoiceSessions: 10}))
	assert.Equal(t, 1, policy.calls)
}

// Direct service calls bypass gin binding, so the service itself must range
// check: out-of-range values fail 400 without ever touching the store.
func TestUpdateTrialTermsRangeValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*requests.UpdateTrialTerms)
	}{
		{name: "zero days rejected", mutate: func(r *requests.UpdateTrialTerms) { r.Days = 0 }},
		{name: "negative days rejected", mutate: func(r *requests.UpdateTrialTerms) { r.Days = -1 }},
		{name: "above-ceiling days rejected", mutate: func(r *requests.UpdateTrialTerms) { r.Days = common.MaxTrialDays + 1 }},
		{name: "negative credits rejected", mutate: func(r *requests.UpdateTrialTerms) { r.Credits = -1 }},
		{name: "above-ceiling credits rejected", mutate: func(r *requests.UpdateTrialTerms) { r.Credits = common.MaxPlanLimitValue + 1 }},
		{name: "negative voice sessions rejected", mutate: func(r *requests.UpdateTrialTerms) { r.VoiceSessions = -1 }},
		{name: "above-ceiling voice sessions rejected", mutate: func(r *requests.UpdateTrialTerms) {
			r.VoiceSessions = common.MaxPlanLimitValue + 1
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No tx/store expectations: any tx.Run call fails the test.
			tx := repomocks.NewMockTransactionManager(t)
			policy := &stubRefresher{}
			svc := accessservices.NewAdminAccessService(
				repomocks.NewMockPlanFeatureLimitRepository(t), nil, nil, nil,
				repomocks.NewMockTrialSettingRepository(t), tx, policy)
			req := requests.UpdateTrialTerms{Days: 10, Credits: 40, VoiceSessions: 10}
			tt.mutate(&req)
			requireAppError(t, svc.UpdateTrialTerms(context.Background(), "admin_1", req), 400)
			assert.Equal(t, 0, policy.calls, "a rejected edit never touches the snapshot")
		})
	}
}
