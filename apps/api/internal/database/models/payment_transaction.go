package models

import (
	"time"

	"engflex-api/internal/common/enums"

	"gorm.io/datatypes"
)

// PaymentTransaction is one VNPay checkout attempt. It is the audit trail
// and the idempotency key for VNPay's IPN callback: a replayed
// provider_txn_ref is a no-op once this row is already succeeded/failed.
type PaymentTransaction struct {
	ID             string                         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SubscriptionID string                         `gorm:"type:uuid;not null;index" json:"subscriptionId"`
	Provider       string                         `gorm:"not null;default:vnpay" json:"provider"`
	ProviderTxnRef string                         `gorm:"not null;uniqueIndex:idx_payment_transactions_provider_ref" json:"providerTxnRef"`
	ProviderTxnNo  *string                        `json:"providerTxnNo"`
	Amount         int64                          `gorm:"not null" json:"amount"`
	Status         enums.PaymentTransactionStatus `gorm:"not null;default:pending" json:"status"`
	ResponseCode   *string                        `json:"responseCode"`
	RawResponse    datatypes.JSON                 `gorm:"type:jsonb" json:"rawResponse"`
	CreatedAt      time.Time                      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time                      `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (PaymentTransaction) TableName() string { return "payment_transactions" }
