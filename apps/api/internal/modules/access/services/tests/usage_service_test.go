// Reserve's replay matrix (spec §7.2) — the exactly-once contract. The
// concurrency and SQL behaviors are pinned by the integration suite; these
// mock tests pin the decision logic and error mapping.
package tests

import (
	"context"
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
	accessservices "engflex-api/internal/modules/access/services"
)

type fakeLimitSource struct {
	limits map[enums.PlanKey]map[enums.CounterKind]accessservices.Limit
}

func (f fakeLimitSource) LimitsForPlan(ctx context.Context, plan enums.PlanKey) (map[enums.CounterKind]accessservices.Limit, error) {
	return f.limits[plan], nil
}

func voiceCharges() []accessservices.Charge {
	return []accessservices.Charge{
		{Counter: enums.CounterKindAICredit, Amount: 1},
		{Counter: enums.CounterKindVoiceSession, Amount: 1},
	}
}

func eventRows(reservationID, userID string, status enums.UsageEventStatus) []*models.UsageEvent {
	return []*models.UsageEvent{
		{ID: "00000000-0000-0000-0000-0000000000e1", ReservationID: reservationID, UserID: userID,
			CounterKey: enums.CounterKindAICredit, PeriodKey: "2026-11", Amount: 1,
			Status: status, IdempotencyKey: "conversation:00000000-0000-0000-0000-0000000000f1:ai_credit"},
		{ID: "00000000-0000-0000-0000-0000000000e2", ReservationID: reservationID, UserID: userID,
			CounterKey: enums.CounterKindVoiceSession, PeriodKey: "2026-11", Amount: 1,
			Status: status, IdempotencyKey: "conversation:00000000-0000-0000-0000-0000000000f1:voice_session"},
	}
}

func TestReserveReplayMatrix(t *testing.T) {
	const (
		user   = "user_1"
		convID = "00000000-0000-0000-0000-0000000000f1"
	)
	fixedNow := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name        string
		existing    []*models.UsageEvent
		wantStatus  accessservices.ReservationStatus
		wantErrCode string
		wantErrStat int
	}{
		{
			name:       "committed events replay as committed, no new charge",
			existing:   eventRows("00000000-0000-0000-0000-0000000000aa", user, enums.UsageEventStatusCommitted),
			wantStatus: accessservices.ReservationStatusCommitted,
		},
		{
			name:        "unexpired reserved events report in progress",
			existing:    eventRows("00000000-0000-0000-0000-0000000000ab", user, enums.UsageEventStatusReserved),
			wantStatus:  accessservices.ReservationStatusInProgress,
			wantErrCode: "in_progress", wantErrStat: 409,
		},
		{
			name:       "released events are reused and re-reserved",
			existing:   eventRows("00000000-0000-0000-0000-0000000000ac", user, enums.UsageEventStatusReleased),
			wantStatus: accessservices.ReservationStatusReserved,
		},
		{
			name:       "expired events are reused and re-reserved",
			existing:   eventRows("00000000-0000-0000-0000-0000000000ad", user, enums.UsageEventStatusExpired),
			wantStatus: accessservices.ReservationStatusReserved,
		},
		{
			name:       "no existing rows reserve fresh",
			existing:   nil,
			wantStatus: accessservices.ReservationStatusReserved,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counters := repomocks.NewMockUsageCounterRepository(t)
			events := repomocks.NewMockUsageEventRepository(t)
			tx := fakeTxManager{}
			svc := accessservices.NewUsageService(counters, events, tx, fakeLimitSource{
				limits: map[enums.PlanKey]map[enums.CounterKind]accessservices.Limit{
					enums.PlanKeyFree: {
						enums.CounterKindAICredit:     {Enabled: true, Value: 15, Period: enums.UsagePeriodMonth},
						enums.CounterKindVoiceSession: {Enabled: true, Value: 3, Period: enums.UsagePeriodMonth},
					},
				},
			})

			counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
			events.EXPECT().WithTx(mock.Anything).Return(events).Once()
			events.EXPECT().ListByIdempotencyKeysForUpdate(mock.Anything, user, mock.Anything).Return(tt.existing, nil).Once()

			if tt.wantStatus == accessservices.ReservationStatusReserved {
				for _, e := range tt.existing {
					events.EXPECT().ReuseAsReserved(mock.Anything, e.ID, mock.Anything, mock.Anything).Return(nil).Once()
				}
				missing := 2 - len(tt.existing)
				for i := 0; i < missing; i++ {
					events.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil).Once()
				}
				counters.EXPECT().Reserve(mock.Anything, user, enums.CounterKindAICredit, "2026-11", 1, 15).Return(true, nil).Once()
				counters.EXPECT().Reserve(mock.Anything, user, enums.CounterKindVoiceSession, "2026-11", 1, 3).Return(true, nil).Once()
			}

			res, err := svc.Reserve(context.Background(), user, enums.PlanKeyFree, voiceCharges(),
				"conversation:"+convID, enums.UsageRefTypeConversation, convID, 2*time.Minute, fixedNow, time.UTC)
			if tt.wantErrCode != "" {
				appErr := &common.AppError{}
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.wantErrStat, appErr.Status)
				assert.Equal(t, tt.wantErrCode, appErr.Code)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantStatus, res.Status)
		})
	}
}

