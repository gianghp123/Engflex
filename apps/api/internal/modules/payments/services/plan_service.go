package services

import (
	"log/slog"

	"context"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/payments/dtos/requests"
	"engflex-api/internal/modules/payments/dtos/responses"
)

// PlanService owns the admin-editable price of each fixed-duration
// package. Rows are fixed to the three seeded enums.SubscriptionPlan
// values (migration-seeded) — List/Update only, no Create/Delete.
type PlanService interface {
	// List returns every plan, active or not (admin view).
	List(ctx context.Context) ([]*responses.Plan, error)
	// ListActive returns only plans the learner can currently buy
	// (public pricing page).
	ListActive(ctx context.Context) ([]*responses.Plan, error)
	Update(ctx context.Context, plan enums.SubscriptionPlan, req requests.UpdatePlan) (*responses.Plan, error)
}

type planService struct {
	repo repositories.PlanRepository
}

// NewPlanService builds the service over the repository interface.
func NewPlanService(repo repositories.PlanRepository) PlanService {
	return &planService{repo: repo}
}

func toPlanResponse(m *models.Plan) *responses.Plan {
	return &responses.Plan{
		Plan:      m.Plan,
		Months:    m.Months,
		Amount:    m.Amount,
		Currency:  m.Currency,
		IsActive:  m.IsActive,
		UpdatedAt: m.UpdatedAt,
	}
}

func (s *planService) List(ctx context.Context) ([]*responses.Plan, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "plan")
		logger.Report(ctx, "list plans failed", appErr)
		return nil, appErr
	}
	out := make([]*responses.Plan, 0, len(items))
	for _, m := range items {
		out = append(out, toPlanResponse(m))
	}
	return out, nil
}

func (s *planService) ListActive(ctx context.Context) ([]*responses.Plan, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "plan")
		logger.Report(ctx, "list active plans failed", appErr)
		return nil, appErr
	}
	out := make([]*responses.Plan, 0, len(items))
	for _, m := range items {
		if !m.IsActive {
			continue
		}
		out = append(out, toPlanResponse(m))
	}
	return out, nil
}

func (s *planService) Update(ctx context.Context, plan enums.SubscriptionPlan, req requests.UpdatePlan) (*responses.Plan, error) {
	m := &models.Plan{
		Plan:     plan,
		Months:   req.Months,
		Amount:   req.Amount,
		Currency: req.Currency,
		IsActive: req.IsActive,
	}
	if err := s.repo.Update(ctx, m); err != nil {
		appErr := common.FromDBError(err, "plan")
		logger.Report(ctx, "update plan failed", appErr, "plan", plan)
		return nil, appErr
	}
	updated, err := s.repo.GetByPlan(ctx, plan)
	if err != nil {
		appErr := common.FromDBError(err, "plan")
		logger.Report(ctx, "reload plan after update failed", appErr, "plan", plan)
		return nil, appErr
	}
	slog.InfoContext(ctx, "plan updated", "plan", plan, "amount", req.Amount, "months", req.Months, "is_active", req.IsActive)
	return toPlanResponse(updated), nil
}
