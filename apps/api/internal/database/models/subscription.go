package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// Subscription is what access a user currently has: a fixed-duration premium
// package that starts pending and becomes active once a payment_transactions
// row succeeds. user_id (Clerk subject) has no FK, same convention as
// UserProfile.
type Subscription struct {
	ID        string                   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string                   `gorm:"not null;index:idx_subscriptions_user_created,priority:1;index:idx_subscriptions_user_status,priority:1" json:"userId"`
	Plan      enums.SubscriptionPlan   `gorm:"not null" json:"plan"`
	Status    enums.SubscriptionStatus `gorm:"not null;default:pending;index:idx_subscriptions_user_status,priority:2" json:"status"`
	Amount    int64                    `gorm:"not null" json:"amount"`
	Currency  string                   `gorm:"not null;default:VND" json:"currency"`
	StartedAt *time.Time               `json:"startedAt"`
	ExpiresAt *time.Time               `json:"expiresAt"`
	CreatedAt time.Time                `gorm:"autoCreateTime;index:idx_subscriptions_user_created,priority:2,sort:desc" json:"createdAt"`
	UpdatedAt time.Time                `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Subscription) TableName() string { return "subscriptions" }
