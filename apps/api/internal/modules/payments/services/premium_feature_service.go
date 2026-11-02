package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/payments/dtos/requests"
	"engflex-api/internal/utils"
)

// PremiumFeatureService owns the admin-managed premium-gating registry:
// named app capabilities and whether each currently requires premium. This
// is a management/config surface only — no route enforces these flags yet.
type PremiumFeatureService interface {
	List(ctx context.Context, limit, offset int) ([]*models.PremiumFeature, int64, error)
	Create(ctx context.Context, req requests.CreatePremiumFeature) (*models.PremiumFeature, error)
	Update(ctx context.Context, id string, req requests.UpdatePremiumFeature) (*models.PremiumFeature, error)
	Delete(ctx context.Context, id string) error
}

type premiumFeatureService struct {
	repo repositories.PremiumFeatureRepository
}

// NewPremiumFeatureService builds the service over the repository interface.
func NewPremiumFeatureService(repo repositories.PremiumFeatureRepository) PremiumFeatureService {
	return &premiumFeatureService{repo: repo}
}

func (s *premiumFeatureService) List(ctx context.Context, limit, offset int) ([]*models.PremiumFeature, int64, error) {
	items, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "premium feature")
		logger.Report(ctx, "list premium features failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.repo.Count(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "premium feature")
		logger.Report(ctx, "count premium features failed", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "premiumfeature list completed", "count", len(items), "total", total)
	return items, total, nil
}

func (s *premiumFeatureService) Create(ctx context.Context, req requests.CreatePremiumFeature) (*models.PremiumFeature, error) {
	m := &models.PremiumFeature{}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map create premium feature failed", err)
		return nil, common.Internal()
	}
	if err := s.repo.Create(ctx, m); err != nil {
		appErr := common.FromDBError(err, "premium feature")
		logger.Report(ctx, "create premium feature failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "premium feature created", "id", m.ID, "key", m.Key)
	return m, nil
}

func (s *premiumFeatureService) Update(ctx context.Context, id string, req requests.UpdatePremiumFeature) (*models.PremiumFeature, error) {
	m, err := s.repo.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "premium feature")
		logger.Report(ctx, "get premium feature for update failed", appErr, "id", id)
		return nil, appErr
	}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map update premium feature failed", err)
		return nil, common.Internal()
	}
	if err := s.repo.Update(ctx, m); err != nil {
		appErr := common.FromDBError(err, "premium feature")
		logger.Report(ctx, "update premium feature failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "premium feature updated", "id", id)
	return m, nil
}

func (s *premiumFeatureService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "premium feature")
		logger.Report(ctx, "delete premium feature failed", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "premium feature deleted", "id", id)
	return nil
}
