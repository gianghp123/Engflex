package models

import (
	"time"

	"engflex-api/internal/common/enums"
)

// UserProfile is keyed by the Clerk subject. ID remains a stable public
// identifier for compatibility; preferences stay nested in the read graph.
type UserProfile struct {
	ID                    string      `gorm:"type:uuid;uniqueIndex;default:gen_random_uuid()" json:"id"`
	UserID                string      `gorm:"column:clerk_user_id;primaryKey;type:text" json:"userId"`
	Preferences           Preferences `gorm:"embedded" json:"preferences"`
	OnboardingCompletedAt *time.Time  `json:"onboardingCompletedAt"`
	CreatedAt             time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt             time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (UserProfile) TableName() string { return "user_profiles" }

// Preferences embeds checked scalar/array columns while preserving the
// nested application and wire shape. CEFR is derived from Level.
type Preferences struct {
	Level              enums.Level        `gorm:"type:text;not null;default:beginner" json:"level"`
	Goals              GoalArray          `gorm:"type:text[];not null;default:'{}'" json:"goals"`
	DailyCommitmentMin int                `gorm:"not null;default:15" json:"dailyCommitmentMin"`
	Topics             InterestTopicArray `gorm:"type:text[];not null;default:'{}'" json:"topics"`
}
