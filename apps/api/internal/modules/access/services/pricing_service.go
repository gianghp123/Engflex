package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/access/dtos/responses"
	"engflex-api/internal/utils"
)

// PricingService composes the public pricing catalog from the active plans
// list plus the policy snapshot (free limits, trial terms, per-action
// costs). Any plan-store read failure is a 503 policy_unavailable: the
// public catalog fails closed like every other paid surface.
type PricingService struct {
	plans  repositories.PlanRepository
	policy *PolicyService
}

// NewPricingService builds the catalog service over the plans repo and the
// shared policy snapshot. Never pass nil; wiring owns both handles.
func NewPricingService(plans repositories.PlanRepository, policy *PolicyService) *PricingService {
	return &PricingService{plans: plans, policy: policy}
}

// Get assembles the public pricing catalog: active plans with their monthly
// limits, the free-tier entitlements, trial terms, and per-action costs.
func (s *PricingService) Get(ctx context.Context) (*responses.PricingCatalog, error) {
	plans, err := s.plans.ListActive(ctx)
	if err != nil {
		appErr := common.ServiceUnavailable("pricing unavailable").WithCode(common.CodePolicyUnavailable)
		logger.Report(ctx, "public pricing plans read failed", appErr)
		return nil, appErr
	}
	free, err := s.policy.LimitsForPlan(ctx, enums.PlanKeyFree)
	if err != nil {
		appErr := common.ServiceUnavailable("pricing unavailable").WithCode(common.CodePolicyUnavailable)
		logger.Report(ctx, "public pricing free limits read failed", appErr)
		return nil, appErr
	}
	terms := s.policy.TrialTerms()
	out := &responses.PricingCatalog{
		Free: responses.PricingEntitlements{
			Credits:       free[enums.CounterKindAICredit].Value,
			VoiceSessions: free[enums.CounterKindVoiceSession].Value,
			CreditsPeriod: free[enums.CounterKindAICredit].Period,
			VoicePeriod:   free[enums.CounterKindVoiceSession].Period,
		},
		Trial: responses.PricingTrial{Days: terms.Days, Credits: terms.Credits, VoiceSessions: terms.VoiceSessions},
		Plans: make([]responses.PricingPlan, 0, len(plans)),
	}
	for _, p := range plans {
		credits, voice, err := s.policy.PlanMonthlyLimits(ctx, p.Plan)
		if err != nil {
			appErr := common.ServiceUnavailable("pricing unavailable").WithCode(common.CodePolicyUnavailable)
			logger.Report(ctx, "public pricing plan limits read failed", appErr, "plan", string(p.Plan))
			return nil, appErr
		}
		out.Plans = append(out.Plans, responses.PricingPlan{
			Plan: p.Plan, Months: p.Months, Amount: p.Amount, Currency: p.Currency,
			Credits: credits, VoiceSessions: voice,
		})
	}
	out.Costs = make([]responses.ActionCost, 0, len(paidActions))
	_ = utils.MapSlice(&out.Costs, s.policy.ActionCosts())
	slog.InfoContext(ctx, "public pricing composed", "plans", len(out.Plans))
	return out, nil
}
