package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/payments/dtos/requests"
	"engflex-api/internal/modules/payments/dtos/responses"
)

// SubscriptionService answers "what access does this user currently have"
// and owns the admin manual-grant path (no payment, straight to active).
// The payment-driven pending->active/failed transition lives in
// VNPayService, which reuses the same repository.
type SubscriptionService interface {
	// IsPremium has exactly this (bool, error) shape so the service
	// satisfies middleware.PremiumChecker structurally, with no adapter.
	IsPremium(ctx context.Context, userID string) (bool, error)
	GetActive(ctx context.Context, userID string) (*responses.Subscription, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*responses.Subscription, int64, error)
	Grant(ctx context.Context, userID string, req requests.GrantSubscription) (*responses.Subscription, error)
	// Revoke cancels the user's currently-active subscription, if any. A
	// no-op (not an error) when the user has no active subscription — the
	// admin console's "None" plan option always calls this, whether or not
	// there was anything to revoke.
	Revoke(ctx context.Context, userID string) error
	// ListAll backs the admin-wide subscriptions view (every user).
	// status/plan filter when non-empty.
	ListAll(ctx context.Context, status enums.SubscriptionStatus, plan enums.SubscriptionPlan, limit, offset int) ([]*responses.Subscription, int64, error)
}

type subscriptionService struct {
	repo  repositories.SubscriptionRepository
	plans repositories.PlanRepository
}

// NewSubscriptionService builds the service over the repository interfaces.
func NewSubscriptionService(repo repositories.SubscriptionRepository, plans repositories.PlanRepository) SubscriptionService {
	return &subscriptionService{repo: repo, plans: plans}
}

func toSubscriptionResponse(m *models.Subscription) *responses.Subscription {
	return &responses.Subscription{
		ID:        m.ID,
		UserID:    m.UserID,
		Plan:      m.Plan,
		Status:    m.Status,
		Amount:    m.Amount,
		Currency:  m.Currency,
		StartedAt: m.StartedAt,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// IsPremium reports whether the user has a currently-active, unexpired
// subscription. A not-found row from the repository is not an error here —
// it just means the user isn't premium.
func (s *subscriptionService) IsPremium(ctx context.Context, userID string) (bool, error) {
	_, err := s.repo.GetActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "check premium status failed", appErr, "user_id", userID)
		return false, appErr
	}
	return true, nil
}

// GetActive returns the user's current active subscription, or nil (not an
// error) when they have none.
func (s *subscriptionService) GetActive(ctx context.Context, userID string) (*responses.Subscription, error) {
	m, err := s.repo.GetActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "get active subscription failed", appErr, "user_id", userID)
		return nil, appErr
	}
	return toSubscriptionResponse(m), nil
}

func (s *subscriptionService) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*responses.Subscription, int64, error) {
	items, err := s.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "list subscriptions failed", appErr, "user_id", userID)
		return nil, 0, appErr
	}
	total, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "count subscriptions failed", appErr, "user_id", userID)
		return nil, 0, appErr
	}
	out := make([]*responses.Subscription, 0, len(items))
	for _, m := range items {
		out = append(out, toSubscriptionResponse(m))
	}
	return out, total, nil
}

// Grant is the admin manual path: created directly as active, no payment
// attempt and no payment_transactions row. Pulls the current price from
// the plans table — a price edited after this grant never changes what
// this subscription already recorded.
func (s *subscriptionService) Grant(ctx context.Context, userID string, req requests.GrantSubscription) (*responses.Subscription, error) {
	def, err := s.plans.GetByPlan(ctx, req.Plan)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.BadRequest("unknown plan")
		}
		appErr := common.FromDBError(err, "plan")
		logger.Report(ctx, "load plan for grant failed", appErr, "plan", req.Plan)
		return nil, appErr
	}
	now := time.Now()
	end := now.AddDate(0, def.Months, 0)
	m := &models.Subscription{
		UserID:    userID,
		Plan:      req.Plan,
		Status:    enums.SubscriptionStatusActive,
		Amount:    def.Amount,
		Currency:  def.Currency,
		StartedAt: &now,
		ExpiresAt: &end,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "grant subscription failed", appErr, "user_id", userID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "subscription granted", "id", m.ID, "user_id", userID, "plan", req.Plan, "note", req.Note)
	return toSubscriptionResponse(m), nil
}

// Revoke cancels the user's active subscription (status -> cancelled), so
// IsPremium/GetActive stop seeing it immediately. Looks up the row via the
// same query IsPremium uses, so "nothing active" and "already cancelled"
// are indistinguishable and both resolve as a no-op.
func (s *subscriptionService) Revoke(ctx context.Context, userID string) error {
	m, err := s.repo.GetActiveByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "load active subscription for revoke failed", appErr, "user_id", userID)
		return appErr
	}
	if err := s.repo.UpdateStatus(ctx, m.ID, enums.SubscriptionStatusCancelled, nil, nil); err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "revoke subscription failed", appErr, "user_id", userID, "id", m.ID)
		return appErr
	}
	slog.InfoContext(ctx, "subscription revoked", "id", m.ID, "user_id", userID)
	return nil
}

func (s *subscriptionService) ListAll(ctx context.Context, status enums.SubscriptionStatus, plan enums.SubscriptionPlan, limit, offset int) ([]*responses.Subscription, int64, error) {
	items, err := s.repo.ListAll(ctx, status, plan, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "list all subscriptions failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.CountAll(ctx, status, plan)
	if err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "count all subscriptions failed", appErr)
		return nil, 0, appErr
	}
	out := make([]*responses.Subscription, 0, len(items))
	for _, m := range items {
		out = append(out, toSubscriptionResponse(m))
	}
	return out, total, nil
}
