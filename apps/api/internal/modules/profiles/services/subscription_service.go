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
)

// SubscriptionService owns the premium-grant lifecycle and the paywall
// check. A missing active row is a legitimate "not premium" outcome, not an
// error: IsPremium reports it as (false, nil).
type SubscriptionService interface {
	IsPremium(ctx context.Context, userID string) (bool, error)
	Grant(ctx context.Context, userID, plan string, expiresAt *time.Time) (*models.Subscription, error)
	ListByUserID(ctx context.Context, userID string) ([]*models.Subscription, error)
}

type subscriptionService struct {
	repo repositories.SubscriptionRepository
}

// NewSubscriptionService builds the service over the repository interface.
func NewSubscriptionService(repo repositories.SubscriptionRepository) SubscriptionService {
	return &subscriptionService{repo: repo}
}

func (s *subscriptionService) IsPremium(ctx context.Context, userID string) (bool, error) {
	_, err := s.repo.GetActiveByUserID(ctx, userID)
	if err == nil {
		slog.InfoContext(ctx, "premium check", "userID", userID, "premium", true)
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		slog.InfoContext(ctx, "premium check", "userID", userID, "premium", false)
		return false, nil
	}
	appErr := common.FromDBError(err, "subscription")
	logger.Report(ctx, "premium check failed", appErr, "userID", userID)
	return false, appErr
}

func (s *subscriptionService) Grant(ctx context.Context, userID, plan string, expiresAt *time.Time) (*models.Subscription, error) {
	m := &models.Subscription{
		UserID:    userID,
		Plan:      plan,
		Status:    enums.SubscriptionStatusActive,
		StartedAt: time.Now(),
		ExpiresAt: expiresAt,
	}
	if err := s.repo.Create(ctx, m); err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "grant subscription failed", appErr, "userID", userID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "subscription granted", "userID", userID, "plan", plan, "id", m.ID)
	return m, nil
}

func (s *subscriptionService) ListByUserID(ctx context.Context, userID string) ([]*models.Subscription, error) {
	items, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "subscription")
		logger.Report(ctx, "list subscriptions failed", appErr, "userID", userID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "subscriptions listed", "userID", userID, "count", len(items))
	return items, nil
}
