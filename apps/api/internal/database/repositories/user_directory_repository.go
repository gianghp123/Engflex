package repositories

import (
	"context"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"gorm.io/gorm"
)

type UserDirectoryRepository interface {
	Enrich(ctx context.Context, users []*models.DirectoryUser) ([]*models.DirectoryUser, error)
}
type userDirectoryRepository struct{ db *gorm.DB }

func NewUserDirectoryRepository(db *gorm.DB) UserDirectoryRepository {
	return &userDirectoryRepository{db: db}
}

// One query for the entire Clerk page. Relation assembly stays in the
// repository; a database outage is surfaced, never rendered as no premium.
func (r *userDirectoryRepository) Enrich(ctx context.Context, users []*models.DirectoryUser) ([]*models.DirectoryUser, error) {
	if len(users) == 0 {
		return users, nil
	}
	ids := make([]string, 0, len(users))
	for _, user := range users {
		ids = append(ids, user.ID)
		user.Subscription = nil
	}
	var subscriptions []*models.Subscription
	if err := r.db.WithContext(ctx).Where("user_id IN ? AND status = ? AND expires_at > now()", ids, enums.SubscriptionStatusActive).Order("expires_at DESC, id DESC").Find(&subscriptions).Error; err != nil {
		return nil, err
	}
	active := make(map[string]*models.Subscription, len(subscriptions))
	for _, subscription := range subscriptions {
		if active[subscription.UserID] == nil {
			active[subscription.UserID] = subscription
		}
	}
	for _, user := range users {
		user.Subscription = active[user.ID]
	}
	return users, nil
}
