//go:build integration

// Real-Postgres verification (spec §14, and the repo rule that the emitted
// SQL must run against a real database): the atomic limit guard under
// concurrency, never-negative counters, sweeper expiry, and period keys.
// Run with: DATABASE_URL=... go test -tags=integration ./internal/modules/access/... -count=1 -v
package tests

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/repositories"
	accessservices "engflex-api/internal/modules/access/services"
)

func newIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set; usage integration suite requires a migrated Postgres")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	return db
}

func uniqueUser(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("it_user_%d", time.Now().UnixNano())
}

type staticLimits struct{}

func (staticLimits) LimitsForPlan(ctx context.Context, plan enums.PlanKey) (map[enums.CounterKind]accessservices.Limit, error) {
	return map[enums.CounterKind]accessservices.Limit{
		enums.CounterKindAICredit:     {Enabled: true, Value: 1, Period: enums.UsagePeriodMonth},
		enums.CounterKindVoiceSession: {Enabled: true, Value: 1, Period: enums.UsagePeriodMonth},
	}, nil
}

func TestReserveAtomicUnderConcurrency(t *testing.T) {
	db := newIntegrationDB(t)
	svc := accessservices.NewUsageService(
		repositories.NewUsageCounterRepository(db),
		repositories.NewUsageEventRepository(db),
		repositories.NewTransactionManager(db),
		staticLimits{},
	)
	user := uniqueUser(t)
	ctx := context.Background()

	// Keep the stress meaningful without exhausting the shared dev Postgres
	// connection limit: 100 concurrent transactions can produce fail-closed
	// 503s ("too many clients") unrelated to the atomicity under test.
	const goroutines = 20
	start := make(chan struct{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed, denied429, denied409 := 0, 0, 0
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			_, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionExamReview,
				[]accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}},
				fmt.Sprintf("exam_review:%d", i), enums.UsageRefTypeExamAttempt, fmt.Sprintf("%d", i),
				2*time.Minute, time.Now().UTC(), time.UTC)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				allowed++
				return
			}
			appErr := &common.AppError{}
			if assert.ErrorAs(t, err, &appErr) {
				switch appErr.Status {
				case 429:
					denied429++
				case 409:
					denied409++
				default:
					t.Errorf("unexpected status %d: %s", appErr.Status, appErr.Message)
				}
			}
		}(i)
	}
	close(start)
	wg.Wait()

	assert.Equal(t, 1, allowed, "exactly one reservation may pass a limit of 1")
	assert.Equal(t, goroutines-1, denied429+denied409)

	type row struct {
		Used     int
		Reserved int
	}
	var r row
	require.NoError(t, db.Raw("SELECT used, reserved FROM usage_counters WHERE user_id = ? AND counter_key = 'ai_credit'", user).Scan(&r).Error)
	assert.Equal(t, 1, r.Used+r.Reserved, "exactly one unit is held")
	assert.GreaterOrEqual(t, r.Reserved, 0)
	assert.GreaterOrEqual(t, r.Used, 0, "counters never go negative")
}

func TestReserveReuseAfterReleaseAndExpiry(t *testing.T) {
	db := newIntegrationDB(t)
	svc := accessservices.NewUsageService(
		repositories.NewUsageCounterRepository(db),
		repositories.NewUsageEventRepository(db),
		repositories.NewTransactionManager(db),
		staticLimits{},
	)
	user := uniqueUser(t)
	ctx := context.Background()
	const base = "exam_review:reuse-test"

	// Reserve then release: the counter hold must vanish.
	res, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionExamReview,
		[]accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}},
		base, enums.UsageRefTypeExamAttempt, "reuse", 2*time.Minute, time.Now().UTC(), time.UTC)
	require.NoError(t, err)
	require.Equal(t, accessservices.ReservationStatusReserved, res.Status)
	require.NoError(t, svc.Release(ctx, res.ID))

	// Same key again: the row is reused, a fresh hold applies, and a new
	// reservation id is returned — no second event row is created.
	res2, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionExamReview,
		[]accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}},
		base, enums.UsageRefTypeExamAttempt, "reuse", 2*time.Minute, time.Now().UTC(), time.UTC)
	require.NoError(t, err)
	require.Equal(t, accessservices.ReservationStatusReserved, res2.Status)
	assert.NotEqual(t, res.ID, res2.ID, "reuse mints a new group id")

	require.NoError(t, svc.Commit(ctx, res2.ID))
	// Committed replays are free.
	res3, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionExamReview,
		[]accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}},
		base, enums.UsageRefTypeExamAttempt, "reuse", 2*time.Minute, time.Now().UTC(), time.UTC)
	require.NoError(t, err)
	assert.Equal(t, accessservices.ReservationStatusCommitted, res3.Status)

	var events int64
	require.NoError(t, db.Raw("SELECT count(*) FROM usage_events WHERE user_id = ? AND idempotency_key LIKE ?", user, base+":%").Scan(&events).Error)
	assert.Equal(t, int64(1), events, "reuse never duplicates audit rows")
}

