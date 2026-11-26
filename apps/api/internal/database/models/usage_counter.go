package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// UsageCounter is the authoritative metered state for one (user, counter,
// period): used = committed consumption, reserved = outstanding holds
// (spec §6.2). Never negative — CHECK-guarded at the database too.
type UsageCounter struct {
	UserID     string            `gorm:"primaryKey;type:text" json:"userId"`
	CounterKey enums.CounterKind `gorm:"primaryKey;type:text" json:"counterKey"`
	PeriodKey  string            `gorm:"primaryKey;type:text" json:"periodKey"`
	Used       int               `gorm:"not null;default:0" json:"used"`
	Reserved   int               `gorm:"not null;default:0" json:"reserved"`
	UpdatedAt  time.Time         `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (UsageCounter) TableName() string { return "usage_counters" }
