package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

// AdminUserSummary is one row in the admin user list — identity from Clerk
// plus the derived premium flag.
type AdminUserSummary struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Role      string    `json:"role"`
	Premium   bool      `json:"premium"`
	CreatedAt time.Time `json:"createdAt"`
}

// Subscription is one premium grant in a user's history.
type Subscription struct {
	ID        string                   `json:"id"`
	Plan      string                   `json:"plan"`
	Status    enums.SubscriptionStatus `json:"status"`
	StartedAt time.Time                `json:"startedAt"`
	ExpiresAt *time.Time               `json:"expiresAt"`
}

// AdminUserDetail is the full admin view of one user: Clerk identity,
// onboarding profile (nil if the user never onboarded), and subscription
// history.
type AdminUserDetail struct {
	AdminUserSummary
	Profile       *Profile       `json:"profile"`
	Subscriptions []Subscription `json:"subscriptions"`
}
