package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// ExamCategory groups tests by collection, such as ETS 2024, and exam type.
type ExamCategory struct {
	ID        string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug      string         `gorm:"type:text;not null;uniqueIndex" json:"slug"`
	Title     string         `gorm:"type:text;not null" json:"title"`
	Type      enums.ExamType `gorm:"type:text;not null" json:"type"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (ExamCategory) TableName() string { return "exam_categories" }

// ExamCategoryCatalog keeps the category identity with its live catalog counts.
// Aggregate columns belong to this read model, so ordinary relation joins never
// attempt to select them from the physical exam_categories table.
type ExamCategoryCatalog struct {
	ExamCategory  `gorm:"embedded"`
	TestsCount    int64 `gorm:"->"`
	AttemptsCount int64 `gorm:"->"`
}
