package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// Plan is one fixed-duration premium package's current price. Admin-editable
// via the premium console; the row set itself is fixed to the three
// enums.SubscriptionPlan values seeded by migration — rows are updated,
// never created/deleted from the app.
type Plan struct {
	Plan      enums.SubscriptionPlan `gorm:"primaryKey" json:"plan"`
	Months    int                    `gorm:"not null" json:"months"`
	Amount    int64                  `gorm:"not null" json:"amount"`
	Currency  string                 `gorm:"not null;default:VND" json:"currency"`
	IsActive  bool                   `gorm:"not null;default:true" json:"isActive"`
	CreatedAt time.Time              `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time              `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Plan) TableName() string { return "plans" }
