package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
)

// ReservationStatus reports what a Reserve call found or did.
type ReservationStatus string

const (
	ReservationStatusReserved   ReservationStatus = "reserved"
	ReservationStatusCommitted  ReservationStatus = "committed"
	ReservationStatusInProgress ReservationStatus = "in_progress"
)

// Reservation is the group handle a service holds between Reserve and
// Commit/Release. ID is the group id shared by every event row of the hold.
type Reservation struct {
	ID     string
	Status ReservationStatus
}

// LimitSource re-resolves limits at Reserve time so a plan change between
// the advisory Decide and the authoritative Reserve applies immediately
// (spec §7.1). PolicyService implements it structurally.
type LimitSource interface {
	LimitsForPlan(ctx context.Context, plan enums.PlanKey) (map[enums.CounterKind]Limit, error)
}

// UsageService owns the metering writes: Reserve/Commit/Release plus the
// expiry sweeper (spec §7). Constructed once and held by the access module.
type UsageService struct {
	counters repositories.UsageCounterRepository
	events   repositories.UsageEventRepository
	tx       repositories.TransactionManager
	limits   LimitSource
}

// NewUsageService builds the metering service. Never pass nil; a nil
// LimitSource must be caught by the caller (wiring) instead.
func NewUsageService(counters repositories.UsageCounterRepository, events repositories.UsageEventRepository, tx repositories.TransactionManager, limits LimitSource) *UsageService {
	return &UsageService{counters: counters, events: events, tx: tx, limits: limits}
}

// periodKeyFor cuts the usage period key in the configured location. The
// location is provided by the caller (PolicyService owns APP_TIMEZONE).
func periodKeyFor(period enums.UsagePeriod, now time.Time, loc *time.Location) string {
	local := now.In(loc)
	switch period {
	case enums.UsagePeriodDay:
		return local.Format("2006-01-02")
	case enums.UsagePeriodTotal:
		return "all"
	default:
		return local.Format("2006-01")
	}
}

