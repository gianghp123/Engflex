package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"

	"engflex-api/config"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
)

// SubscriptionPlanSource is the narrow payments boundary (spec §5.2): the
// access module never imports the payments module; the subscription
// repository satisfies this structurally.
type SubscriptionPlanSource interface {
	GetActiveByUserID(ctx context.Context, userID string) (*models.Subscription, error)
}

// TrialInfo is the resolved trial state for /premium/me.
type TrialInfo struct {
	Active         bool
	EndsAt         time.Time
	CreditsGranted int
}

// PolicyService resolves principals, caches the limit catalog, and turns
// actions into decisions. Decide is the advisory pre-flight; Reserve
// (UsageService) is authoritative (spec §5.5).
type PolicyService struct {
	subscriptions SubscriptionPlanSource
	trials        repositories.UserTrialRepository
	counters      repositories.UsageCounterRepository
	limits        repositories.PlanFeatureLimitRepository
	cfg           config.AccessConfig

	mu         sync.RWMutex
	snapshot   map[enums.PlanKey]map[enums.CounterKind]Limit
	snapshotAt time.Time
}

// PolicyService is the UsageService LimitSource: limits re-resolve at Reserve
// time so a plan change between Decide and Reserve applies immediately.
var _ LimitSource = (*PolicyService)(nil)

// NewPolicyService builds the policy service. Never pass nil; wiring builds
// the snapshot before serving traffic.
func NewPolicyService(subscriptions SubscriptionPlanSource, trials repositories.UserTrialRepository, counters repositories.UsageCounterRepository, limits repositories.PlanFeatureLimitRepository, cfg config.AccessConfig) *PolicyService {
	if cfg.Timezone == "" {
		cfg.Timezone = "Asia/Ho_Chi_Minh"
	}
	if _, err := time.LoadLocation(cfg.Timezone); err != nil {
		slog.Error("invalid APP_TIMEZONE, period keys fall back to UTC", "value", cfg.Timezone)
	}
	return &PolicyService{subscriptions: subscriptions, trials: trials, counters: counters, limits: limits, cfg: cfg}
}

// SnapshotLimits loads the catalog into memory. On failure it logs and keeps
// serving: the built-in defaults are installed only when no snapshot exists
// yet, otherwise the last-good catalog (and its timestamp) is left untouched.
// The process never fails to boot (spec §5.6).
func (s *PolicyService) SnapshotLimits(ctx context.Context) error {
	rows, err := s.limits.ListAll(ctx)
	if err != nil {
		appErr := common.Internal()
		logger.Report(ctx, "policy snapshot load failed", appErr, "cause", err.Error())
		s.mu.Lock()
		if s.snapshot == nil {
			s.snapshot = defaultLimits()
			s.snapshotAt = time.Now()
		}
		s.mu.Unlock()
		return nil
	}
	next := map[enums.PlanKey]map[enums.CounterKind]Limit{}
	for _, r := range rows {
		plan, ok := next[r.PlanKey]
		if !ok {
			plan = map[enums.CounterKind]Limit{}
			next[r.PlanKey] = plan
		}
		plan[r.CounterKey] = Limit{Enabled: r.IsEnabled, Value: r.LimitValue, Period: r.Period}
	}
	s.mu.Lock()
	s.snapshot = next
	s.snapshotAt = time.Now()
	s.mu.Unlock()
	slog.InfoContext(ctx, "policy snapshot loaded", "rows", len(rows))
	return nil
}

// refreshSnapshot reloads when the TTL (30s) has passed; the admin PUT
// invalidates the writer's snapshot immediately (spec §5.6).
func (s *PolicyService) refreshSnapshot(ctx context.Context) {
	s.mu.RLock()
	stale := time.Since(s.snapshotAt) > 30*time.Second
	s.mu.RUnlock()
	if !stale {
		return
	}
	_ = s.SnapshotLimits(ctx)
}

// LimitsForPlan implements UsageService's LimitSource: resolved from the
// cached snapshot, refreshed on TTL. The returned map is a per-call copy;
// the cached snapshot stays immutable to callers.
func (s *PolicyService) LimitsForPlan(ctx context.Context, plan enums.PlanKey) (map[enums.CounterKind]Limit, error) {
	s.refreshSnapshot(ctx)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.snapshot == nil {
		return defaultLimits()[plan], nil
	}
	if planLimits, ok := s.snapshot[plan]; ok {
		out := make(map[enums.CounterKind]Limit, len(planLimits))
		for k, v := range planLimits {
			out[k] = v
		}
		return out, nil
	}
	return map[enums.CounterKind]Limit{}, nil
}

// PlanMonthlyLimits returns a paid plan's (ai_credit, voice_session) monthly
// limits for the pricing table (spec §10). The plan key maps structurally
// onto the subscription plan's wire value. Absent plans yield (0, 0, nil) —
// the pricing row renders without entitlements rather than failing.
func (s *PolicyService) PlanMonthlyLimits(ctx context.Context, plan enums.SubscriptionPlan) (credits int, voice int, err error) {
	limits, err := s.LimitsForPlan(ctx, enums.PlanKey(plan))
	if err != nil {
		return 0, 0, err
	}
	return limits[enums.CounterKindAICredit].Value, limits[enums.CounterKindVoiceSession].Value, nil
}

