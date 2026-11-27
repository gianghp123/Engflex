package models

import "time"

// TrialSetting is the single-row trial duration setting. Trial caps stay in
// plan_feature_limits so enforcement reads one catalog.
type TrialSetting struct {
	ID           bool      `gorm:"primaryKey;default:true" json:"id"`
	DurationDays int       `gorm:"not null;default:7" json:"durationDays"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	UpdatedBy    string    `gorm:"not null;default:system" json:"updatedBy"`
}

func (TrialSetting) TableName() string { return "trial_settings" }
