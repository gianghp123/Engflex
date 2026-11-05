package models

import (
	"time"

	"gorm.io/datatypes"

	"engflex-api/internal/exams/toeic"
)

// ExamTest represents a complete, ready-to-take examination test (e.g. TOEIC full test).
// Its question tree, audio/image assets, transcripts, and answer key live in Content (jsonb).
type ExamTest struct {
	ID            string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CategoryID    string         `gorm:"type:uuid;not null" json:"categoryId"`
	SourceKey     string         `gorm:"type:text;not null;uniqueIndex" json:"sourceKey"`
	Slug          string         `gorm:"type:text;not null;uniqueIndex" json:"slug"`
	Title         string         `gorm:"type:text;not null" json:"title"`
	Content       datatypes.JSON `gorm:"type:jsonb;not null" json:"content"`
	ParsedContent *toeic.Content `gorm:"-" json:"-"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`

	Category *ExamCategory `gorm:"->;foreignKey:CategoryID;references:ID" json:"category,omitempty"`
}

func (ExamTest) TableName() string { return "exam_tests" }

// ExamTestCatalog extends the same test graph with live catalog metadata.
// It is read directly by GORM; no layer rebuilds an ExamTest from a row shape.
type ExamTestCatalog struct {
	ExamTest        `gorm:"embedded"`
	AttemptsCount   int64 `gorm:"->"`
	DurationMinutes int   `gorm:"-"`
}