// Reserve atomically holds every counter an action charges (spec §7.1).
// The unique (user_id, idempotency_key) rows are the exactly-once guard;
// committed rows replay as committed, unexpired holds report in_progress,
// and released/expired rows are reused — never re-inserted.
func (s *UsageService) Reserve(ctx context.Context, userID string, plan enums.PlanKey, action enums.AccessAction, charges []Charge, idempotencyBase string, refType enums.UsageRefType, refID string, lease time.Duration, now time.Time, loc *time.Location) (*Reservation, error) {
	limits, err := s.limits.LimitsForPlan(ctx, plan)
	if err != nil {
		appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
		logger.Report(ctx, "reserve limits unavailable", appErr, "userID", userID, "plan", string(plan))
		return nil, appErr
	}
	reservationID := NewReservationID()
	res := &Reservation{ID: reservationID, Status: ReservationStatusReserved}

	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		events := s.events.WithTx(tx)
		counters := s.counters.WithTx(tx)

		keys := make([]string, 0, len(charges))
		byKey := map[string]Charge{}
		for _, ch := range charges {
			key := idempotencyBase + ":" + string(ch.Counter)
			keys = append(keys, key)
			byKey[key] = ch
		}
		existing, err := events.ListByIdempotencyKeysForUpdate(ctx, userID, keys)
		if err != nil {
			return err
		}
		found := map[string]*models.UsageEvent{}
		for _, e := range existing {
			found[e.IdempotencyKey] = e
		}

		// Replay semantics (spec §7.2), decided on locked rows. A partial
		// group (some keys present, some absent) is an inconsistent state a
		// crash cannot produce through this path — fail closed, never mint
		// a hold around it.
		if len(found) > 0 && len(found) < len(keys) {
			return errInconsistentReservation
		}
		if len(found) == len(keys) && len(charges) > 0 {
			committed, reservedLive := 0, 0
			for _, e := range found {
				switch e.Status {
				case enums.UsageEventStatusCommitted:
					committed++
				case enums.UsageEventStatusReserved:
					if e.ExpiresAt == nil || e.ExpiresAt.After(now) {
						reservedLive++
					}
				}
			}
			if committed == len(keys) {
				res.ID = firstReservationID(found)
				res.Status = ReservationStatusCommitted
				return nil
			}
			if committed > 0 {
				// Some rows settled while siblings did not: the group is
				// torn, and settling around it would double-charge.
				return errInconsistentReservation
			}
			if reservedLive > 0 {
				appErr := common.Conflict("a request for this item is already in progress").WithCode("in_progress")
				return &inProgressError{appErr}
			}
		}

		expiresAt := now.Add(lease)
		for _, key := range keys {
			ch := byKey[key]
			if e, ok := found[key]; ok {
				// Released or expired: reuse the row under the new group,
				// rewriting the charge to the current action/amount so the
				// later CommitTx settles exactly what the counters reserved.
				// A row that did not move was settled concurrently — fail
				// closed rather than inserting a duplicate idempotency key.
				reused, err := events.ReuseAsReserved(ctx, e.ID, reservationID, action, ch.Amount, expiresAt)
				if err != nil {
					return err
				}
				if !reused {
					return errInconsistentReservation
				}
				continue
			}
			lim := limits[ch.Counter]
			if err := events.Insert(ctx, &models.UsageEvent{
				ReservationID:  reservationID,
				UserID:         userID,
				CounterKey:     ch.Counter,
				PeriodKey:      periodKeyFor(lim.Period, now, loc),
				Action:         action,
				Amount:         ch.Amount,
				Status:         enums.UsageEventStatusReserved,
				IdempotencyKey: key,
				RefType:        refType,
				RefID:          refID,
				ExpiresAt:      &expiresAt,
			}); err != nil {
				// A duplicate key means a concurrent Reserve won the same
				// idempotency key first: report in_progress (409), not a
				// store outage — the other holder owns the action.
				if isDuplicateKeyError(err) {
					appErr := common.Conflict("a request for this item is already in progress").WithCode("in_progress")
					return &inProgressError{appErr}
				}
				return err
			}
		}
		for _, key := range keys {
			ch := byKey[key]
			lim := limits[ch.Counter]
			ok, err := counters.Reserve(ctx, userID, ch.Counter, periodKeyFor(lim.Period, now, loc), ch.Amount, lim.Value)
			if err != nil {
				return err
			}
			if !ok {
				used, remaining := currentUsage(ctx, counters, userID, ch.Counter, lim, now, loc)
				appErr := common.TooManyRequests(limitMessageFor(ch.Counter)).
					WithCode(limitCodeFor(ch.Counter)).
					WithDetails(quotaDetails(ch, lim, used, remaining, resetAtFor(lim.Period, now, loc)))
				return &denyError{appErr}
			}
		}
		return nil
	})
	if err != nil {
		var inProgress *inProgressError
		if errors.As(err, &inProgress) {
			return &Reservation{ID: reservationID, Status: ReservationStatusInProgress}, inProgress.AppError
		}
		var deny *denyError
		if errors.As(err, &deny) {
			logger.Report(ctx, "reserve denied", deny.AppError, "userID", userID, "base", idempotencyBase)
			return nil, deny.AppError
		}
		appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
		logger.Report(ctx, "reserve failed", appErr, "userID", userID, "base", idempotencyBase, "cause", err.Error())
		return nil, appErr
	}
	slog.InfoContext(ctx, "usage reserved", "userID", userID, "reservationID", reservationID, "base", idempotencyBase)
	return res, nil
}

