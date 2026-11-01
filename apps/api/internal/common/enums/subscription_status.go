package enums

// SubscriptionStatus is the premium-subscription lifecycle. A row starts
// pending (checkout created, payment not yet confirmed) and moves to
// active on a successful VNPay IPN, or failed if the payment didn't go
// through. expired/cancelled are reached later (expiry sweep, support
// action) — nothing in this pass sets them yet.
type SubscriptionStatus string

const (
	SubscriptionStatusPending   SubscriptionStatus = "pending"
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
	SubscriptionStatusFailed    SubscriptionStatus = "failed"
)
