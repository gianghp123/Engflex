package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
)

// SubscriptionRepository is the DB access layer for the subscriptions table.
//
// The read surface arrives with the module that needs it: the paywall
// middleware needs GetActiveByUserID, admin user detail needs ListByUserID,
// and manual grant needs Create. WithTx scopes multi-write flows later.
type SubscriptionRepository interface {
	Create(ctx context.Context, s *models.Subscription) error
	GetActiveByUserID(ctx context.Context, userID string) (*models.Subscription, error)
	ListByUserID(ctx context.Context, userID string) ([]*models.Subscription, error)
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

// GetActiveByUserID returns the caller's current premium grant: status
// active and either open-ended (expires_at NULL) or not yet expired. Most
// recently started wins when more than one row qualifies.
func (r *subscriptionRepository) GetActiveByUserID(ctx context.Context, userID string) (*models.Subscription, error) {
	var m models.Subscription
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND (expires_at IS NULL OR expires_at > ?)", userID, enums.SubscriptionStatusActive, time.Now()).
		Order("started_at DESC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *subscriptionRepository) ListByUserID(ctx context.Context, userID string) ([]*models.Subscription, error) {
	out := make([]*models.Subscription, 0)
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("started_at DESC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
