package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// UserVocabulary is the per-user state over a shared vocabulary item:
// where it came from, the personal note, mastery, and the SRS due date
// (indexed for "due today" queries).
type UserVocabulary struct {
	UserID         string                 `gorm:"type:text;primaryKey" json:"userId"`
	ItemID         string                 `gorm:"type:uuid;primaryKey" json:"itemId"`
	SourceType     enums.VocabularySource `gorm:"column:source_type;not null" json:"sourceType"`
	SourceID       *string                `gorm:"type:uuid" json:"sourceId"`
	Note           *string                `json:"note"`
	Mastered       bool                   `gorm:"not null;default:false" json:"mastered"`
	SRSDueAt       *time.Time             `gorm:"column:srs_due_at" json:"srsDueAt"`
	IntervalDays   int                    `gorm:"column:interval_days;not null;default:0" json:"intervalDays"`
	ReviewCount    int                    `gorm:"column:review_count;not null;default:0" json:"reviewCount"`
	LastReviewedAt *time.Time             `gorm:"column:last_reviewed_at" json:"lastReviewedAt"`
	CreatedAt      time.Time              `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time              `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (UserVocabulary) TableName() string { return "user_vocabulary" }