// Principal resolves the caller's plan (spec §5.5): active paid subscription
// → its plan; else active trial → trial; else free. A subscription lookup
// failure is logged and resolution continues with the trial store — the
// trial store is independent, and a billing-store flap must not downgrade a
// trial user to free. Free content is never blocked by the billing store
// (spec §12).
func (s *PolicyService) Principal(ctx context.Context, userID string, isAdmin bool) (Principal, TrialInfo, error) {
	p := Principal{UserID: userID, Plan: enums.PlanKeyFree}
	if isAdmin {
		p.Plan = enums.PlanKeyAdmin
		p.IsAdmin = true
		return p, TrialInfo{}, nil
	}
	trial := TrialInfo{}
	if sub, err := s.subscriptions.GetActiveByUserID(ctx, userID); err == nil && sub != nil {
		p.Plan = enums.PlanKey(sub.Plan)
		return p, trial, nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		logger.Report(ctx, "subscription lookup failed", common.Internal(), "userID", userID, "cause", err.Error())
	}
	if tr, terr := s.trials.Get(ctx, userID); terr != nil {
		if terr != gorm.ErrRecordNotFound {
			logger.Report(ctx, "trial lookup failed", common.Internal(), "userID", userID, "cause", terr.Error())
		}
	} else if tr != nil && tr.EndsAt.After(time.Now().UTC()) {
		p.Plan = enums.PlanKeyTrial
		trial = TrialInfo{Active: true, EndsAt: tr.EndsAt, CreditsGranted: tr.CreditsGranted}
	}
	return p, trial, nil
}

// PeriodKey cuts the usage period key in the configured location.
func (s *PolicyService) PeriodKey(period enums.UsagePeriod, now time.Time) string {
	return periodKeyFor(period, now, s.cfg.Location())
}

// Location returns the configured quota-period location. Callers that
// reserve (Task 11, Task 12) must cut period keys with this exact location
// so a hold's keys match the snapshot Decide read.
func (s *PolicyService) Location() *time.Location {
	return s.cfg.Location()
}

// ResetAt is the first instant of the next period, in the configured
// location (spec §7.4).
func (s *PolicyService) ResetAt(period enums.UsagePeriod, now time.Time) time.Time {
	loc := s.cfg.Location()
	local := now.In(loc)
	switch period {
	case enums.UsagePeriodDay:
		return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 1)
	case enums.UsagePeriodTotal:
		return time.Time{} // never resets
	default:
		return time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0)
	}
}

// Decide is the advisory pre-flight (spec §5.5): resolve principal, load
// usage, evaluate purely. Denials carry the coded AppError for the wire;
// free actions never touch the stores.
func (s *PolicyService) Decide(ctx context.Context, userID string, isAdmin bool, action Action) (Decision, error) {
	p, _, err := s.Principal(ctx, userID, isAdmin)
	if err != nil {
		return Decision{}, err
	}
	if len(ActionCharges(action)) == 0 {
		return Evaluate(p, action, nil, nil), nil
	}
	limits, err := s.LimitsForPlan(ctx, p.Plan)
	if err != nil {
		return Decision{}, err
	}
	usage := UsageSnapshot{}
	if !isAdmin {
		rows, err := s.counters.ListByUser(ctx, userID)
		if err != nil {
			appErr := common.ServiceUnavailable("policy unavailable").WithCode(common.CodePolicyUnavailable)
			logger.Report(ctx, "decide usage read failed", appErr, "userID", userID, "cause", err.Error())
			return Decision{}, appErr
		}
		for _, r := range rows {
			lim, ok := limits[r.CounterKey]
			if !ok {
				// Counters absent from the plan snapshot carry no period to
				// cut a key with; Evaluate denies them as feature_not_in_plan
				// anyway, so they must not seed the usage snapshot.
				continue
			}
			if r.PeriodKey == s.PeriodKey(lim.Period, time.Now().UTC()) {
				usage[r.CounterKey] = CounterUsage{Used: r.Used, Reserved: r.Reserved}
			}
		}
	}
	d := Evaluate(p, action, limits, usage)
	slog.InfoContext(ctx, "access decision", "action", string(action), "plan", string(p.Plan),
		"allow", d.Allow, "code", d.Code, "remaining", d.Remaining)
	return d, nil
}

// AppError maps a denial Decision onto the wire error (spec §9 example
// body). Allow decisions return nil.
func AppError(d Decision) *common.AppError {
	if d.Allow {
		return nil
	}
	base := common.Forbidden("this feature requires a premium subscription")
	switch d.Code {
	case common.CodeInsufficientCredits:
		base = common.TooManyRequests("insufficient AI credits")
	case common.CodeQuotaExceeded:
		base = common.TooManyRequests("quota reached")
	case common.CodeFeatureNotInPlan:
		base = common.Forbidden("this feature is not in your plan")
	case common.CodePremiumRequired:
		base = common.Forbidden("this feature requires a premium subscription")
	case common.CodeTrialExpired:
		base = common.Forbidden("your trial has ended")
	}
	return base.WithCode(d.Code).WithDetails(map[string]any{
		"code":         d.Code,
		"feature":      string(d.LimitingCounter),
		"limit":        d.Limit,
		"used":         d.Used,
		"remaining":    d.Remaining,
		"requiredPlan": string(d.RequiredPlan),
		"upgradeUrl":   "/payments/pricing",
	})
}
