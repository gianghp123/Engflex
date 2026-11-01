package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/payments/dtos/requests"
	"engflex-api/internal/modules/payments/dtos/responses"
	"engflex-api/internal/utils"
)

// PremiumFeatureService owns the admin-managed premium-gating registry:
// named app capabilities and whether each currently requires premium. This
// is a management/config surface only — no route enforces these flags yet.
type PremiumFeatureService interface {
	List(ctx context.Context, limit, offset int) ([]*responses.PremiumFeature, int64, error)
	Create(ctx context.Context, req requests.CreatePremiumFeature) (*responses.PremiumFeature, error)
	Update(ctx context.Context, id string, req requests.UpdatePremiumFeature) (*responses.PremiumFeature, error)
	Delete(ctx context.Context, id string) error
}

type premiumFeatureService struct {
	repo repositories.PremiumFeatureRepository
}

// NewPremiumFeatureService builds the service over the repository interface.
func NewPremiumFeatureService(repo repositories.PremiumFeatureRepository) PremiumFeatureService {
	return &premiumFeatureService{repo: repo}
}

func toPremiumFeatureResponse(m *models.PremiumFeature) *responses.PremiumFeature {
	dto := &responses.PremiumFeature{}
	_ = utils.Map(dto, m)
	return dto
}

func (s *premiumFeatureService) List(ctx context.Context, limit, offset int) ([]*responses.PremiumFeature, int64, error) {
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
	out := make([]*responses.PremiumFeature, 0, len(items))
	for _, m := range items {
		out = append(out, toPremiumFeatureResponse(m))
	}
	return out, total, nil
}

func (s *premiumFeatureService) Create(ctx context.Context, req requests.CreatePremiumFeature) (*responses.PremiumFeature, error) {
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
	return toPremiumFeatureResponse(m), nil
}

func (s *premiumFeatureService) Update(ctx context.Context, id string, req requests.UpdatePremiumFeature) (*responses.PremiumFeature, error) {
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
	return toPremiumFeatureResponse(m), nil
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
