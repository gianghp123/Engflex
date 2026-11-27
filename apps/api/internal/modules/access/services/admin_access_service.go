package services

import (
	"context"
	"errors"
	"log/slog"
	"sort"

	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/access/dtos/requests"
	"engflex-api/internal/modules/access/dtos/responses"
)

// SnapshotRefresher invalidates the writer's policy snapshot after an
// admin edit (spec §5.6): other replicas converge on the 30s TTL.
type SnapshotRefresher interface {
	SnapshotCatalog(ctx context.Context) error
}

// AdminAccessService serves the admin limits and usage surfaces.
type AdminAccessService struct {
	limits        repositories.PlanFeatureLimitRepository
	counters      repositories.UsageCounterRepository
	events        repositories.UsageEventRepository
	costs         repositories.ActionCreditCostRepository
	trialSettings repositories.TrialSettingRepository
	tx            repositories.TransactionManager
	policy        SnapshotRefresher
}

// NewAdminAccessService builds the admin service.
func NewAdminAccessService(limits repositories.PlanFeatureLimitRepository, counters repositories.UsageCounterRepository, events repositories.UsageEventRepository, costs repositories.ActionCreditCostRepository, trialSettings repositories.TrialSettingRepository, tx repositories.TransactionManager, policy SnapshotRefresher) *AdminAccessService {
	return &AdminAccessService{limits: limits, counters: counters, events: events, costs: costs, trialSettings: trialSettings, tx: tx, policy: policy}
}

// ListLimits returns the whole catalog for the admin table.
func (s *AdminAccessService) ListLimits(ctx context.Context) ([]*models.PlanFeatureLimit, error) {
	rows, err := s.limits.ListAll(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "plan limits")
		logger.Report(ctx, "admin list limits failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "admin limits listed", "rows", len(rows))
	return rows, nil
}

// UpdatePlanLimits validates and batch-upserts one plan's limits, then
// invalidates this process's snapshot immediately.
func (s *AdminAccessService) UpdatePlanLimits(ctx context.Context, plan enums.PlanKey, req requests.UpdatePlanLimits) error {
	if !validPlanKey(plan) {
		return common.BadRequest("unknown plan")
	}
	rows := make([]*models.PlanFeatureLimit, 0, len(req.Counters))
	seen := map[enums.CounterKind]bool{}
	for _, c := range req.Counters {
		if seen[c.CounterKey] {
			return common.UnprocessableEntity("duplicate counter in request")
		}
		seen[c.CounterKey] = true
		rows = append(rows, &models.PlanFeatureLimit{
			PlanKey:    plan,
			CounterKey: c.CounterKey,
			IsEnabled:  c.IsEnabled,
			LimitValue: c.LimitValue,
			Period:     c.Period,
		})
	}
	if len(rows) == 0 {
		return common.BadRequest("counters are required")
	}
	if err := s.limits.UpsertBatch(ctx, rows); err != nil {
		appErr := common.FromDBError(err, "plan limits")
		logger.Report(ctx, "admin update limits failed", appErr, "plan", string(plan))
		return appErr
	}
	if err := s.policy.SnapshotCatalog(ctx); err != nil {
		// The write landed; snapshot refresh is best-effort (TTL converges).
		slog.WarnContext(ctx, "snapshot refresh after admin edit failed", "plan", string(plan))
	}
	slog.InfoContext(ctx, "admin limits updated", "plan", string(plan), "rows", len(rows))
	return nil
}

// UserUsage returns a user's counters and recent usage events (spec §8.6).
func (s *AdminAccessService) UserUsage(ctx context.Context, userID string) (*responses.UserUsage, error) {
	counters, err := s.counters.ListByUser(ctx, userID)
	if err != nil {
		appErr := common.FromDBError(err, "usage counters")
		logger.Report(ctx, "admin usage read failed", appErr, "userID", userID)
		return nil, appErr
	}
	events, err := s.events.ListByUser(ctx, userID, 50)
	if err != nil {
		appErr := common.FromDBError(err, "usage events")
		logger.Report(ctx, "admin events read failed", appErr, "userID", userID)
		return nil, appErr
	}
	out := &responses.UserUsage{Counters: counters, Events: events}
	slog.InfoContext(ctx, "admin usage listed", "userID", userID, "counters", len(counters), "events", len(events))
	return out, nil
}

func validPlanKey(plan enums.PlanKey) bool {
	switch plan {
	case enums.PlanKeyFree, enums.PlanKeyTrial, enums.PlanKey1M, enums.PlanKey3M, enums.PlanKey12M:
		return true
	default:
		return false
	}
}

// ListActionCosts returns the admin-editable per-action credit prices in the
// fixed paid-action order.
func (s *AdminAccessService) ListActionCosts(ctx context.Context) ([]*models.ActionCreditCost, error) {
	rows, err := s.costs.List(ctx)
	if err != nil {
		appErr := common.FromDBError(err, "action costs")
		logger.Report(ctx, "admin list action costs failed", appErr)
		return nil, appErr
	}
	order := map[enums.AccessAction]int{
		enums.AccessActionVoiceStart: 0, enums.AccessActionExamReview: 1, enums.AccessActionStudyPlan: 2,
	}
	sort.SliceStable(rows, func(i, j int) bool { return order[rows[i].Action] < order[rows[j].Action] })
	slog.InfoContext(ctx, "admin action costs listed", "rows", len(rows))
	return rows, nil
}

// UpdateActionCosts validates and batch-upserts all paid action costs, then
// invalidates this process's snapshot immediately.
func (s *AdminAccessService) UpdateActionCosts(ctx context.Context, actor string, req requests.UpdateActionCosts) error {
	rows := make([]*models.ActionCreditCost, 0, len(req.Costs))
	seen := map[enums.AccessAction]bool{}
	for _, c := range req.Costs {
		if _, ok := defaultCreditCosts[Action(c.Action)]; !ok {
			return common.UnprocessableEntity("unknown or free action")
		}
		if c.Credits < 0 || c.Credits > common.MaxPlanLimitValue {
			return common.BadRequest("credits out of range")
		}
		if seen[c.Action] {
			return common.UnprocessableEntity("duplicate action in request")
		}
		seen[c.Action] = true
		rows = append(rows, &models.ActionCreditCost{Action: c.Action, Credits: c.Credits, UpdatedBy: actor})
	}
	if len(rows) != len(paidActions) {
		return common.BadRequest("all paid action costs are required")
	}
	if err := s.costs.UpsertBatch(ctx, rows); err != nil {
		appErr := common.FromDBError(err, "action costs")
		logger.Report(ctx, "admin update action costs failed", appErr, "actor", actor)
		return appErr
	}
	if err := s.policy.SnapshotCatalog(ctx); err != nil {
		slog.WarnContext(ctx, "snapshot refresh after action cost edit failed", "actor", actor)
	}
	slog.InfoContext(ctx, "admin action costs updated", "actor", actor, "rows", len(rows))
	return nil
}

// GetTrialTerms reads the trial duration plus both trial caps (defaults when
// the settings row or caps are missing).
func (s *AdminAccessService) GetTrialTerms(ctx context.Context) (*responses.AdminTrialTerms, error) {
	setting, err := s.trialSettings.Get(ctx)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		appErr := common.FromDBError(err, "trial settings")
		logger.Report(ctx, "admin get trial terms failed", appErr)
		return nil, appErr
	}
	rows, err := s.limits.ListByPlan(ctx, enums.PlanKeyTrial)
	if err != nil {
		appErr := common.FromDBError(err, "trial limits")
		logger.Report(ctx, "admin get trial limits failed", appErr)
		return nil, appErr
	}
	out := &responses.AdminTrialTerms{Days: common.DefaultTrialDays, Credits: common.DefaultTrialCredits, VoiceSessions: common.DefaultTrialVoiceSessions}
	if setting != nil {
		out.Days = setting.DurationDays
		out.UpdatedAt = setting.UpdatedAt
		out.UpdatedBy = setting.UpdatedBy
	}
	for _, r := range rows {
		switch r.CounterKey {
		case enums.CounterKindAICredit:
			out.Credits = r.LimitValue
		case enums.CounterKindVoiceSession:
			out.VoiceSessions = r.LimitValue
		}
	}
	slog.InfoContext(ctx, "admin trial terms read", "days", out.Days, "credits", out.Credits, "voice", out.VoiceSessions)
	return out, nil
}

