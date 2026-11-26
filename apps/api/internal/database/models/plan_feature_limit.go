package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// PlanFeatureLimit is one row of the entitlement catalog: how much of one
// counter a plan allows per period. Admin-editable; the policy snapshot
// caches these rows.
type PlanFeatureLimit struct {
	PlanKey    enums.PlanKey     `gorm:"primaryKey;type:text" json:"planKey"`
	CounterKey enums.CounterKind `gorm:"primaryKey;type:text" json:"counterKey"`
	IsEnabled  bool              `gorm:"not null;default:true" json:"isEnabled"`
	LimitValue int               `gorm:"not null;default:0" json:"limitValue"`
	Period     enums.UsagePeriod `gorm:"not null;default:month;type:text" json:"period"`
	CreatedAt  time.Time         `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt  time.Time         `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (PlanFeatureLimit) TableName() string { return "plan_feature_limits" }
