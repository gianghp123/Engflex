package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

type Story struct {
	ID              string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug            string              `gorm:"type:text;not null;unique" json:"slug"`
	Title           string              `gorm:"type:text;not null" json:"title"`
	TitleVietnamese *string             `gorm:"type:text" json:"titleVietnamese"`
	CEFRLevel       *enums.CEFR         `gorm:"column:cefr_level" json:"cefrLevel"`
	Topic           string              `gorm:"type:text;not null;default:'general'" json:"topic"`
	ThumbnailURL    *string             `gorm:"type:text" json:"thumbnailUrl"`
	AudioURL        string              `gorm:"type:text;not null" json:"audioUrl"`
	Duration        float64             `gorm:"type:double precision;not null;default:0" json:"duration"`
	Status          enums.PublishStatus `gorm:"not null;default:draft" json:"status"`
	IsPremium       bool                `gorm:"not null;default:false" json:"isPremium"`
	CreatedAt       time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`

	// Sentences is the child timeline, loaded with stories. gorm:"->" marks it read-only.
	Sentences []*StorySentence `gorm:"->;foreignKey:StoryID;references:ID" json:"sentences,omitempty"`
}

func (Story) TableName() string { return "stories" }
