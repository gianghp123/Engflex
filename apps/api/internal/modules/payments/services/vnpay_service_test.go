package services

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
)

const testSecret = "test-secret"

func signedIPNQuery(t *testing.T, overrides map[string]string) url.Values {
	t.Helper()
	params := map[string]string{
		"vnp_TxnRef":        "ref-1",
		"vnp_Amount":        "9900000",
		"vnp_ResponseCode":  "00",
		"vnp_TransactionNo": "VNP123",
	}
	for k, v := range overrides {
		params[k] = v
	}
	hash := vnpaySign(params, testSecret)

	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	q.Set("vnp_SecureHash", hash)
	return q
}

func TestVNPayService_VerifyAndHandleIPN_BadSignature(t *testing.T) {
	subs := mocks.NewMockSubscriptionRepository(t)
	txns := mocks.NewMockPaymentTransactionRepository(t)
	plans := mocks.NewMockPlanRepository(t)
	svc := NewVNPayService(VNPayConfig{HashSecret: testSecret}, nil, subs, txns, plans)

	q := signedIPNQuery(t, nil)
	q.Set("vnp_Amount", "1") // tampered after signing

	code, err := svc.VerifyAndHandleIPN(context.Background(), q)

	require.NoError(t, err)
	assert.Equal(t, vnpayRspInvalidSignature, code)
}

func TestVNPayService_VerifyAndHandleIPN_UnknownTxnRef(t *testing.T) {
	subs := mocks.NewMockSubscriptionRepository(t)
	txns := mocks.NewMockPaymentTransactionRepository(t)
	plans := mocks.NewMockPlanRepository(t)
	txns.On("GetByProviderTxnRefForUpdate", mock.Anything, "vnpay", "ref-1").Return(nil, gorm.ErrRecordNotFound)
	svc := NewVNPayService(VNPayConfig{HashSecret: testSecret}, nil, subs, txns, plans)

	code, err := svc.VerifyAndHandleIPN(context.Background(), signedIPNQuery(t, nil))

	require.NoError(t, err)
	assert.Equal(t, vnpayRspOrderNotFound, code)
}

