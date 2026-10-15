package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

// Preferences mirrors user_profiles.preferences (jsonb). The CEFR range is
// derived from Level in Go, not stored.
type Preferences struct {
	Level              enums.Level           `json:"level"`
	Goals              []enums.Goal          `json:"goals"`
	DailyCommitmentMin int                   `json:"dailyCommitmentMin"`
	Topics             []enums.InterestTopic `json:"topics"`
}

// Profile is the onboarding profile of the authenticated user.
type Profile struct {
	ID                    string      `json:"id"`
	UserID                string      `json:"userId"`
	Preferences           Preferences `json:"preferences"`
	OnboardingCompletedAt *time.Time  `json:"onboardingCompletedAt"`
	CreatedAt             time.Time   `json:"createdAt"`
	UpdatedAt             time.Time   `json:"updatedAt"`
}