// Commit settles the whole group in one transaction: used += amount,
// reserved -= amount per event row. Idempotent — a second commit settles
// zero rows and touches nothing (spec §7.1).
func (s *UsageService) Commit(ctx context.Context, reservationID string) error {
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		return s.CommitTx(ctx, tx, reservationID)
	})
	if err != nil {
		appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
		logger.Report(ctx, "commit failed", appErr, "reservationID", reservationID, "cause", err.Error())
		return appErr
	}
	slog.InfoContext(ctx, "usage committed", "reservationID", reservationID)
	return nil
}

// CommitTx is Commit's transactional body over the caller's handle: settles
// the group's rows and commits the counters without opening a nested
// transaction. Consuming modules run it inside their own tx.Run together
// with the state write, so the output row and its charge commit atomically.
func (s *UsageService) CommitTx(ctx context.Context, tx *gorm.DB, reservationID string) error {
	events := s.events.WithTx(tx)
	counters := s.counters.WithTx(tx)
	rows, err := events.SettleReserved(ctx, reservationID, enums.UsageEventStatusCommitted)
	if err != nil {
		return err
	}
	for _, e := range rows {
		if err := counters.Commit(ctx, e.UserID, e.CounterKey, e.PeriodKey, e.Amount); err != nil {
			return err
		}
	}
	return nil
}

// Release returns the whole group's holds. Idempotent in the same way.
func (s *UsageService) Release(ctx context.Context, reservationID string) error {
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		events := s.events.WithTx(tx)
		counters := s.counters.WithTx(tx)
		rows, err := events.SettleReserved(ctx, reservationID, enums.UsageEventStatusReleased)
		if err != nil {
			return err
		}
		for _, e := range rows {
			if err := counters.Release(ctx, e.UserID, e.CounterKey, e.PeriodKey, e.Amount); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
		logger.Report(ctx, "release failed", appErr, "reservationID", reservationID, "cause", err.Error())
		return appErr
	}
	slog.InfoContext(ctx, "usage released", "reservationID", reservationID)
	return nil
}

// SweepExpired releases every expired hold (spec §7.3): crash recovery for
// reservations abandoned between Reserve and Commit/Release. The expiry
// transition and every counter release run in ONE transaction — the count
// is returned only when the whole batch settled.
func (s *UsageService) SweepExpired(ctx context.Context) (int, error) {
	var rows []*models.UsageEvent
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		rows, err = s.events.WithTx(tx).ReleaseExpired(ctx, time.Now().UTC(), common.SweeperBatch)
		if err != nil {
			return err
		}
		counters := s.counters.WithTx(tx)
		for _, e := range rows {
			if err := counters.Release(ctx, e.UserID, e.CounterKey, e.PeriodKey, e.Amount); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
		logger.Report(ctx, "sweep failed", appErr, "cause", err.Error())
		return 0, appErr
	}
	if len(rows) > 0 {
		slog.InfoContext(ctx, "swept expired reservations", "count", len(rows))
	}
	return len(rows), nil
}

// RefundCommitted returns settled charges for one idempotency base (spec
// §8.1 engine-loss refund): the engine forgot a session the meter already
// settled, so the committed events move back to released and each used
// amount is returned, in one transaction. A base with nothing committed is
// a silent no-op.
func (s *UsageService) RefundCommitted(ctx context.Context, userID, idempotencyBase string) error {
	keys := []string{
		idempotencyBase + ":" + string(enums.CounterKindAICredit),
		idempotencyBase + ":" + string(enums.CounterKindVoiceSession),
	}
	var refunded []*models.UsageEvent
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		refunded, err = s.events.WithTx(tx).RefundCommitted(ctx, userID, keys)
		if err != nil {
			return err
		}
		counters := s.counters.WithTx(tx)
		for _, e := range refunded {
			if err := counters.Refund(ctx, e.UserID, e.CounterKey, e.PeriodKey, e.Amount); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
		logger.Report(ctx, "refund committed failed", appErr, "userID", userID, "base", idempotencyBase, "cause", err.Error())
		return appErr
	}
	if len(refunded) > 0 {
		slog.InfoContext(ctx, "committed usage refunded", "userID", userID, "base", idempotencyBase, "count", len(refunded))
	}
	return nil
}

