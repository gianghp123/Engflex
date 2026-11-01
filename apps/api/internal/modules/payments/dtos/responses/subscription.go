package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

// Subscription is one premium-package row, learner- and admin-facing alike.
type Subscription struct {
	ID        string                   `json:"id"`
	UserID    string                   `json:"userId"`
	Plan      enums.SubscriptionPlan   `json:"plan"`
	Status    enums.SubscriptionStatus `json:"status"`
	Amount    int64                    `json:"amount"`
	Currency  string                   `json:"currency"`
	StartedAt *time.Time               `json:"startedAt"`
	ExpiresAt *time.Time               `json:"expiresAt"`
	CreatedAt time.Time                `json:"createdAt"`
	UpdatedAt time.Time                `json:"updatedAt"`
}

// Plan is one fixed-duration package's current price, admin-editable.
type Plan struct {
	Plan      enums.SubscriptionPlan `json:"plan"`
	Months    int                    `json:"months"`
	Amount    int64                  `json:"amount"`
	Currency  string                 `json:"currency"`
	IsActive  bool                   `json:"isActive"`
	UpdatedAt time.Time              `json:"updatedAt"`
}

// CheckoutSession is the VNPay hand-off: the learner's browser is
// redirected to PaymentURL to complete payment. Named distinctly from
// requests.Checkout (the plan-selection payload) so the two don't collide
// under the same generated TS name in the shared contracts barrel.
type CheckoutSession struct {
	PaymentURL string `json:"paymentUrl"`
	TxnRef     string `json:"txnRef"`
}

// MyPremiumStatus is the caller's own premium status — the one piece of
// subscription data exposed to a learner about themselves (everything else
// under /admin/premium/* is admin-only). The frontend uses this to decide
// whether a premium-gated card's action should proceed or redirect to
// pricing, instead of always navigating in and finding out via a 403.
type MyPremiumStatus struct {
	IsPremium bool       `json:"isPremium"`
	ExpiresAt *time.Time `json:"expiresAt"`
}
