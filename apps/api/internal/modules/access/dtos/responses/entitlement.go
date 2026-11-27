// Package responses holds the access module's wire DTOs (tygo inputs).
package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

// CounterView is one metered resource's state for the UI (spec §10).
// ResetAt is the first instant the period rolls over, RFC3339; omitted for
// never-resetting total periods (the UI shows "no reset" instead).
type CounterView struct {
	Limit     int        `json:"limit"`
	Used      int        `json:"used"`
	Reserved  int        `json:"reserved"`
	Remaining int        `json:"remaining"`
	ResetAt   *time.Time `json:"resetAt,omitempty"`
}

// TrialView reports the one-per-user trial state; null when never started.
type TrialView struct {
	Active         bool      `json:"active"`
	EndsAt         time.Time `json:"endsAt"`
	CreditsGranted int       `json:"creditsGranted"`
}

// ActionCost is one metered action's price for confirm dialogs and pricing.
type ActionCost struct {
	Action  enums.AccessAction `json:"action"`
	Credits int                `json:"credits"`
}

// MyEntitlements is the GET /premium/me body (spec §10). IsPremium is true
// only for a paid active subscription; a trial resolves to plan "trial" and
// is reported separately.
type MyEntitlements struct {
	IsPremium bool          `json:"isPremium"`
	Plan      enums.PlanKey `json:"plan"`
	Trial     *TrialView    `json:"trial"`
	ExpiresAt *time.Time    `json:"expiresAt"`
	Credits   CounterView   `json:"credits"`
	Voice     CounterView   `json:"voice"`
	Costs     []ActionCost  `json:"costs"`
}
