package requests

import "time"

// GrantSubscription manually grants premium to a user (admin-only). ExpiresAt
// nil means open-ended.
type GrantSubscription struct {
	Plan      string     `json:"plan" binding:"required"`
	ExpiresAt *time.Time `json:"expiresAt"`
}
