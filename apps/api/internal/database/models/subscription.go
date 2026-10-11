package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// Subscription is one premium grant for a user (manual admin grant or a
// future paid plan). The caller's current premium status is derived from
// whichever row is "active" and not yet expired, not stored redundantly
// elsewhere.
type Subscription struct {
	ID        string                   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string                   `gorm:"not null" json:"userId"`
	Plan      string                   `gorm:"not null" json:"plan"`
	Status    enums.SubscriptionStatus `gorm:"not null;default:active" json:"status"`
	StartedAt time.Time                `gorm:"not null;default:now()" json:"startedAt"`
	ExpiresAt *time.Time               `json:"expiresAt"`
	CreatedAt time.Time                `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time                `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Subscription) TableName() string { return "subscriptions" }
