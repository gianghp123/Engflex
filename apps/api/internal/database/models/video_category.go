package models

import "time"

// VideoCategory is the content-format taxonomy of video_exercises
// ("trailer", "podcast", ...). Author-owned.
type VideoCategory struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Slug      string    `gorm:"not null" json:"slug"`
	Name      string    `gorm:"not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
	VideoCount int64	`gorm:"->" json:"videoCount"`
	Videos     []*VideoExercise `gorm:"->;foreignKey:CategoryID;references:ID" json:"videos,omitempty"`
}

func (VideoCategory) TableName() string { return "video_categories" }
