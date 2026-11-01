package repositories

import (
	"context"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PaymentTransactionRepository is the DB access layer for the
// payment_transactions table. GetByProviderTxnRef is the lookup VNPay's IPN
// callback uses; provider+provider_txn_ref is unique, so the first
// transition away from pending makes a replay a no-op (service layer's job,
// not this repository's — it only reads/writes rows).
type PaymentTransactionRepository interface {
	Create(ctx context.Context, t *models.PaymentTransaction) error
	GetByProviderTxnRef(ctx context.Context, provider, txnRef string) (*models.PaymentTransaction, error)
	// GetByProviderTxnRefForUpdate is GetByProviderTxnRef with a row lock
	// (SELECT ... FOR UPDATE). The IPN handler calls this one inside its
	// transaction so two concurrent deliveries of the same vnp_TxnRef (VNPay
	// retries IPN until it gets an ack) serialize on this row instead of
	// both reading "pending" and both applying the result.
	GetByProviderTxnRefForUpdate(ctx context.Context, provider, txnRef string) (*models.PaymentTransaction, error)
	ListBySubscriptionID(ctx context.Context, subscriptionID string) ([]*models.PaymentTransaction, error)
	UpdateResult(ctx context.Context, id string, status enums.PaymentTransactionStatus, providerTxnNo, responseCode *string, rawResponse datatypes.JSON) error
	WithTx(tx *gorm.DB) PaymentTransactionRepository
}

type paymentTransactionRepository struct {
	db *gorm.DB
}

// NewPaymentTransactionRepository builds the repository over the shared DB handle.
func NewPaymentTransactionRepository(db *gorm.DB) PaymentTransactionRepository {
	return &paymentTransactionRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *paymentTransactionRepository) WithTx(tx *gorm.DB) PaymentTransactionRepository {
	return &paymentTransactionRepository{db: tx}
}

func (r *paymentTransactionRepository) Create(ctx context.Context, t *models.PaymentTransaction) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *paymentTransactionRepository) GetByProviderTxnRef(ctx context.Context, provider, txnRef string) (*models.PaymentTransaction, error) {
	var m models.PaymentTransaction
	if err := r.db.WithContext(ctx).First(&m, "provider = ? AND provider_txn_ref = ?", provider, txnRef).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *paymentTransactionRepository) GetByProviderTxnRefForUpdate(ctx context.Context, provider, txnRef string) (*models.PaymentTransaction, error) {
	var m models.PaymentTransaction
	if err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&m, "provider = ? AND provider_txn_ref = ?", provider, txnRef).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *paymentTransactionRepository) ListBySubscriptionID(ctx context.Context, subscriptionID string) ([]*models.PaymentTransaction, error) {
	var out []*models.PaymentTransaction
	if err := r.db.WithContext(ctx).Where("subscription_id = ?", subscriptionID).
		Order("created_at DESC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *paymentTransactionRepository) UpdateResult(ctx context.Context, id string, status enums.PaymentTransactionStatus, providerTxnNo, responseCode *string, rawResponse datatypes.JSON) error {
	updates := map[string]any{
		"status":          status,
		"provider_txn_no": providerTxnNo,
		"response_code":   responseCode,
		"raw_response":    rawResponse,
	}
	return r.db.WithContext(ctx).Model(&models.PaymentTransaction{}).Where("id = ?", id).Updates(updates).Error
}