// StartSweeper runs SweepExpired on the house cadence until the context is
// cancelled. Started once from the access module wiring (process lifetime).
func (s *UsageService) StartSweeper(ctx context.Context) {
	ticker := time.NewTicker(common.SweeperInterval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if _, err := s.SweepExpired(ctx); err != nil {
					// Already logged by SweepExpired; the loop must survive.
					continue
				}
			}
		}
	}()
}

// inProgressError and denyError carry pre-built AppErrors through the tx
// boundary without them being mistaken for store failures.
// errInconsistentReservation fails the tx closed when the locked rows form a
// group no crash path can produce: partial keys, mixed committed/sibling
// states, or a reuse that moved nothing. The outer mapping turns it into a
// 503 policy_unavailable — paid AI fails closed (spec §12).
var errInconsistentReservation = errors.New("inconsistent reservation state")

type inProgressError struct{ *common.AppError }
type denyError struct{ *common.AppError }

func (e *inProgressError) Error() string { return e.AppError.Error() }
func (e *denyError) Error() string       { return e.AppError.Error() }

// isDuplicateKeyError reports a unique-violation (23505): a concurrent
// Reserve won the same idempotency key first, so the loser reports
// in_progress rather than a store outage.
func isDuplicateKeyError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func limitCodeFor(counter enums.CounterKind) string {
	if counter == enums.CounterKindAICredit {
		return common.CodeInsufficientCredits
	}
	return common.CodeQuotaExceeded
}

func limitMessageFor(counter enums.CounterKind) string {
	if counter == enums.CounterKindAICredit {
		return "insufficient AI credits"
	}
	return "voice session limit reached"
}

func quotaDetails(ch Charge, lim Limit, used, remaining int, resetAt *time.Time) map[string]any {
	var resetAtValue any
	if resetAt != nil {
		resetAtValue = resetAt.Format(time.RFC3339)
	}
	return map[string]any{
		"code":       limitCodeFor(ch.Counter),
		"feature":    string(ch.Counter),
		"required":   ch.Amount,
		"limit":      lim.Value,
		"used":       used,
		"remaining":  remaining,
		"resetAt":    resetAtValue,
		"upgradeUrl": "/payments/pricing",
	}
}

// resetAtFor is the first instant of the next period in loc — the same math
// PolicyService.ResetAt applies, so advisory (Decide) and authoritative
// (Reserve) denials report the same reset instant. Total periods never
// reset: nil.
func resetAtFor(period enums.UsagePeriod, now time.Time, loc *time.Location) *time.Time {
	if loc == nil {
		loc = time.UTC
	}
	local := now.In(loc)
	var next time.Time
	switch period {
	case enums.UsagePeriodDay:
		next = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
	case enums.UsagePeriodTotal:
		return nil
	default:
		next = time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0)
	}
	return &next
}

// currentUsage reads the live used+reserved total for one counter inside the
// Reserve transaction so the denial can report used/remaining. A read
// failure degrades to zeros — the denial itself (429) still stands.
func currentUsage(ctx context.Context, counters repositories.UsageCounterRepository, userID string, counter enums.CounterKind, lim Limit, now time.Time, loc *time.Location) (used, remaining int) {
	rows, err := counters.ListByUser(ctx, userID)
	if err != nil {
		return 0, 0
	}
	key := periodKeyFor(lim.Period, now, loc)
	for _, r := range rows {
		if r.CounterKey == counter && r.PeriodKey == key {
			used = r.Used + r.Reserved
		}
	}
	remaining = lim.Value - used
	if remaining < 0 {
		remaining = 0
	}
	return used, remaining
}

func firstReservationID(found map[string]*models.UsageEvent) string {
	for _, e := range found {
		return e.ReservationID
	}
	return ""
}
