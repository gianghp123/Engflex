package services

import (
	"context"
	"log/slog"
	"time"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/access/dtos/responses"
)

// EntitlementsService assembles the /premium/me view from the policy
// service's resolution and the user's counters.
type EntitlementsService struct {
	policy *PolicyService
}

// NewEntitlementsService builds the view service over the shared policy.
func NewEntitlementsService(policy *PolicyService) *EntitlementsService {
	return &EntitlementsService{policy: policy}
}

// MyEntitlements resolves plan, trial, and both counter views for one user.
func (s *EntitlementsService) MyEntitlements(ctx context.Context, userID string, isAdmin bool) (*responses.MyEntitlements, error) {
	p, trial, err := s.policy.Principal(ctx, userID, isAdmin)
	if err != nil {
		return nil, err
	}
	limits, err := s.policy.LimitsForPlan(ctx, p.Plan)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	usage := map[enums.CounterKind]CounterUsage{}
	if !isAdmin {
		rows, err := s.policy.counters.ListByUser(ctx, userID)
		if err != nil {
			appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
			logger.Report(ctx, "entitlements usage read failed", appErr, "userID", userID)
			return nil, appErr
		}
		for _, r := range rows {
			period := limits[r.CounterKey].Period
			if r.PeriodKey == s.policy.PeriodKey(period, now) {
				usage[r.CounterKey] = CounterUsage{Used: r.Used, Reserved: r.Reserved}
			}
		}
	}

	out := &responses.MyEntitlements{Plan: p.Plan}
	if p.Plan == enums.PlanKeyTrial && trial.Active {
		t := responses.TrialView{Active: true, EndsAt: trial.EndsAt, CreditsGranted: trial.CreditsGranted}
		out.Trial = &t
	}
	if sub := s.activeSubscription(ctx, userID); sub != nil {
		out.IsPremium = true
		out.ExpiresAt = sub.ExpiresAt
	}
	out.Credits = s.counterView(enums.CounterKindAICredit, limits, usage, now)
	out.Voice = s.counterView(enums.CounterKindVoiceSession, limits, usage, now)
	out.Costs = []responses.ActionCost{
		{Action: string(ActionVoiceStart), Credits: 1},
		{Action: string(ActionExamReview), Credits: 1},
		{Action: string(ActionStudyPlan), Credits: 2},
	}
	slog.InfoContext(ctx, "entitlements resolved", "userID", userID, "plan", string(p.Plan), "isPremium", out.IsPremium)
	return out, nil
}

func (s *EntitlementsService) counterView(counter enums.CounterKind, limits map[enums.CounterKind]Limit, usage map[enums.CounterKind]CounterUsage, now time.Time) responses.CounterView {
	lim := limits[counter]
	u := usage[counter]
	remaining := lim.Value - u.Used - u.Reserved
	if !lim.Enabled || remaining < 0 {
		remaining = 0
	}
	view := responses.CounterView{Limit: lim.Value, Used: u.Used, Reserved: u.Reserved, Remaining: remaining}
	if lim.Period != enums.UsagePeriodTotal && lim.Enabled {
		view.ResetAt = s.policy.ResetAt(lim.Period, now)
	}
	return view
}

// activeSubscription is a second read owned by this service only because
// IsPremium must reflect the billing row itself, not the derived plan.
func (s *EntitlementsService) activeSubscription(ctx context.Context, userID string) *models.Subscription {
	sub, err := s.policy.subscriptions.GetActiveByUserID(ctx, userID)
	if err != nil || sub == nil {
		return nil
	}
	return sub
}
