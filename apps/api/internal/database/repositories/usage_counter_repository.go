package repositories

import (
	"context"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// UsageCounterRepository owns the authoritative atomic meter updates
// (spec §7.1). Reserve is a single INSERT ... ON CONFLICT with a limit
// guard — no read-modify-write anywhere.
type UsageCounterRepository interface {
	// Reserve atomically adds amount to reserved, returning false when the
	// limit guard fails (used + reserved + amount > limit). One statement;
	// the guard covers BOTH paths: the INSERT...SELECT's WHERE rejects an
	// over-limit fresh row, and the ON CONFLICT WHERE rejects an over-limit
	// top-up. RowsAffected 0 therefore means denied on either path.
	Reserve(ctx context.Context, userID string, counter enums.CounterKind, periodKey string, amount, limit int) (bool, error)
	// Commit moves amount from reserved to used.
	Commit(ctx context.Context, userID string, counter enums.CounterKind, periodKey string, amount int) error
	// Release drops amount from reserved (guarded: reserved >= amount).
	Release(ctx context.Context, userID string, counter enums.CounterKind, periodKey string, amount int) error
	// ListByUser returns every counter row for the user (few rows; the
	// service filters by period key in Go).
	ListByUser(ctx context.Context, userID string) ([]*models.UsageCounter, error)
	WithTx(tx *gorm.DB) UsageCounterRepository
}

type usageCounterRepository struct{ db *gorm.DB }

func NewUsageCounterRepository(db *gorm.DB) UsageCounterRepository {
	return &usageCounterRepository{db: db}
}

func (r *usageCounterRepository) WithTx(tx *gorm.DB) UsageCounterRepository {
	return &usageCounterRepository{db: tx}
}

func (r *usageCounterRepository) Reserve(ctx context.Context, userID string, counter enums.CounterKind, periodKey string, amount, limit int) (bool, error) {
	res := r.db.WithContext(ctx).Exec(`
INSERT INTO usage_counters (user_id, counter_key, period_key, used, reserved, updated_at)
SELECT ?, ?, ?, 0, ?, now()
WHERE ?::int <= ?::int
ON CONFLICT (user_id, counter_key, period_key)
DO UPDATE SET reserved = usage_counters.reserved + EXCLUDED.reserved,
              updated_at = now()
WHERE usage_counters.used + usage_counters.reserved + EXCLUDED.reserved <= ?
RETURNING used, reserved`, userID, string(counter), periodKey, amount, amount, limit, limit)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (r *usageCounterRepository) Commit(ctx context.Context, userID string, counter enums.CounterKind, periodKey string, amount int) error {
	return r.db.WithContext(ctx).Exec(`
UPDATE usage_counters
SET used = used + ?, reserved = reserved - ?, updated_at = now()
WHERE user_id = ? AND counter_key = ? AND period_key = ? AND reserved >= ?`,
		amount, amount, userID, string(counter), periodKey, amount).Error
}

func (r *usageCounterRepository) Release(ctx context.Context, userID string, counter enums.CounterKind, periodKey string, amount int) error {
	return r.db.WithContext(ctx).Exec(`
UPDATE usage_counters
SET reserved = reserved - ?, updated_at = now()
WHERE user_id = ? AND counter_key = ? AND period_key = ? AND reserved >= ?`,
		amount, userID, string(counter), periodKey, amount).Error
}

func (r *usageCounterRepository) ListByUser(ctx context.Context, userID string) ([]*models.UsageCounter, error) {
	out := make([]*models.UsageCounter, 0)
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