func TestVNPayService_VerifyAndHandleIPN_AlreadyProcessedIsNoOp(t *testing.T) {
	subs := mocks.NewMockSubscriptionRepository(t)
	txns := mocks.NewMockPaymentTransactionRepository(t)
	plans := mocks.NewMockPlanRepository(t)
	txns.On("GetByProviderTxnRefForUpdate", mock.Anything, "vnpay", "ref-1").
		Return(&models.PaymentTransaction{ID: "txn_1", SubscriptionID: "sub_1", Amount: 99000, Status: enums.PaymentTransactionStatusSucceeded}, nil)
	svc := NewVNPayService(VNPayConfig{HashSecret: testSecret}, nil, subs, txns, plans)

	code, err := svc.VerifyAndHandleIPN(context.Background(), signedIPNQuery(t, nil))

	require.NoError(t, err)
	assert.Equal(t, vnpayRspAlreadyConfirmed, code)
	txns.AssertNotCalled(t, "UpdateResult", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	subs.AssertNotCalled(t, "UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestVNPayService_VerifyAndHandleIPN_AmountMismatch(t *testing.T) {
	subs := mocks.NewMockSubscriptionRepository(t)
	txns := mocks.NewMockPaymentTransactionRepository(t)
	plans := mocks.NewMockPlanRepository(t)
	txns.On("GetByProviderTxnRefForUpdate", mock.Anything, "vnpay", "ref-1").
		Return(&models.PaymentTransaction{ID: "txn_1", SubscriptionID: "sub_1", Amount: 99000, Status: enums.PaymentTransactionStatusPending}, nil)
	svc := NewVNPayService(VNPayConfig{HashSecret: testSecret}, nil, subs, txns, plans)

	code, err := svc.VerifyAndHandleIPN(context.Background(), signedIPNQuery(t, map[string]string{"vnp_Amount": "1"}))

	require.NoError(t, err)
	assert.Equal(t, vnpayRspInvalidAmount, code)
}

func TestVNPayService_VerifyAndHandleIPN_Success(t *testing.T) {
	subs := mocks.NewMockSubscriptionRepository(t)
	txns := mocks.NewMockPaymentTransactionRepository(t)
	plans := mocks.NewMockPlanRepository(t)
	txns.On("GetByProviderTxnRefForUpdate", mock.Anything, "vnpay", "ref-1").
		Return(&models.PaymentTransaction{ID: "txn_1", SubscriptionID: "sub_1", Amount: 99000, Status: enums.PaymentTransactionStatusPending}, nil)
	txns.On("UpdateResult", mock.Anything, "txn_1", enums.PaymentTransactionStatusSucceeded, mock.Anything, mock.Anything, mock.AnythingOfType("datatypes.JSON")).
		Return(nil)
	subs.On("GetByID", mock.Anything, "sub_1").
		Return(&models.Subscription{ID: "sub_1", Plan: enums.SubscriptionPlanOneMonth}, nil)
	plans.On("GetByPlan", mock.Anything, enums.SubscriptionPlanOneMonth).
		Return(&models.Plan{Plan: enums.SubscriptionPlanOneMonth, Months: 1, Amount: 99_000, Currency: "VND", IsActive: true}, nil)
	subs.On("UpdateStatus", mock.Anything, "sub_1", enums.SubscriptionStatusActive, mock.Anything, mock.Anything).
		Return(nil)
	svc := NewVNPayService(VNPayConfig{HashSecret: testSecret}, nil, subs, txns, plans)

	code, err := svc.VerifyAndHandleIPN(context.Background(), signedIPNQuery(t, nil))

	require.NoError(t, err)
	assert.Equal(t, vnpayRspSuccess, code)
}

func TestVNPayService_VerifyAndHandleIPN_PaymentFailed(t *testing.T) {
	subs := mocks.NewMockSubscriptionRepository(t)
	txns := mocks.NewMockPaymentTransactionRepository(t)
	plans := mocks.NewMockPlanRepository(t)
	txns.On("GetByProviderTxnRefForUpdate", mock.Anything, "vnpay", "ref-1").
		Return(&models.PaymentTransaction{ID: "txn_1", SubscriptionID: "sub_1", Amount: 99000, Status: enums.PaymentTransactionStatusPending}, nil)
	txns.On("UpdateResult", mock.Anything, "txn_1", enums.PaymentTransactionStatusFailed, mock.Anything, mock.Anything, mock.AnythingOfType("datatypes.JSON")).
		Return(nil)
	subs.On("UpdateStatus", mock.Anything, "sub_1", enums.SubscriptionStatusFailed, mock.Anything, mock.Anything).Return(nil)
	svc := NewVNPayService(VNPayConfig{HashSecret: testSecret}, nil, subs, txns, plans)

	code, err := svc.VerifyAndHandleIPN(context.Background(), signedIPNQuery(t, map[string]string{"vnp_ResponseCode": "24"}))

	require.NoError(t, err)
	assert.Equal(t, vnpayRspSuccess, code, "VNPay ack is 00 once the callback is processed, regardless of payment outcome")
	subs.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
}

func TestVNPayService_BuildCheckoutURL(t *testing.T) {
	subs := mocks.NewMockSubscriptionRepository(t)
	txns := mocks.NewMockPaymentTransactionRepository(t)
	plans := mocks.NewMockPlanRepository(t)
	plans.On("GetByPlan", mock.Anything, enums.SubscriptionPlanOneMonth).
		Return(&models.Plan{Plan: enums.SubscriptionPlanOneMonth, Months: 1, Amount: 99_000, Currency: "VND", IsActive: true}, nil)
	subs.On("Create", mock.Anything, mock.MatchedBy(func(m *models.Subscription) bool {
		return m.UserID == "user_1" && m.Status == enums.SubscriptionStatusPending
	})).Run(func(args mock.Arguments) {
		args.Get(1).(*models.Subscription).ID = "sub_1"
	}).Return(nil)
	txns.On("Create", mock.Anything, mock.MatchedBy(func(m *models.PaymentTransaction) bool {
		return m.SubscriptionID == "sub_1" && m.Status == enums.PaymentTransactionStatusPending && m.ProviderTxnRef != ""
	})).Return(nil)
	svc := NewVNPayService(VNPayConfig{
		TmnCode:    "TMN01",
		HashSecret: testSecret,
		PayURL:     "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html",
		ReturnURL:  "http://localhost:3000/payments/vnpay/return",
	}, nil, subs, txns, plans)

	res, err := svc.BuildCheckoutURL(context.Background(), "user_1", enums.SubscriptionPlanOneMonth, "127.0.0.1")

	require.NoError(t, err)
	require.NotNil(t, res)
	assert.NotEmpty(t, res.TxnRef)
	assert.Contains(t, res.PaymentURL, "vnp_SecureHash=")
	assert.Contains(t, res.PaymentURL, "vnp_TmnCode=TMN01")
}

func TestVNPayService_BuildCheckoutURL_UnknownPlan(t *testing.T) {
	subs := mocks.NewMockSubscriptionRepository(t)
	txns := mocks.NewMockPaymentTransactionRepository(t)
	plans := mocks.NewMockPlanRepository(t)
	plans.On("GetByPlan", mock.Anything, enums.SubscriptionPlan("bogus")).
		Return(nil, gorm.ErrRecordNotFound)
	svc := NewVNPayService(VNPayConfig{HashSecret: testSecret}, nil, subs, txns, plans)

	_, err := svc.BuildCheckoutURL(context.Background(), "user_1", enums.SubscriptionPlan("bogus"), "127.0.0.1")

	require.Error(t, err)
	subs.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestVNPayService_BuildCheckoutURL_InactivePlan(t *testing.T) {
	subs := mocks.NewMockSubscriptionRepository(t)
	txns := mocks.NewMockPaymentTransactionRepository(t)
	plans := mocks.NewMockPlanRepository(t)
	plans.On("GetByPlan", mock.Anything, enums.SubscriptionPlanOneMonth).
		Return(&models.Plan{Plan: enums.SubscriptionPlanOneMonth, Months: 1, Amount: 99_000, Currency: "VND", IsActive: false}, nil)
	svc := NewVNPayService(VNPayConfig{HashSecret: testSecret}, nil, subs, txns, plans)

	_, err := svc.BuildCheckoutURL(context.Background(), "user_1", enums.SubscriptionPlanOneMonth, "127.0.0.1")

	require.Error(t, err)
	subs.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}
