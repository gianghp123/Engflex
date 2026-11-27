package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// ActionCreditCost is the admin-editable AI-credit price of one paid action.
// Counter structure (e.g. voice also consumes a session slot) stays in code.
type ActionCreditCost struct {
	Action    enums.AccessAction `gorm:"primaryKey;type:text" json:"action"`
	Credits   int                `gorm:"not null;default:0" json:"credits"`
	UpdatedAt time.Time          `gorm:"autoUpdateTime" json:"updatedAt"`
	UpdatedBy string             `gorm:"not null;default:system" json:"updatedBy"`
}

func (ActionCreditCost) TableName() string { return "action_credit_costs" }
