// Reserve's replay matrix (spec §7.2) — the exactly-once contract. The
// concurrency and SQL behaviors are pinned by the integration suite; these
// mock tests pin the decision logic and error mapping.
package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
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
					events.EXPECT().ReuseAsReserved(mock.Anything, e.ID, mock.Anything, accessservices.ActionVoiceStart, 1, mock.Anything).Return(true, nil).Once()
				}
				missing := 2 - len(tt.existing)
				for i := 0; i < missing; i++ {
					events.EXPECT().Insert(mock.Anything, mock.Anything).Return(nil).Once()
				}
				counters.EXPECT().Reserve(mock.Anything, user, enums.CounterKindAICredit, "2026-11", 1, 15).Return(true, nil).Once()
				counters.EXPECT().Reserve(mock.Anything, user, enums.CounterKindVoiceSession, "2026-11", 1, 3).Return(true, nil).Once()
			}

			res, err := svc.Reserve(context.Background(), user, enums.PlanKeyFree, accessservices.ActionVoiceStart, voiceCharges(),
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
	// The denial reports the live used/remaining so the UI can explain it.
	counters.EXPECT().ListByUser(mock.Anything, "user_1").Return([]*models.UsageCounter{
		{UserID: "user_1", CounterKey: enums.CounterKindVoiceSession, PeriodKey: "2026-11", Used: 3, Reserved: 0},
	}, nil).Once()

	_, err := svc.Reserve(context.Background(), "user_1", enums.PlanKeyFree, accessservices.ActionVoiceStart, voiceCharges(),
		"conversation:00000000-0000-0000-0000-0000000000f2", enums.UsageRefTypeConversation,
		"00000000-0000-0000-0000-0000000000f2", 2*time.Minute, fixedNow, time.UTC)
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 429, appErr.Status)
	require.Equal(t, common.CodeQuotaExceeded, appErr.Code, "the non-credit counter denies")
	details, ok := appErr.Details.(map[string]any)
	require.True(t, ok, "denials carry a Details payload")
	assert.Equal(t, 1, details["required"], "the limiting charge amount is reported")
	assert.Equal(t, 0, details["remaining"], "the counter is exhausted")
	assert.NotNil(t, details["resetAt"], "a month-limit denial names its reset instant")
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
	_, err := svc.Reserve(context.Background(), "user_1", enums.PlanKeyFree, accessservices.ActionVoiceStart, voiceCharges(),
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

// Reserve reuse guards (spec §7.2): the locked rows must form a whole group
// in a reusable state. Partial groups, mixed committed/sibling states, and
// reuses that move nothing all fail closed with 503 — never a duplicate
// insert, never a half-charged group.
func TestReserveReuseGuardsFailClosed(t *testing.T) {
	fixedNow := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)
	newSvc := func(t *testing.T, counters *repomocks.MockUsageCounterRepository, events *repomocks.MockUsageEventRepository) *accessservices.UsageService {
		t.Helper()
		return accessservices.NewUsageService(counters, events, fakeTxManager{}, fakeLimitSource{
			limits: map[enums.PlanKey]map[enums.CounterKind]accessservices.Limit{
				enums.PlanKeyFree: {
					enums.CounterKindAICredit:     {Enabled: true, Value: 15, Period: enums.UsagePeriodMonth},
					enums.CounterKindVoiceSession: {Enabled: true, Value: 3, Period: enums.UsagePeriodMonth},
				},
			},
		})
	}
	reserve := func(t *testing.T, svc *accessservices.UsageService) error {
		t.Helper()
		_, err := svc.Reserve(context.Background(), "user_1", enums.PlanKeyFree, accessservices.ActionVoiceStart, voiceCharges(),
			"conversation:00000000-0000-0000-0000-0000000000f1", enums.UsageRefTypeConversation,
			"00000000-0000-0000-0000-0000000000f1", 2*time.Minute, fixedNow, time.UTC)
		return err
	}

	tests := []struct {
		name  string
		setup func(*repomocks.MockUsageCounterRepository, *repomocks.MockUsageEventRepository)
	}{
		{
			name: "partial group fails closed",
			setup: func(counters *repomocks.MockUsageCounterRepository, events *repomocks.MockUsageEventRepository) {
				counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
				events.EXPECT().WithTx(mock.Anything).Return(events).Once()
				// One of two keys present: a crash cannot produce this.
				events.EXPECT().ListByIdempotencyKeysForUpdate(mock.Anything, "user_1", mock.Anything).
					Return(eventRows("00000000-0000-0000-0000-0000000000ac", "user_1", enums.UsageEventStatusReleased)[:1], nil).Once()
			},
		},
		{
			name: "mixed committed and released fails closed",
			setup: func(counters *repomocks.MockUsageCounterRepository, events *repomocks.MockUsageEventRepository) {
				counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
				events.EXPECT().WithTx(mock.Anything).Return(events).Once()
				rows := eventRows("00000000-0000-0000-0000-0000000000ac", "user_1", enums.UsageEventStatusReleased)
				rows[0].Status = enums.UsageEventStatusCommitted
				events.EXPECT().ListByIdempotencyKeysForUpdate(mock.Anything, "user_1", mock.Anything).
					Return(rows, nil).Once()
			},
		},
		{
			name: "reuse that moves nothing fails closed",
			setup: func(counters *repomocks.MockUsageCounterRepository, events *repomocks.MockUsageEventRepository) {
				counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
				events.EXPECT().WithTx(mock.Anything).Return(events).Once()
				events.EXPECT().ListByIdempotencyKeysForUpdate(mock.Anything, "user_1", mock.Anything).
					Return(eventRows("00000000-0000-0000-0000-0000000000ac", "user_1", enums.UsageEventStatusReleased), nil).Once()
				// A concurrent settle moved the row first: nothing reusable.
				events.EXPECT().ReuseAsReserved(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counters := repomocks.NewMockUsageCounterRepository(t)
			events := repomocks.NewMockUsageEventRepository(t)
			tt.setup(counters, events)
			err := reserve(t, newSvc(t, counters, events))
			appErr := &common.AppError{}
			require.ErrorAs(t, err, &appErr)
			require.Equal(t, 503, appErr.Status, "inconsistent reservation state fails closed")
			require.Equal(t, common.CodePolicyUnavailable, appErr.Code)
		})
	}
}

// A duplicate idempotency insert means a concurrent Reserve won the same key:
// the loser reports in_progress (409), not a store outage.
func TestReserveDuplicateInsertReportsInProgress(t *testing.T) {
	fixedNow := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)
	counters := repomocks.NewMockUsageCounterRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	svc := accessservices.NewUsageService(counters, events, fakeTxManager{}, fakeLimitSource{
		limits: map[enums.PlanKey]map[enums.CounterKind]accessservices.Limit{
			enums.PlanKeyFree: {
				enums.CounterKindAICredit:     {Enabled: true, Value: 15, Period: enums.UsagePeriodMonth},
				enums.CounterKindVoiceSession: {Enabled: true, Value: 3, Period: enums.UsagePeriodMonth},
			},
		},
	})
	counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
	events.EXPECT().WithTx(mock.Anything).Return(events).Once()
	events.EXPECT().ListByIdempotencyKeysForUpdate(mock.Anything, "user_1", mock.Anything).Return(nil, nil).Once()
	dupErr := &pgconn.PgError{Code: "23505", Message: `duplicate key value violates unique constraint "usage_events_user_id_idem"`}
	events.EXPECT().Insert(mock.Anything, mock.Anything).Return(dupErr).Once()

	_, err := svc.Reserve(context.Background(), "user_1", enums.PlanKeyFree, accessservices.ActionVoiceStart, voiceCharges(),
		"conversation:00000000-0000-0000-0000-0000000000f1", enums.UsageRefTypeConversation,
		"00000000-0000-0000-0000-0000000000f1", 2*time.Minute, fixedNow, time.UTC)
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 409, appErr.Status)
	require.Equal(t, "in_progress", appErr.Code)
}

