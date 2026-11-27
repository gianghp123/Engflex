package repositories

import (
	"context"
	"time"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UsageEventRepository owns the audit ledger's writes and the reservation
// lookups (spec §7.1-7.2). All methods here run inside the caller's
// transaction during Reserve/Commit/Release; ListByIdempotencyKeysForUpdate
// takes row locks so same-key reserves serialize.
type UsageEventRepository interface {
	ListByIdempotencyKeysForUpdate(ctx context.Context, userID string, keys []string) ([]*models.UsageEvent, error)
	Insert(ctx context.Context, m *models.UsageEvent) error
	// ReuseAsReserved transitions a released/expired row back to reserved
	// under a new reservation id with a fresh lease (spec §7.2 reuse path).
	// It also rewrites action/amount to the current charge: an admin price
	// change between release and retry must not strand the difference in
	// reserved when CommitTx later settles the row's amount.
	// It reports whether exactly one row moved: false means the row was not
	// in a reusable state (already settled or concurrently reused) and the
	// caller must fail closed, never insert a duplicate.
	ReuseAsReserved(ctx context.Context, id, reservationID string, action enums.AccessAction, amount int, expiresAt time.Time) (bool, error)
	// SettleReserved transitions every reserved row of one reservation to
	// the target status, returning the settled rows so the caller adjusts
	// counters per (counter, period, amount). There is intentionally no
	// expires_at predicate: the lease check lives in the Reserve replay
	// semantics (via ListByIdempotencyKeysForUpdate) and the sweeper owns
	// expiry through ReleaseExpired.
	SettleReserved(ctx context.Context, reservationID string, status enums.UsageEventStatus) ([]*models.UsageEvent, error)
	// ReleaseExpired transitions expired reserved rows to expired (SKIP
	// LOCKED, batched), returning them for the counter decrement.
	ReleaseExpired(ctx context.Context, before time.Time, limit int) ([]*models.UsageEvent, error)
	// HasCommitted reports whether any of the keys has a committed event —
	// the AnalyzeTurn "was this session paid" guard.
	HasCommitted(ctx context.Context, userID string, keys []string) (bool, error)
	// RefundCommitted releases committed rows for the given idempotency keys
	// (spec §8.1 engine-loss refund): the engine forgot a session the meter
	// already settled, so the charge is returned without deleting the audit
	// rows. Returns the refunded rows so the caller can counter-refund each.
	RefundCommitted(ctx context.Context, userID string, keys []string) ([]*models.UsageEvent, error)
	// ListByUser returns the most recent events for one user, newest first
	// (constant query backing the admin usage lookup).
	ListByUser(ctx context.Context, userID string, limit int) ([]*models.UsageEvent, error)
	// ListByUserStatus returns one page of a user's events in the given
	// statuses, newest first (backing GET /premium/usage history).
	ListByUserStatus(ctx context.Context, userID string, statuses []enums.UsageEventStatus, limit, offset int) ([]*models.UsageEvent, error)
	// CountByUserStatus counts a user's events in the given statuses
	// (the history page total).
	CountByUserStatus(ctx context.Context, userID string, statuses []enums.UsageEventStatus) (int64, error)
	WithTx(tx *gorm.DB) UsageEventRepository
}

type usageEventRepository struct{ db *gorm.DB }

func NewUsageEventRepository(db *gorm.DB) UsageEventRepository {
	return &usageEventRepository{db: db}
}

func (r *usageEventRepository) WithTx(tx *gorm.DB) UsageEventRepository {
	return &usageEventRepository{db: tx}
}

func (r *usageEventRepository) ListByIdempotencyKeysForUpdate(ctx context.Context, userID string, keys []string) ([]*models.UsageEvent, error) {
	out := make([]*models.UsageEvent, 0)
	if len(keys) == 0 {
		return out, nil
	}
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND idempotency_key IN ?", userID, keys).
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *usageEventRepository) Insert(ctx context.Context, m *models.UsageEvent) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *usageEventRepository) ReuseAsReserved(ctx context.Context, id, reservationID string, action enums.AccessAction, amount int, expiresAt time.Time) (bool, error) {
	res := r.db.WithContext(ctx).Exec(`
UPDATE usage_events
SET status = 'reserved', action = ?, amount = ?, reservation_id = ?, expires_at = ?, settled_at = NULL
WHERE id = ? AND status IN ('released', 'expired')`,
		string(action), amount, reservationID, expiresAt, id)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (r *usageEventRepository) SettleReserved(ctx context.Context, reservationID string, status enums.UsageEventStatus) ([]*models.UsageEvent, error) {
	out := make([]*models.UsageEvent, 0)
	err := r.db.WithContext(ctx).Raw(`
UPDATE usage_events
SET status = ?, settled_at = now()
WHERE reservation_id = ? AND status = 'reserved'
RETURNING *`, string(status), reservationID).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *usageEventRepository) ReleaseExpired(ctx context.Context, before time.Time, limit int) ([]*models.UsageEvent, error) {
	out := make([]*models.UsageEvent, 0)
	err := r.db.WithContext(ctx).Raw(`
UPDATE usage_events
SET status = 'expired', settled_at = now()
WHERE id IN (
    SELECT id FROM usage_events
    WHERE status = 'reserved' AND expires_at < ?
    ORDER BY expires_at
    LIMIT ?
    FOR UPDATE SKIP LOCKED
)
RETURNING *`, before, limit).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *usageEventRepository) HasCommitted(ctx context.Context, userID string, keys []string) (bool, error) {
	if len(keys) == 0 {
		return false, nil
	}
	var n int64
	err := r.db.WithContext(ctx).Model(&models.UsageEvent{}).
		Where("user_id = ? AND idempotency_key IN ? AND status = ?", userID, keys, enums.UsageEventStatusCommitted).
		Count(&n).Error
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *usageEventRepository) RefundCommitted(ctx context.Context, userID string, keys []string) ([]*models.UsageEvent, error) {
	out := make([]*models.UsageEvent, 0)
	if len(keys) == 0 {
		return out, nil
	}
	err := r.db.WithContext(ctx).Raw(`
UPDATE usage_events
SET status = 'released', settled_at = now()
WHERE user_id = ? AND idempotency_key IN ? AND status = 'committed'
RETURNING *`, userID, keys).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *usageEventRepository) ListByUser(ctx context.Context, userID string, limit int) ([]*models.UsageEvent, error) {
	out := make([]*models.UsageEvent, 0)
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *usageEventRepository) ListByUserStatus(ctx context.Context, userID string, statuses []enums.UsageEventStatus, limit, offset int) ([]*models.UsageEvent, error) {
	out := make([]*models.UsageEvent, 0)
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status IN ?", userID, statuses).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *usageEventRepository) CountByUserStatus(ctx context.Context, userID string, statuses []enums.UsageEventStatus) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.UsageEvent{}).
		Where("user_id = ? AND status IN ?", userID, statuses).
		Count(&n).Error
	if err != nil {
		return 0, err
	}
	return n, nil
}
