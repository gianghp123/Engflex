package models

import "time"

// UserTrial records the single 7-day trial a user may start. The user_id
// PK is the once-per-user guard; expiry is ends_at, after which the user
// resolves to the free plan.
type UserTrial struct {
	UserID         string    `gorm:"primaryKey;type:text" json:"userId"`
	StartedAt      time.Time `json:"startedAt"`
	EndsAt         time.Time `json:"endsAt"`
	CreditsGranted int       `json:"creditsGranted"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"createdAt"`
}

func (UserTrial) TableName() string { return "user_trials" }
