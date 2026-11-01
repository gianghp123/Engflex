package enums

// PaymentTransactionStatus is one VNPay checkout attempt's own lifecycle —
// separate from SubscriptionStatus because a user can retry a failed or
// abandoned attempt, and the transaction row is the idempotency key for
// VNPay's IPN callback (one attempt, one vnp_TxnRef).
type PaymentTransactionStatus string

const (
	PaymentTransactionStatusPending   PaymentTransactionStatus = "pending"
	PaymentTransactionStatusSucceeded PaymentTransactionStatus = "succeeded"
	PaymentTransactionStatusFailed    PaymentTransactionStatus = "failed"
)