func TestReserveInsufficientCreditsRollsBack(t *testing.T) {
	fixedNow := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)
	counters := repomocks.NewMockUsageCounterRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	svc := accessservices.NewUsageService(counters, events, fakeTxManager{}, fakeLimitSource{
		limits: map[enums.PlanKey]map[enums.CounterKind]accessservices.Limit{
			enums.PlanKeyFree: {
				enums.CounterKindAICredit:     {Enabled: true, Value: 1, Period: enums.UsagePeriodMonth},
				enums.CounterKindVoiceSession: {Enabled: true, Value: 3, Period: enums.UsagePeriodMonth},
			},
		},
	})
	counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
	events.EXPECT().WithTx(mock.Anything).Return(events).Once()
	events.EXPECT().ListByIdempotencyKeysForUpdate(mock.Anything, "user_1", mock.Anything).Return(nil, nil).Once()
	events.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil).Times(2)
	// First counter fits, second is exhausted: the whole group must fail and
	// the tx rollback discards the first hold.
	counters.EXPECT().Reserve(mock.Anything, "user_1", enums.CounterKindAICredit, "2026-11", 1, 1).Return(true, nil).Once()
	counters.EXPECT().Reserve(mock.Anything, "user_1", enums.CounterKindVoiceSession, "2026-11", 1, 3).Return(false, nil).Once()

	_, err := svc.Reserve(context.Background(), "user_1", enums.PlanKeyFree, voiceCharges(),
		"conversation:00000000-0000-0000-0000-0000000000f2", enums.UsageRefTypeConversation,
		"00000000-0000-0000-0000-0000000000f2", 2*time.Minute, fixedNow, time.UTC)
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 429, appErr.Status)
	require.Equal(t, common.CodeQuotaExceeded, appErr.Code, "the non-credit counter denies")
}

func TestCommitIsIdempotent(t *testing.T) {
	counters := repomocks.NewMockUsageCounterRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	svc := accessservices.NewUsageService(counters, events, fakeTxManager{}, fakeLimitSource{})
	settled := eventRows("00000000-0000-0000-0000-0000000000ba", "user_1", enums.UsageEventStatusReserved)
	counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
	events.EXPECT().WithTx(mock.Anything).Return(events).Once()
	events.EXPECT().SettleReserved(mock.Anything, "00000000-0000-0000-0000-0000000000ba", enums.UsageEventStatusCommitted).
		Return(settled, nil).Once()
	for _, e := range settled {
		counters.EXPECT().Commit(mock.Anything, e.UserID, e.CounterKey, e.PeriodKey, e.Amount).Return(nil).Once()
	}
	require.NoError(t, svc.Commit(context.Background(), "00000000-0000-0000-0000-0000000000ba"))

	// Second commit: zero rows settle, no counter write happens.
	counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
	events.EXPECT().WithTx(mock.Anything).Return(events).Once()
	events.EXPECT().SettleReserved(mock.Anything, "00000000-0000-0000-0000-0000000000ba", enums.UsageEventStatusCommitted).
		Return(nil, nil).Once()
	require.NoError(t, svc.Commit(context.Background(), "00000000-0000-0000-0000-0000000000ba"),
		"double commit must be a silent no-op (spec §7.1)")
}

func TestReleaseDropsHolds(t *testing.T) {
	counters := repomocks.NewMockUsageCounterRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	svc := accessservices.NewUsageService(counters, events, fakeTxManager{}, fakeLimitSource{})
	released := eventRows("00000000-0000-0000-0000-0000000000bb", "user_1", enums.UsageEventStatusReserved)
	counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
	events.EXPECT().WithTx(mock.Anything).Return(events).Once()
	events.EXPECT().SettleReserved(mock.Anything, "00000000-0000-0000-0000-0000000000bb", enums.UsageEventStatusReleased).
		Return(released, nil).Once()
	for _, e := range released {
		counters.EXPECT().Release(mock.Anything, e.UserID, e.CounterKey, e.PeriodKey, e.Amount).Return(nil).Once()
	}
	require.NoError(t, svc.Release(context.Background(), "00000000-0000-0000-0000-0000000000bb"))
}

func TestReserveFailsClosedOnStoreError(t *testing.T) {
	fixedNow := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)
	counters := repomocks.NewMockUsageCounterRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	svc := accessservices.NewUsageService(counters, events, fakeTxManager{}, fakeLimitSource{})
	counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
	events.EXPECT().WithTx(mock.Anything).Return(events).Once()
	events.EXPECT().ListByIdempotencyKeysForUpdate(mock.Anything, "user_1", mock.Anything).
		Return(nil, gorm.ErrInvalidDB).Once()
	_, err := svc.Reserve(context.Background(), "user_1", enums.PlanKeyFree, voiceCharges(),
		"conversation:00000000-0000-0000-0000-0000000000f3", enums.UsageRefTypeConversation,
		"00000000-0000-0000-0000-0000000000f3", 2*time.Minute, fixedNow, time.UTC)
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 503, appErr.Status, "paid AI fails closed when the usage store is down (spec §12)")
	require.Equal(t, common.CodePolicyUnavailable, appErr.Code)
}

// fakeTxManager runs the fn against a nil *gorm.DB; the mock repos ignore it.
type fakeTxManager struct{}

func (fakeTxManager) Run(ctx context.Context, fn func(tx *gorm.DB) error) error { return fn(nil) }