// CommitTx settles one group over the caller's handle: no nested
// transaction, per-row counter commits through the joined repos.
func TestCommitTxSettlesGroup(t *testing.T) {
	tests := []struct {
		name    string
		settled []*models.UsageEvent
		wantErr bool
	}{
		{name: "settles each row through the joined counters"},
		{name: "empty settle is a silent no-op", settled: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counters := repomocks.NewMockUsageCounterRepository(t)
			events := repomocks.NewMockUsageEventRepository(t)
			svc := accessservices.NewUsageService(counters, events, fakeTxManager{}, fakeLimitSource{})
			settled := tt.settled
			if settled == nil && !tt.wantErr {
				settled = eventRows("00000000-0000-0000-0000-0000000000ba", "user_1", enums.UsageEventStatusReserved)
				if tt.name == "empty settle is a silent no-op" {
					settled = nil
				}
			}
			counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
			events.EXPECT().WithTx(mock.Anything).Return(events).Once()
			events.EXPECT().SettleReserved(mock.Anything, "00000000-0000-0000-0000-0000000000ba", enums.UsageEventStatusCommitted).
				Return(settled, nil).Once()
			for _, e := range settled {
				counters.EXPECT().Commit(mock.Anything, e.UserID, e.CounterKey, e.PeriodKey, e.Amount).Return(nil).Once()
			}
			require.NoError(t, svc.CommitTx(context.Background(), nil, "00000000-0000-0000-0000-0000000000ba"))
		})
	}
}

