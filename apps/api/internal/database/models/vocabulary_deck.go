package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// VocabularyDeck is a flashcard deck: either personal (IsDefault false,
// owned by UserID, managed by that user) or system-curated (IsDefault true,
// UserID is the admin who authored it, managed by admin only, gated by
// Status before learners other than the author can see it).
type VocabularyDeck struct {
	ID           string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string              `gorm:"not null" json:"userId"`
	CategoryID   *string             `gorm:"type:uuid" json:"categoryId"`
	Name         string              `gorm:"not null" json:"name"`
	Description  string              `gorm:"not null;default:''" json:"description"`
	ThumbnailURL string              `json:"thumbnailUrl"`
	Level        string              `json:"level"`
	IsDefault    bool                `gorm:"default:false" json:"isDefault"`
	Status       enums.PublishStatus `gorm:"not null;default:draft" json:"status"`
	CreatedAt    time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`
	Category     *VocabularyCategory `gorm:"->;foreignKey:CategoryID;references:ID" json:"category,omitempty"`
}

func (VocabularyDeck) TableName() string { return "vocabulary_decks" }
