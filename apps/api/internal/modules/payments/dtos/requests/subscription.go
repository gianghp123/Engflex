package requests

import "engflex-api/internal/common/enums"

// Checkout starts a VNPay payment for a fixed-duration plan.
type Checkout struct {
	Plan enums.SubscriptionPlan `json:"plan" binding:"required,oneof=1m 3m 12m"`
}

// GrantSubscription is the admin manual-grant path: no payment, straight
// to active. Note is optional context for support/audit, not stored
// structured anywhere beyond the log line.
type GrantSubscription struct {
	Plan enums.SubscriptionPlan `json:"plan" binding:"required,oneof=1m 3m 12m"`
	Note string                 `json:"note" binding:"omitempty,max=500"`
}

// UpdatePlan edits an existing fixed-duration package's price (admin-only).
// The plan code itself comes from the path, not the body — rows are fixed
// to the three seeded enums.SubscriptionPlan values, never created/deleted.
type UpdatePlan struct {
	Months   int    `json:"months" binding:"required,min=1,max=36"`
	Amount   int64  `json:"amount" binding:"required,min=0"`
	Currency string `json:"currency" binding:"required,len=3"`
	IsActive bool   `json:"isActive"`
}
