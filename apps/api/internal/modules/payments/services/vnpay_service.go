package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/payments/dtos/responses"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// VNPay IPN response codes, per VNPay's integration spec. The ack body is
// {"RspCode": code, "Message": message} — not the house common.ApiResponse
// envelope, because VNPay's own server parses this reply, not a browser.
const (
	vnpayRspSuccess          = "00"
	vnpayRspOrderNotFound    = "01"
	vnpayRspAlreadyConfirmed = "02"
	vnpayRspInvalidAmount    = "04"
	vnpayRspInvalidSignature = "97"
	vnpayRspUnknownError     = "99"
)

var vnpayRspMessages = map[string]string{
	vnpayRspSuccess:          "Confirm Success",
	vnpayRspOrderNotFound:    "Order not found",
	vnpayRspAlreadyConfirmed: "Order already confirmed",
	vnpayRspInvalidAmount:    "Invalid amount",
	vnpayRspInvalidSignature: "Invalid signature",
	vnpayRspUnknownError:     "Unknown error",
}

// VNPayAckMessage maps a VNPay RspCode to its spec-mandated message text,
// for the controller to build the {"RspCode","Message"} IPN ack body.
func VNPayAckMessage(code string) string {
	if msg, ok := vnpayRspMessages[code]; ok {
		return msg
	}
	return vnpayRspMessages[vnpayRspUnknownError]
}

// VNPayConfig is the subset of config.PaymentConfig the service needs,
// narrowed here so this package doesn't import the top-level config
// package (which would be the only such cross-import in this module tree).
type VNPayConfig struct {
	TmnCode    string
	HashSecret string
	PayURL     string
	ReturnURL  string
}

// VNPayService builds VNPay checkout URLs and verifies/handles the IPN
// callback that confirms or fails a payment attempt.
type VNPayService interface {
	BuildCheckoutURL(ctx context.Context, userID string, plan enums.SubscriptionPlan, clientIP string) (*responses.CheckoutSession, error)
	VerifyAndHandleIPN(ctx context.Context, query url.Values) (rspCode string, err error)
}

type vnpayService struct {
	cfg   VNPayConfig
	db    *gorm.DB
	subs  repositories.SubscriptionRepository
	txns  repositories.PaymentTransactionRepository
	plans repositories.PlanRepository
}

// NewVNPayService builds the service over the repository interfaces and
// gateway config. db is used only to open the transactions BuildCheckoutURL
// and VerifyAndHandleIPN need across the subscriptions/payment_transactions
// tables — it may be nil (unit tests construct the service this way,
// exercising the mocked repositories directly without a live transaction);
// production wiring always supplies the real handle.
func NewVNPayService(cfg VNPayConfig, db *gorm.DB, subs repositories.SubscriptionRepository, txns repositories.PaymentTransactionRepository, plans repositories.PlanRepository) VNPayService {
	return &vnpayService{cfg: cfg, db: db, subs: subs, txns: txns, plans: plans}
}

// runInTx runs fn inside a real DB transaction when s.db is set, or calls
// fn(nil) directly otherwise. The two call sites below pass tx straight to
// WithTx on the repositories they use, so in the nil case those repos are
// left pointed at their original (non-transactional) handle — correct for
// unit tests, since the mocked repositories never look at tx.
func (s *vnpayService) runInTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if s.db == nil {
		return fn(nil)
	}
	return s.db.WithContext(ctx).Transaction(fn)
}

// generateTxnRef returns a time-prefixed random reference, unique enough
// for VNPay's per-merchant vnp_TxnRef requirement and the DB's own
// UNIQUE (provider, provider_txn_ref) constraint.
func generateTxnRef() string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return time.Now().Format("20060102150405") + hex.EncodeToString(buf)
}

// BuildCheckoutURL creates the pending subscription + payment_transaction
// rows and returns the signed VNPay checkout URL to redirect the learner's
// browser to.
func (s *vnpayService) BuildCheckoutURL(ctx context.Context, userID string, plan enums.SubscriptionPlan, clientIP string) (*responses.CheckoutSession, error) {
	def, err := s.plans.GetByPlan(ctx, plan)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.BadRequest("unknown plan")
		}
		appErr := common.FromDBError(err, "plan")
		logger.Report(ctx, "load plan for checkout failed", appErr, "plan", plan)
		return nil, appErr
	}
	if !def.IsActive {
		return nil, common.BadRequest("plan is not available")
	}

	sub := &models.Subscription{
		UserID:   userID,
		Plan:     plan,
		Status:   enums.SubscriptionStatusPending,
		Amount:   def.Amount,
		Currency: def.Currency,
	}
	txnRef := generateTxnRef()
	txn := &models.PaymentTransaction{
		Provider:       "vnpay",
		ProviderTxnRef: txnRef,
		Amount:         def.Amount,
		Status:         enums.PaymentTransactionStatusPending,
	}

	// The pending subscription and its payment_transaction must land
	// together or not at all — a checkout session with one row but not the
	// other is an inconsistent state nothing else in this flow expects.
	if err := s.runInTx(ctx, func(tx *gorm.DB) error {
		subs, txns := s.subs, s.txns
		if tx != nil {
			subs, txns = subs.WithTx(tx), txns.WithTx(tx)
		}
		if err := subs.Create(ctx, sub); err != nil {
			return err
		}
		txn.SubscriptionID = sub.ID
		return txns.Create(ctx, txn)
	}); err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "create checkout session failed", appErr, "user_id", userID)
		return nil, appErr
	}

	params := map[string]string{
		"vnp_Version":    "2.1.0",
		"vnp_Command":    "pay",
		"vnp_TmnCode":    s.cfg.TmnCode,
		"vnp_Amount":     strconv.FormatInt(def.Amount*100, 10),
		"vnp_CurrCode":   "VND",
		"vnp_TxnRef":     txnRef,
		"vnp_OrderInfo":  fmt.Sprintf("Thanh toan goi premium %s", plan),
		"vnp_OrderType":  "other",
		"vnp_Locale":     "vn",
		"vnp_ReturnUrl":  s.cfg.ReturnURL,
		"vnp_IpAddr":     clientIP,
		"vnp_CreateDate": time.Now().Format("20060102150405"),
	}
	secureHash := vnpaySign(params, s.cfg.HashSecret)

	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	q.Set("vnp_SecureHash", secureHash)

	paymentURL := s.cfg.PayURL + "?" + q.Encode()
	slog.InfoContext(ctx, "vnpay checkout created", "subscription_id", sub.ID, "txn_ref", txnRef, "plan", plan)
	return &responses.CheckoutSession{PaymentURL: paymentURL, TxnRef: txnRef}, nil
}