// The sweeper settles expiry atomically: expiry transition plus every
// counter release in ONE tx, count only on success.
func TestSweepExpiredIsAtomic(t *testing.T) {
	tests := []struct {
		name      string
		rows      []*models.UsageEvent
		releaseEr error
		want      int
		wantErr   int
	}{
		{
			name: "releases every expired hold in one tx",
			rows: eventRows("00000000-0000-0000-0000-0000000000cc", "user_1", enums.UsageEventStatusExpired),
			want: 2,
		},
		{
			name:      "counter failure rolls back the whole sweep",
			rows:      eventRows("00000000-0000-0000-0000-0000000000cc", "user_1", enums.UsageEventStatusExpired),
			releaseEr: errors.New("counter down"),
			wantErr:   503,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counters := repomocks.NewMockUsageCounterRepository(t)
			events := repomocks.NewMockUsageEventRepository(t)
			svc := accessservices.NewUsageService(counters, events, fakeTxManager{}, fakeLimitSource{})
			counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
			events.EXPECT().WithTx(mock.Anything).Return(events).Once()
			events.EXPECT().ReleaseExpired(mock.Anything, mock.Anything, mock.Anything).Return(tt.rows, nil).Once()
			if tt.releaseEr != nil {
				counters.EXPECT().Release(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(tt.releaseEr).Once()
			} else {
				for _, e := range tt.rows {
					counters.EXPECT().Release(mock.Anything, e.UserID, e.CounterKey, e.PeriodKey, e.Amount).Return(nil).Once()
				}
			}
			got, err := svc.SweepExpired(context.Background())
			if tt.wantErr != 0 {
				appErr := &common.AppError{}
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, tt.wantErr, appErr.Status)
				require.Equal(t, 0, got, "a failed sweep reports no count")
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// RefundCommitted returns settled charges: committed events move back to
// released and each used amount is returned, in one tx. Nothing committed
// is a silent no-op.
func TestRefundCommitted(t *testing.T) {
	tests := []struct {
		name     string
		refunded []*models.UsageEvent
		wantErr  bool
	}{
		{
			name:     "committed rows refund events and counters",
			refunded: eventRows("00000000-0000-0000-0000-0000000000dd", "user_1", enums.UsageEventStatusCommitted),
		},
		{name: "nothing committed is a silent no-op", refunded: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counters := repomocks.NewMockUsageCounterRepository(t)
			events := repomocks.NewMockUsageEventRepository(t)
			svc := accessservices.NewUsageService(counters, events, fakeTxManager{}, fakeLimitSource{})
			counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
			events.EXPECT().WithTx(mock.Anything).Return(events).Once()
			events.EXPECT().RefundCommitted(mock.Anything, "user_1", []string{
				"conversation:00000000-0000-0000-0000-0000000000f6:ai_credit",
				"conversation:00000000-0000-0000-0000-0000000000f6:voice_session",
			}).Return(tt.refunded, nil).Once()
			for _, e := range tt.refunded {
				counters.EXPECT().Refund(mock.Anything, e.UserID, e.CounterKey, e.PeriodKey, e.Amount).Return(nil).Once()
			}
			require.NoError(t, svc.RefundCommitted(context.Background(), "user_1", "conversation:00000000-0000-0000-0000-0000000000f6"))
		})
	}
}

// TestReserveWritesActionOnEvents pins the Task 3 ledger contract: freshly
// inserted event rows carry the reserving action.
func TestReserveWritesActionOnEvents(t *testing.T) {
	fixedNow := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)
	counters := repomocks.NewMockUsageCounterRepository(t)
	events := repomocks.NewMockUsageEventRepository(t)
	svc := accessservices.NewUsageService(counters, events, fakeTxManager{}, fakeLimitSource{
		limits: map[enums.PlanKey]map[enums.CounterKind]accessservices.Limit{
			enums.PlanKeyFree: {
				enums.CounterKindAICredit:     {Enabled: true, Value: 15, Period: enums.UsagePeriodMonth},
				enums.CounterKindVoiceSession: {Enabled: true, Value: 3, Period: enums.UsagePeriodMonth},
			},
		},
	})
	counters.EXPECT().WithTx(mock.Anything).Return(counters).Once()
	events.EXPECT().WithTx(mock.Anything).Return(events).Once()
	events.EXPECT().ListByIdempotencyKeysForUpdate(mock.Anything, "user_1", mock.Anything).Return(nil, nil).Once()
	events.EXPECT().Insert(mock.Anything, mock.MatchedBy(func(e *models.UsageEvent) bool {
		return e.Action == enums.AccessActionVoiceStart
	})).Return(nil).Twice()
	counters.EXPECT().Reserve(mock.Anything, "user_1", enums.CounterKindAICredit, "2026-11", 1, 15).Return(true, nil).Once()
	counters.EXPECT().Reserve(mock.Anything, "user_1", enums.CounterKindVoiceSession, "2026-11", 1, 3).Return(true, nil).Once()

	_, err := svc.Reserve(context.Background(), "user_1", enums.PlanKeyFree, accessservices.ActionVoiceStart, voiceCharges(),
		"conversation:00000000-0000-0000-0000-0000000000f1", enums.UsageRefTypeConversation,
		"00000000-0000-0000-0000-0000000000f1", 2*time.Minute, fixedNow, time.UTC)
	require.NoError(t, err)
}
