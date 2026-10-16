package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/scenarios/dtos/requests"
	"engflex-api/internal/utils"
)

// ScenarioService owns scenario browsing.
type ScenarioService struct {
	scenarios repositories.ScenarioRepository
	topics    repositories.ScenarioTopicRepository
}

// NewScenarioService builds the service over the repository interfaces.
func NewScenarioService(scenarios repositories.ScenarioRepository, topics repositories.ScenarioTopicRepository) *ScenarioService {
	return &ScenarioService{scenarios: scenarios, topics: topics}
}

// List pages catalog scenarios, optionally topic-filtered. Every row is a
// seeded catalog entry, so the caller's identity never narrows the result.
func (s *ScenarioService) List(ctx context.Context, req requests.ListScenarios) ([]*models.Scenario, int64, error) {
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = common.DefaultPage
	}
	if pageSize < 1 {
		pageSize = common.DefaultPageSize
	}
	offset := utils.Offset(page, pageSize)

	items, err := s.scenarios.List(ctx, req.TopicID, req.Random, pageSize, offset)
	if err != nil {
		appErr := common.FromDBError(err, "scenario")
		logger.Report(ctx, "list scenarios failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.scenarios.Count(ctx, req.TopicID)
	if err != nil {
		appErr := common.FromDBError(err, "scenario")
		logger.Report(ctx, "count scenarios failed", appErr)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "scenarios listed", "count", len(items), "total", total)
	return items, total, nil
}

// ListTopicsWithPreview returns every topic with up to previewK scenarios
// attached. previewK is clamped (default 3, max 6): the browser is
// preview-only, full paging stays on List.
func (s *ScenarioService) ListTopicsWithPreview(ctx context.Context, previewK int, difficulty *string, search string) ([]*models.ScenarioTopic, error) {
	if previewK < 1 {
		previewK = 3
	}
	if previewK > 6 {
		previewK = 6
	}
	out, err := s.topics.ListWithPreview(ctx, previewK, difficulty, search)
	if err != nil {
		appErr := common.FromDBError(err, "scenario topic")
		logger.Report(ctx, "list topics with preview failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "scenario topics with preview listed", "count", len(out))
	return out, nil
}

// GetDetail returns the scenario with its topic and partner persona for
// the detail page, in one joined query. Unknown ids map to 404; a missing
// persona comes back nil (generic partner label) rather than failing.
func (s *ScenarioService) GetDetail(ctx context.Context, id string) (*models.Scenario, error) {
	m, err := s.scenarios.GetDetail(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "scenario")
		logger.Report(ctx, "get scenario detail failed", appErr, "scenarioID", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "scenario detail found", "scenarioID", m.ID)
	return m, nil
}

// GetByID returns one scenario for the room countdown. Unknown ids map to
// 404 via FromDBError; malformed ids map to 400 via the repo's parseID.
func (s *ScenarioService) GetByID(ctx context.Context, id string) (*models.Scenario, error) {
	m, err := s.scenarios.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "scenario")
		logger.Report(ctx, "get scenario failed", appErr, "scenarioID", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "scenario found", "scenarioID", m.ID)
	return m, nil
}
