package models

import "time"

// PremiumFeature is one admin-defined entry in the premium-gating registry:
// a named app capability (e.g. "voice_conversations") and whether it
// currently requires premium. This is a management/config surface only —
// no route in this codebase enforces it yet; wiring a specific feature's
// IsActive flag into its own handler is a separate, per-feature decision.
type PremiumFeature struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Key         string    `gorm:"not null;uniqueIndex" json:"key"`
	Label       string    `gorm:"not null" json:"label"`
	Description string    `gorm:"not null;default:''" json:"description"`
	IsActive    bool      `gorm:"not null;default:true" json:"isActive"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (PremiumFeature) TableName() string { return "premium_features" }