// UpdateTrialTerms rewrites the trial duration and both trial caps in one
// transaction, then invalidates this process's snapshot immediately.
func (s *AdminAccessService) UpdateTrialTerms(ctx context.Context, actor string, req requests.UpdateTrialTerms) error {
	// Service-level range checks mirror the gin binding ceilings so direct
	// calls fail 4xx instead of reaching the DB CHECK as a 500.
	if req.Days < 1 || req.Days > common.MaxTrialDays {
		return common.BadRequest("days out of range")
	}
	if req.Credits < 0 || req.Credits > common.MaxPlanLimitValue {
		return common.BadRequest("credits out of range")
	}
	if req.VoiceSessions < 0 || req.VoiceSessions > common.MaxPlanLimitValue {
		return common.BadRequest("voice sessions out of range")
	}
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		if err := s.trialSettings.WithTx(tx).Upsert(ctx, &models.TrialSetting{
			ID: true, DurationDays: req.Days, UpdatedBy: actor,
		}); err != nil {
			return err
		}
		return s.limits.WithTx(tx).UpsertBatch(ctx, []*models.PlanFeatureLimit{
			{PlanKey: enums.PlanKeyTrial, CounterKey: enums.CounterKindAICredit, IsEnabled: true, LimitValue: req.Credits, Period: enums.UsagePeriodTotal},
			{PlanKey: enums.PlanKeyTrial, CounterKey: enums.CounterKindVoiceSession, IsEnabled: true, LimitValue: req.VoiceSessions, Period: enums.UsagePeriodTotal},
		})
	})
	if err != nil {
		appErr := common.FromDBError(err, "trial terms")
		logger.Report(ctx, "admin update trial terms failed", appErr, "actor", actor)
		return appErr
	}
	if err := s.policy.SnapshotCatalog(ctx); err != nil {
		slog.WarnContext(ctx, "snapshot refresh after trial edit failed", "actor", actor)
	}
	slog.InfoContext(ctx, "admin trial terms updated", "actor", actor, "days", req.Days, "credits", req.Credits, "voice", req.VoiceSessions)
	return nil
}
