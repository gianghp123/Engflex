package services

import (
	"context"
	"log/slog"

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
	SnapshotLimits(ctx context.Context) error
}

// AdminAccessService serves the admin limits and usage surfaces.
type AdminAccessService struct {
	limits   repositories.PlanFeatureLimitRepository
	counters repositories.UsageCounterRepository
	events   repositories.UsageEventRepository
	policy   SnapshotRefresher
}

// NewAdminAccessService builds the admin service.
func NewAdminAccessService(limits repositories.PlanFeatureLimitRepository, counters repositories.UsageCounterRepository, events repositories.UsageEventRepository, policy SnapshotRefresher) *AdminAccessService {
	return &AdminAccessService{limits: limits, counters: counters, events: events, policy: policy}
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
	if err := s.policy.SnapshotLimits(ctx); err != nil {
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
