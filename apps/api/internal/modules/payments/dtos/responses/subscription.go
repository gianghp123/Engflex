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
	// Entitlement summary for the pricing table (spec §10): filled from the
	// plan_feature_limits catalog, not hardcoded, so admin edits flow through.
	Credits       int `json:"credits"`
	VoiceSessions int `json:"voiceSessions"`
}

// CheckoutSession is the VNPay hand-off: the learner's browser is
// redirected to PaymentURL to complete payment. Named distinctly from
// requests.Checkout (the plan-selection payload) so the two don't collide
// under the same generated TS name in the shared contracts barrel.
type CheckoutSession struct {
	PaymentURL string `json:"paymentUrl"`
	TxnRef     string `json:"txnRef"`
}

// SepayCheckoutSession is the SePay hand-off: no redirect — the frontend
// renders QRUrl inline and the learner pays by scanning it in their own
// banking app, then the frontend polls GetStatus(Reference) until it
// settles. Named distinctly from CheckoutSession so the two don't collide
// in the shared contracts barrel.
type SepayCheckoutSession struct {
	QRUrl             string `json:"qrUrl"`
	Reference         string `json:"reference"`
	Amount            int64  `json:"amount"`
	BankAccountNumber string `json:"bankAccountNumber"`
	BankShortName     string `json:"bankShortName"`
	AccountHolderName string `json:"accountHolderName"`
}

// SepayStatus is one poll result for a reference returned by
// SepayCheckoutSession. Status mirrors enums.PaymentTransactionStatus
// ("pending"/"succeeded"/"failed") as a plain string rather than the typed
// enum — the frontend only branches on the three literal values, not on
// package-level semantics.
type SepayStatus struct {
	Status    string     `json:"status"`
	ExpiresAt *time.Time `json:"expiresAt"`
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