func TestSweeperReleasesExpiredReservations(t *testing.T) {
	db := newIntegrationDB(t)
	svc := accessservices.NewUsageService(
		repositories.NewUsageCounterRepository(db),
		repositories.NewUsageEventRepository(db),
		repositories.NewTransactionManager(db),
		staticLimits{},
	)
	user := uniqueUser(t)
	ctx := context.Background()

	// Reserve with a lease already in the past: the sweeper must pick it up.
	res, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionExamReview,
		[]accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}},
		"exam_review:sweep-test", enums.UsageRefTypeExamAttempt, "sweep", -time.Minute, time.Now().UTC(), time.UTC)
	require.NoError(t, err)
	require.Equal(t, accessservices.ReservationStatusReserved, res.Status)

	released, err := svc.SweepExpired(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, released, 1)

	var reservedHolds int64
	require.NoError(t, db.Raw(
		"SELECT count(*) FROM usage_events WHERE user_id = ? AND status = 'reserved'", user).Scan(&reservedHolds).Error)
	assert.Equal(t, int64(0), reservedHolds)

	type row struct {
		Used     int
		Reserved int
	}
	var r row
	require.NoError(t, db.Raw(
		"SELECT used, reserved FROM usage_counters WHERE user_id = ? AND counter_key = 'ai_credit'", user).Scan(&r).Error)
	assert.Equal(t, 0, r.Reserved, "the expired hold is released")
	assert.Equal(t, 0, r.Used)
}

func TestReserveRejectsOverLimitOnFreshRow(t *testing.T) {
	db := newIntegrationDB(t)
	svc := accessservices.NewUsageService(
		repositories.NewUsageCounterRepository(db),
		repositories.NewUsageEventRepository(db),
		repositories.NewTransactionManager(db),
		staticLimits{},
	)
	user := uniqueUser(t)
	ctx := context.Background()

	// Fresh user, amount 2 against a limit of 1: the guarded fresh-insert
	// path must deny without creating a counter row (Task 6 fix).
	_, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionStudyPlan,
		[]accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 2}},
		"exam_review:fresh-over-limit", enums.UsageRefTypeExamAttempt, "fresh-over-limit",
		2*time.Minute, time.Now().UTC(), time.UTC)
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 429, appErr.Status)
	assert.Equal(t, common.CodeInsufficientCredits, appErr.Code)

	var held int64
	require.NoError(t, db.Raw(
		"SELECT count(*) FROM usage_counters WHERE user_id = ? AND reserved > 0", user).Scan(&held).Error)
	assert.Equal(t, int64(0), held, "no counter row may hold reserved units for a denied fresh reserve")

	var events int64
	require.NoError(t, db.Raw(
		"SELECT count(*) FROM usage_events WHERE user_id = ?", user).Scan(&events).Error)
	assert.Equal(t, int64(0), events, "the denied reserve's event insert rolls back with the tx")
}

