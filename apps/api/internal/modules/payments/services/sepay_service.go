package services

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
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

// SepayWebhook is the payload SePay POSTs to the webhook endpoint for every
// bank transaction on the linked account — both directions, not just
// payments received. See https://developer.sepay.vn (sepay-webhooks).
type SepayWebhook struct {
	ID              int    `json:"id"`
	Gateway         string `json:"gateway"`
	TransactionDate string `json:"transactionDate"`
	AccountNumber   string `json:"accountNumber"`
	SubAccount      string `json:"subAccount"`
	Code            string `json:"code"`
	Content         string `json:"content"`
	TransferType    string `json:"transferType"`
	Description     string `json:"description"`
	TransferAmount  int64  `json:"transferAmount"`
	Accumulated     int64  `json:"accumulated"`
	ReferenceCode   string `json:"referenceCode"`
}

// SepayConfig is the subset of config.SepayConfig the service needs,
// narrowed here for the same reason VNPayConfig is (see vnpay_service.go).
type SepayConfig struct {
	APIKey        string
	WebhookSecret string
	BaseURL       string
	BankAccount   string
	BankName      string
	AccountHolder string
}

// SepayCheckoutSession hands the learner's browser a QR image to render
// inline and a reference to poll GetStatus with — there is no redirect:
// SePay pays by the learner scanning the code in their own banking app.
type SepayService interface {
	BuildCheckoutQR(ctx context.Context, userID string, plan enums.SubscriptionPlan) (*responses.SepayCheckoutSession, error)
	// VerifyWebhookAuth checks the incoming "Authorization" header against
	// the configured API key (SePay's "Apikey" auth scheme). Done here, not
	// in middleware, because this is the only route that needs it.
	VerifyWebhookAuth(header string) bool
	HandleWebhook(ctx context.Context, payload SepayWebhook) error
	GetStatus(ctx context.Context, reference string) (*responses.SepayStatus, error)
}

type sepayService struct {
	cfg    SepayConfig
	db     *gorm.DB
	subs   repositories.SubscriptionRepository
	txns   repositories.PaymentTransactionRepository
	plans  repositories.PlanRepository
	client *sepayClient
}

func NewSepayService(cfg SepayConfig, db *gorm.DB, subs repositories.SubscriptionRepository, txns repositories.PaymentTransactionRepository, plans repositories.PlanRepository) SepayService {
	return &sepayService{
		cfg: cfg, db: db, subs: subs, txns: txns, plans: plans,
		client: newSepayClient(cfg.BaseURL, cfg.APIKey),
	}
}

func (s *sepayService) runInTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if s.db == nil {
		return fn(nil)
	}
	return s.db.WithContext(ctx).Transaction(fn)
}

// generateReference returns a short alphanumeric code for the bank
// transfer's "content" field. Short and digit-heavy on purpose: VietQR
// content often gets truncated or has spaces collapsed by the sending
// bank's own app, and an all-uppercase-letters-plus-digits code survives
// that better than a UUID would.
func generateReference() string {
	const alphabet = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ" // no O/I, easy to misread
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	out := make([]byte, 8)
	for i, b := range buf {
		out[i] = alphabet[int(b)%len(alphabet)]
	}
	return "DH" + string(out)
}

// bankAccount resolves which account the QR pays into. A statically
// configured account wins over SePay's API, which is only consulted as a
// fallback: an API key issued for a different SePay organisation than the
// linked bank account returns that other organisation's account instead
// of failing, so the static value is both simpler and harder to get
// silently wrong.
func (s *sepayService) bankAccount(ctx context.Context) (*SepayBankAccount, error) {
	if s.cfg.BankAccount != "" && s.cfg.BankName != "" {
		return &SepayBankAccount{
			AccountNumber:     s.cfg.BankAccount,
			BankShortName:     s.cfg.BankName,
			AccountHolderName: s.cfg.AccountHolder,
		}, nil
	}
	return s.client.PrimaryBankAccount(ctx)
}

