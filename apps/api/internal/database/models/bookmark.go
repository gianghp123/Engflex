package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// Bookmark is the polymorphic "Save for later" state for lessons, stories, or videos.
type Bookmark struct {
	UserID      string            `gorm:"type:text;primaryKey" json:"userId"`
	ContentType enums.ContentType `gorm:"type:text;primaryKey" json:"contentType"`
	ContentID   string            `gorm:"type:uuid;primaryKey" json:"contentId"`
	Note        *string           `json:"note"`
	CreatedAt   time.Time         `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time         `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Bookmark) TableName() string { return "bookmarks" }