// VerifyAndHandleIPN verifies the callback signature, then applies the
// result to the matching payment_transaction + subscription rows. Replaying
// an already-processed vnp_TxnRef (same or different outcome) is a no-op —
// the first result to arrive wins and nothing is double-applied.
func (s *vnpayService) VerifyAndHandleIPN(ctx context.Context, query url.Values) (string, error) {
	gotHash := query.Get("vnp_SecureHash")
	wantHash := vnpaySignQuery(query, s.cfg.HashSecret)
	if gotHash == "" || gotHash != wantHash {
		slog.WarnContext(ctx, "vnpay ipn rejected: bad signature", "txn_ref", query.Get("vnp_TxnRef"))
		return vnpayRspInvalidSignature, nil
	}

	txnRef := query.Get("vnp_TxnRef")
	responseCode := query.Get("vnp_ResponseCode")
	transactionNo := query.Get("vnp_TransactionNo")
	raw, _ := json.Marshal(query)

	var rspCode string
	txErr := s.runInTx(ctx, func(tx *gorm.DB) error {
		subs, txns := s.subs, s.txns
		if tx != nil {
			subs, txns = subs.WithTx(tx), txns.WithTx(tx)
		}

		// Locked read: a concurrent replay of the same vnp_TxnRef (VNPay
		// retries IPN until acked) blocks here until this transaction
		// commits, then sees status != pending and falls into the no-op
		// branch below — the two deliveries can never both apply the result.
		txn, err := txns.GetByProviderTxnRefForUpdate(ctx, "vnpay", txnRef)
		if err != nil {
			slog.WarnContext(ctx, "vnpay ipn rejected: unknown txn_ref", "txn_ref", txnRef)
			rspCode = vnpayRspOrderNotFound
			return nil
		}

		if txn.Status != enums.PaymentTransactionStatusPending {
			// Already processed — replay is a no-op, not an error.
			rspCode = vnpayRspAlreadyConfirmed
			return nil
		}

		gotAmount, parseErr := strconv.ParseInt(query.Get("vnp_Amount"), 10, 64)
		if parseErr != nil || gotAmount != txn.Amount*100 {
			rspCode = vnpayRspInvalidAmount
			return nil
		}

		if err := txns.UpdateResult(ctx, txn.ID, statusForResponseCode(responseCode), strPtr(transactionNo), strPtr(responseCode), datatypes.JSON(raw)); err != nil {
			return err
		}

		if responseCode == "00" {
			now := time.Now()
			sub, err := subs.GetByID(ctx, txn.SubscriptionID)
			if err != nil {
				return err
			}
			planDef, err := s.plans.GetByPlan(ctx, sub.Plan)
			if err != nil {
				return err
			}
			end := now.AddDate(0, planDef.Months, 0)
			if err := subs.UpdateStatus(ctx, txn.SubscriptionID, enums.SubscriptionStatusActive, &now, &end); err != nil {
				return err
			}
			slog.InfoContext(ctx, "vnpay payment succeeded", "subscription_id", txn.SubscriptionID, "txn_ref", txnRef)
		} else {
			if err := subs.UpdateStatus(ctx, txn.SubscriptionID, enums.SubscriptionStatusFailed, nil, nil); err != nil {
				return err
			}
			slog.InfoContext(ctx, "vnpay payment failed", "subscription_id", txn.SubscriptionID, "txn_ref", txnRef, "response_code", responseCode)
		}
		rspCode = vnpayRspSuccess
		return nil
	})
	if txErr != nil {
		appErr := common.FromDBError(txErr, "payment transaction")
		logger.Report(ctx, "handle vnpay ipn failed", appErr, "txn_ref", txnRef)
		return vnpayRspUnknownError, appErr
	}

	return rspCode, nil
}

func statusForResponseCode(code string) enums.PaymentTransactionStatus {
	if code == "00" {
		return enums.PaymentTransactionStatusSucceeded
	}
	return enums.PaymentTransactionStatusFailed
}

func strPtr(s string) *string { return &s }
