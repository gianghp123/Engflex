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

// PlanLimitsReader reads a plan's monthly counter limits for the pricing
// table. *accessservices.PolicyService satisfies it structurally via
// PlanMonthlyLimits.
type PlanLimitsReader interface {
	// PlanMonthlyLimits returns (ai_credit, voice_session) monthly limits.
	PlanMonthlyLimits(ctx context.Context, plan enums.SubscriptionPlan) (credits int, voice int, err error)
}

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
	// limits fills the pricing-table entitlement summary; nil keeps the
	// service catalog-only (tests, callers without the access module).
	limits PlanLimitsReader
}

// NewPlanService builds the service over the repository interface.
func NewPlanService(repo repositories.PlanRepository) PlanService {
	return &planService{repo: repo}
}

// NewPlanServiceWithLimits builds the service with the pricing-table
// entitlement summary filled from the access policy snapshot.
func NewPlanServiceWithLimits(repo repositories.PlanRepository, limits PlanLimitsReader) PlanService {
	return &planService{repo: repo, limits: limits}
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

// toPlanResponse fills the pricing-table entitlement summary best-effort: a
// catalog read failure logs and leaves zeros rather than failing the whole
// pricing page (the snapshot the reader serves already falls back to
// defaults, so this is a second net, not the first).
func (s *planService) toPlanResponse(ctx context.Context, m *models.Plan) *responses.Plan {
	out := toPlanResponse(m)
	if s.limits == nil {
		return out
	}
	credits, voice, err := s.limits.PlanMonthlyLimits(ctx, m.Plan)
	if err != nil {
		logger.Report(ctx, "plan limits lookup failed", common.Internal(), "plan", m.Plan)
		return out
	}
	out.Credits = credits
	out.VoiceSessions = voice
	return out
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
		out = append(out, s.toPlanResponse(ctx, m))
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
		out = append(out, s.toPlanResponse(ctx, m))
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
	return s.toPlanResponse(ctx, updated), nil
}