// Same-base concurrency (spec §7.2 exactly-once): 20 goroutines racing one
// idempotency base elect exactly one Reserved holder; every loser reports
// in_progress (409). The ledger holds exactly 2 event rows (one per
// counter) and each counter holds exactly 1 unit.
func TestReserveSameBaseConcurrent(t *testing.T) {
	db := newIntegrationDB(t)
	svc := accessservices.NewUsageService(
		repositories.NewUsageCounterRepository(db),
		repositories.NewUsageEventRepository(db),
		repositories.NewTransactionManager(db),
		staticLimits{},
	)
	user := uniqueUser(t)
	ctx := context.Background()
	base := fmt.Sprintf("conversation:same-%d", time.Now().UnixNano())

	const goroutines = 20
	start := make(chan struct{})
	var wg sync.WaitGroup
	var mu sync.Mutex
	reserved, inProgress, other := 0, 0, 0
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			// voiceCharges reserves ai_credit + voice_session under the
			// shared base: a whole-group hold.
			_, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionVoiceStart,
				[]accessservices.Charge{
					{Counter: enums.CounterKindAICredit, Amount: 1},
					{Counter: enums.CounterKindVoiceSession, Amount: 1},
				},
				base, enums.UsageRefTypeConversation, "conv-same",
				2*time.Minute, time.Now().UTC(), time.UTC)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				reserved++
				return
			}
			appErr := &common.AppError{}
			if assert.ErrorAs(t, err, &appErr) && appErr.Status == 409 && appErr.Code == "in_progress" {
				inProgress++
				return
			}
			other++
		}()
	}
	close(start)
	wg.Wait()

	assert.Equal(t, 1, reserved, "exactly one holder wins the base")
	assert.Equal(t, goroutines-1, inProgress, "every loser replays in_progress")
	assert.Equal(t, 0, other, "no loser fails any other way")

	var events int64
	require.NoError(t, db.Raw("SELECT count(*) FROM usage_events WHERE user_id = ? AND idempotency_key LIKE ?", user, base+":%").Scan(&events).Error)
	assert.Equal(t, int64(2), events, "one audit row per counter, never duplicated")

	type row struct {
		Counter  string
		Used     int
		Reserved int
	}
	var rows []row
	require.NoError(t, db.Raw("SELECT counter_key AS counter, used, reserved FROM usage_counters WHERE user_id = ?", user).Scan(&rows).Error)
	require.Len(t, rows, 2)
	for _, r := range rows {
		assert.Equal(t, 1, r.Used+r.Reserved, "counter %s holds exactly one unit", r.Counter)
	}
}

// Engine-loss refund (spec §8.1): a committed group the engine forgot is
// refunded, and a fresh Reserve with the same base reuses the released rows
// (a new hold) without inserting new audit rows.
func TestRefundCommittedThenReuse(t *testing.T) {
	db := newIntegrationDB(t)
	svc := accessservices.NewUsageService(
		repositories.NewUsageCounterRepository(db),
		repositories.NewUsageEventRepository(db),
		repositories.NewTransactionManager(db),
		staticLimits{},
	)
	user := uniqueUser(t)
	ctx := context.Background()
	base := fmt.Sprintf("conversation:refund-%d", time.Now().UnixNano())
	charges := []accessservices.Charge{
		{Counter: enums.CounterKindAICredit, Amount: 1},
		{Counter: enums.CounterKindVoiceSession, Amount: 1},
	}

	res, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionVoiceStart, charges,
		base, enums.UsageRefTypeConversation, "conv-refund",
		2*time.Minute, time.Now().UTC(), time.UTC)
	require.NoError(t, err)
	require.Equal(t, accessservices.ReservationStatusReserved, res.Status)
	require.NoError(t, svc.Commit(ctx, res.ID))

	// The engine forgot the session: refund the settled charge.
	require.NoError(t, svc.RefundCommitted(ctx, user, base))

	var released int64
	require.NoError(t, db.Raw("SELECT count(*) FROM usage_events WHERE user_id = ? AND idempotency_key LIKE ? AND status = 'released'", user, base+":%").Scan(&released).Error)
	assert.Equal(t, int64(2), released, "both committed rows move back to released")

	// A fresh Reserve reuses the released rows under a new group id: a new
	// hold, no new audit rows.
	res2, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionVoiceStart, charges,
		base, enums.UsageRefTypeConversation, "conv-refund",
		2*time.Minute, time.Now().UTC(), time.UTC)
	require.NoError(t, err)
	require.Equal(t, accessservices.ReservationStatusReserved, res2.Status)
	assert.NotEqual(t, res.ID, res2.ID, "reuse mints a new group id")

	var events int64
	require.NoError(t, db.Raw("SELECT count(*) FROM usage_events WHERE user_id = ? AND idempotency_key LIKE ?", user, base+":%").Scan(&events).Error)
	assert.Equal(t, int64(2), events, "reuse never duplicates audit rows")

	type hold struct{ Reserved int }
	var holds []hold
	require.NoError(t, db.Raw("SELECT reserved FROM usage_counters WHERE user_id = ?", user).Scan(&holds).Error)
	require.Len(t, holds, 2)
	for _, h := range holds {
		assert.Equal(t, 1, h.Reserved, "the fresh hold is back on the counters")
	}
}

