package repositories

import (
	"context"
	"time"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// SubscriptionRepository is the DB access layer for the subscriptions table.
// "What access does this user currently have" is answered by GetActiveByUserID;
// the payment flow owns transitioning a row from pending to active/failed via
// UpdateStatus, keyed off the row's own ID (set by VNPayService at checkout time).
type SubscriptionRepository interface {
	Create(ctx context.Context, s *models.Subscription) error
	GetByID(ctx context.Context, id string) (*models.Subscription, error)
	GetActiveByUserID(ctx context.Context, userID string) (*models.Subscription, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Subscription, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
	// ListAll/CountAll back the admin-wide subscriptions view (all users).
	// status/plan filter when non-empty; both empty lists everything.
	ListAll(ctx context.Context, status enums.SubscriptionStatus, plan enums.SubscriptionPlan, limit, offset int) ([]*models.Subscription, error)
	CountAll(ctx context.Context, status enums.SubscriptionStatus, plan enums.SubscriptionPlan) (int64, error)
	UpdateStatus(ctx context.Context, id string, status enums.SubscriptionStatus, startedAt, expiresAt *time.Time) error
	WithTx(tx *gorm.DB) SubscriptionRepository
}

type subscriptionRepository struct {
	db *gorm.DB
}

// NewSubscriptionRepository builds the repository over the shared DB handle.
func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *subscriptionRepository) WithTx(tx *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: tx}
}

func (r *subscriptionRepository) Create(ctx context.Context, s *models.Subscription) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id string) (*models.Subscription, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var m models.Subscription
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// GetActiveByUserID returns the most recent subscription that is currently
// active and not yet expired, i.e. what IsPremium checks against. Returns
// gorm.ErrRecordNotFound when the user has no live subscription.
func (r *subscriptionRepository) GetActiveByUserID(ctx context.Context, userID string) (*models.Subscription, error) {
	var m models.Subscription
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND expires_at > now()", userID, enums.SubscriptionStatusActive).
		Order("expires_at DESC").
		First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *subscriptionRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Subscription, error) {
	var out []*models.Subscription
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *subscriptionRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&models.Subscription{}).Where("user_id = ?", userID).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

// scopeAll applies the optional status/plan filters shared by
// ListAll/CountAll.
func scopeAll(db *gorm.DB, status enums.SubscriptionStatus, plan enums.SubscriptionPlan) *gorm.DB {
	if status != "" {
		db = db.Where("status = ?", status)
	}
	if plan != "" {
		db = db.Where("plan = ?", plan)
	}
	return db
}

func (r *subscriptionRepository) ListAll(ctx context.Context, status enums.SubscriptionStatus, plan enums.SubscriptionPlan, limit, offset int) ([]*models.Subscription, error) {
	var out []*models.Subscription
	if err := scopeAll(r.db.WithContext(ctx), status, plan).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *subscriptionRepository) CountAll(ctx context.Context, status enums.SubscriptionStatus, plan enums.SubscriptionPlan) (int64, error) {
	var n int64
	if err := scopeAll(r.db.WithContext(ctx).Model(&models.Subscription{}), status, plan).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (r *subscriptionRepository) UpdateStatus(ctx context.Context, id string, status enums.SubscriptionStatus, startedAt, expiresAt *time.Time) error {
	updates := map[string]any{"status": status}
	if startedAt != nil {
		updates["started_at"] = *startedAt
	}
	if expiresAt != nil {
		updates["expires_at"] = *expiresAt
	}
	return r.db.WithContext(ctx).Model(&models.Subscription{}).Where("id = ?", id).Updates(updates).Error
}