func (s *sepayService) BuildCheckoutQR(ctx context.Context, userID string, plan enums.SubscriptionPlan) (*responses.SepayCheckoutSession, error) {
	def, err := s.plans.GetByPlan(ctx, plan)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.BadRequest("unknown plan")
		}
		appErr := common.FromDBError(err, "plan")
		logger.Report(ctx, "load plan for sepay checkout failed", appErr, "plan", plan)
		return nil, appErr
	}
	if !def.IsActive {
		return nil, common.BadRequest("plan is not available")
	}

	account, err := s.bankAccount(ctx)
	if err != nil {
		logger.Report(ctx, "sepay bank account lookup failed", common.ServiceUnavailable("payment gateway unavailable"))
		return nil, common.ServiceUnavailable("payment gateway unavailable")
	}

	sub := &models.Subscription{
		UserID:   userID,
		Plan:     plan,
		Status:   enums.SubscriptionStatusPending,
		Amount:   def.Amount,
		Currency: def.Currency,
	}
	reference := generateReference()
	txn := &models.PaymentTransaction{
		Provider:       "sepay",
		ProviderTxnRef: reference,
		Amount:         def.Amount,
		Status:         enums.PaymentTransactionStatusPending,
	}

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
		logger.Report(ctx, "create sepay checkout session failed", appErr, "user_id", userID)
		return nil, appErr
	}

	q := url.Values{}
	q.Set("acc", account.AccountNumber)
	q.Set("bank", account.BankShortName)
	q.Set("amount", strconv.FormatInt(def.Amount, 10))
	q.Set("des", reference)
	q.Set("template", "compact")
	qrURL := "https://vietqr.app/img?" + q.Encode()

	slog.InfoContext(ctx, "sepay checkout created", "subscription_id", sub.ID, "reference", reference, "plan", plan)
	return &responses.SepayCheckoutSession{
		QRUrl:             qrURL,
		Reference:         reference,
		Amount:            def.Amount,
		BankAccountNumber: account.AccountNumber,
		BankShortName:     account.BankShortName,
		AccountHolderName: account.AccountHolderName,
	}, nil
}

func (s *sepayService) VerifyWebhookAuth(header string) bool {
	want := "Apikey " + s.cfg.WebhookSecret
	return header != "" && header == want
}

// HandleWebhook matches one incoming bank transaction to a pending
// checkout and, on a match, activates the subscription. Unmatched or
// outgoing transactions are not errors — SePay's webhook fires for every
// transaction on the account, most of which have nothing to do with this
// app (see transferType != "in"), and it still must get back
// {"success": true} either way or SePay will keep retrying.
func (s *sepayService) HandleWebhook(ctx context.Context, payload SepayWebhook) error {
	if payload.TransferType != "in" {
		return nil
	}
	raw, _ := json.Marshal(payload)
	matchText := payload.Content + " " + payload.Code + " " + payload.Description

	return s.runInTx(ctx, func(tx *gorm.DB) error {
		subs, txns := s.subs, s.txns
		if tx != nil {
			subs, txns = subs.WithTx(tx), txns.WithTx(tx)
		}

		txn, err := txns.FindPendingMatchForUpdate(ctx, "sepay", matchText)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				slog.WarnContext(ctx, "sepay webhook: no pending transaction matched", "content", payload.Content)
				return nil
			}
			return err
		}

		if payload.TransferAmount != txn.Amount {
			slog.WarnContext(ctx, "sepay webhook: amount mismatch", "reference", txn.ProviderTxnRef, "want", txn.Amount, "got", payload.TransferAmount)
			return nil
		}

		providerTxnNo := strconv.Itoa(payload.ID)
		responseCode := "00"
		if err := txns.UpdateResult(ctx, txn.ID, enums.PaymentTransactionStatusSucceeded, &providerTxnNo, &responseCode, datatypes.JSON(raw)); err != nil {
			return err
		}

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
		slog.InfoContext(ctx, "sepay payment succeeded", "subscription_id", txn.SubscriptionID, "reference", txn.ProviderTxnRef)
		return nil
	})
}

func (s *sepayService) GetStatus(ctx context.Context, reference string) (*responses.SepayStatus, error) {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return nil, common.BadRequest("reference is required")
	}
	txn, err := s.txns.GetByProviderTxnRef(ctx, "sepay", reference)
	if err != nil {
		appErr := common.FromDBError(err, "payment transaction")
		return nil, appErr
	}
	status := responses.SepayStatus{Status: string(txn.Status)}
	if txn.Status == enums.PaymentTransactionStatusSucceeded {
		sub, err := s.subs.GetByID(ctx, txn.SubscriptionID)
		if err == nil {
			status.ExpiresAt = sub.ExpiresAt
		}
	}
	return &status, nil
}