// generousLimits allows multi-credit holds so the reuse test can change the
// charge between release and retry (staticLimits caps at 1).
type generousLimits struct{}

func (generousLimits) LimitsForPlan(ctx context.Context, plan enums.PlanKey) (map[enums.CounterKind]accessservices.Limit, error) {
	return map[enums.CounterKind]accessservices.Limit{
		enums.CounterKindAICredit:     {Enabled: true, Value: 100, Period: enums.UsagePeriodMonth},
		enums.CounterKindVoiceSession: {Enabled: true, Value: 100, Period: enums.UsagePeriodMonth},
	}, nil
}

// A reused row adopts the current charge (Important 1): reserve amount 1,
// release, reserve the same key with amount 3 under a different action,
// commit. The ledger row must carry amount 3 and the new action, and the
// counter must settle 3 used with nothing stranded in reserved.
func TestReserveReuseAdoptsCurrentCharge(t *testing.T) {
	db := newIntegrationDB(t)
	svc := accessservices.NewUsageService(
		repositories.NewUsageCounterRepository(db),
		repositories.NewUsageEventRepository(db),
		repositories.NewTransactionManager(db),
		generousLimits{},
	)
	user := uniqueUser(t)
	ctx := context.Background()
	const base = "exam_review:reuse-charge"

	res, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionExamReview,
		[]accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}},
		base, enums.UsageRefTypeExamAttempt, "reuse-charge", 2*time.Minute, time.Now().UTC(), time.UTC)
	require.NoError(t, err)
	require.Equal(t, accessservices.ReservationStatusReserved, res.Status)
	require.NoError(t, svc.Release(ctx, res.ID))

	res2, err := svc.Reserve(ctx, user, enums.PlanKeyFree, accessservices.ActionStudyPlan,
		[]accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 3}},
		base, enums.UsageRefTypeExamAttempt, "reuse-charge", 2*time.Minute, time.Now().UTC(), time.UTC)
	require.NoError(t, err)
	require.Equal(t, accessservices.ReservationStatusReserved, res2.Status)
	require.NoError(t, svc.Commit(ctx, res2.ID))

	var amount int64
	require.NoError(t, db.Raw("SELECT amount FROM usage_events WHERE user_id = ? AND idempotency_key = ?",
		user, base+":ai_credit").Scan(&amount).Error)
	assert.Equal(t, int64(3), amount, "the reused row carries the current amount")
	var action string
	require.NoError(t, db.Raw("SELECT action FROM usage_events WHERE user_id = ? AND idempotency_key = ?",
		user, base+":ai_credit").Scan(&action).Error)
	assert.Equal(t, string(enums.AccessActionStudyPlan), action, "the reused row carries the current action")

	type row struct {
		Used     int
		Reserved int
	}
	var r row
	require.NoError(t, db.Raw(
		"SELECT used, reserved FROM usage_counters WHERE user_id = ? AND counter_key = 'ai_credit'", user).Scan(&r).Error)
	assert.Equal(t, 3, r.Used)
	assert.Equal(t, 0, r.Reserved, "no difference strands in reserved")
}
