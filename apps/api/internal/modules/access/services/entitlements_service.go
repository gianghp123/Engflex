package services

import (
	"context"
	"log/slog"
	"time"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/access/dtos/responses"
	"engflex-api/internal/utils"
)

// EntitlementsService assembles the /premium/me and /premium/usage views
// from the policy service's resolution, the user's counters, and (for the
// usage view) the user's settled event history.
type EntitlementsService struct {
	policy *PolicyService
	events repositories.UsageEventRepository
}

// NewEntitlementsService builds the view service over the shared policy and
// the usage-event ledger. Never pass nil; wiring owns both handles.
func NewEntitlementsService(policy *PolicyService, events repositories.UsageEventRepository) *EntitlementsService {
	return &EntitlementsService{policy: policy, events: events}
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
	usage, err := s.currentUsage(ctx, userID, isAdmin, limits, now)
	if err != nil {
		return nil, err
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
	out.Costs = s.actionCosts()
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
		resetAt := s.policy.ResetAt(lim.Period, now)
		view.ResetAt = &resetAt
	}
	return view
}

// MyUsage resolves the GET /premium/usage body for one user: plan, both
// counter views, the per-action costs table, and the paginated history of
// settled charges (committed + reserved, newest first). Owner scoping is by
// the authenticated userID only. A store failure anywhere is a 503
// policy_unavailable (paid metering fails closed); validation stays silent
// per the house rule because paging is sanitized, never rejected.
func (s *EntitlementsService) MyUsage(ctx context.Context, userID string, isAdmin bool, page, pageSize int) (*responses.MyUsage, error) {
	p, _, err := s.policy.Principal(ctx, userID, isAdmin)
	if err != nil {
		return nil, err
	}
	limits, err := s.policy.LimitsForPlan(ctx, p.Plan)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	usage, err := s.currentUsage(ctx, userID, isAdmin, limits, now)
	if err != nil {
		return nil, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = common.DefaultPageSize
	}
	if pageSize > common.MaxPageSize {
		pageSize = common.MaxPageSize
	}
	offset := (page - 1) * pageSize
	statuses := []enums.UsageEventStatus{enums.UsageEventStatusCommitted, enums.UsageEventStatusReserved}
	events, err := s.events.ListByUserStatus(ctx, userID, statuses, pageSize, offset)
	if err != nil {
		appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
		logger.Report(ctx, "usage history read failed", appErr, "userID", userID)
		return nil, appErr
	}
	total, err := s.events.CountByUserStatus(ctx, userID, statuses)
	if err != nil {
		appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
		logger.Report(ctx, "usage history count failed", appErr, "userID", userID)
		return nil, appErr
	}

	out := &responses.MyUsage{
		Plan: p.Plan,
		Counters: []responses.UsageCounterView{
			s.usageCounterView(enums.CounterKindAICredit, limits, usage, now),
			s.usageCounterView(enums.CounterKindVoiceSession, limits, usage, now),
		},
		Costs: s.actionCosts(),
		History: responses.UsageHistory{
			Items:    make([]responses.UsageHistoryItem, 0, len(events)),
			Page:     page,
			PageSize: pageSize,
			Total:    total,
		},
	}
	for _, e := range events {
		out.History.Items = append(out.History.Items, responses.UsageHistoryItem{
			ID:        e.ID,
			Action:    string(e.Action),
			Credits:   e.Amount,
			Status:    e.Status,
			CreatedAt: e.CreatedAt,
			RefType:   e.RefType,
			RefID:     e.RefID,
		})
	}
	slog.InfoContext(ctx, "usage resolved", "userID", userID, "plan", string(p.Plan),
		"page", page, "pageSize", pageSize, "total", total)
	return out, nil
}

// currentUsage reads the caller's counters for the live period of each plan
// limit. Admins bypass metering (zero views, like /premium/me); a counter
// row whose period key does not match the plan's current key belongs to a
// rolled-over period and is ignored.
func (s *EntitlementsService) currentUsage(ctx context.Context, userID string, isAdmin bool, limits map[enums.CounterKind]Limit, now time.Time) (map[enums.CounterKind]CounterUsage, error) {
	usage := map[enums.CounterKind]CounterUsage{}
	if isAdmin {
		return usage, nil
	}
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
	return usage, nil
}

// usageCounterView is the keyed form of counterView for the /premium/usage
// counters list. One helper owns both shapes so the two views cannot drift.
func (s *EntitlementsService) usageCounterView(counter enums.CounterKind, limits map[enums.CounterKind]Limit, usage map[enums.CounterKind]CounterUsage, now time.Time) responses.UsageCounterView {
	v := s.counterView(counter, limits, usage, now)
	return responses.UsageCounterView{
		Counter:   counter,
		Limit:     v.Limit,
		Used:      v.Used,
		Reserved:  v.Reserved,
		Remaining: v.Remaining,
		ResetAt:   v.ResetAt,
	}
}

// actionCosts renders the policy catalog snapshot as the per-action costs
// table for the pricing/upgrade surfaces. copier matches CostEntry to
// ActionCost by Go field name; the admin/public DTOs share the same shape
// so the closed action set cannot drift between them.
func (s *EntitlementsService) actionCosts() []responses.ActionCost {
	out := make([]responses.ActionCost, 0, len(paidActions))
	_ = utils.MapSlice(&out, s.policy.ActionCosts())
	return out
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
