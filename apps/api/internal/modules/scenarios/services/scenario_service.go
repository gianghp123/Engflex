package services

import (
	"context"
	"log/slog"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
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

// List pages catalog scenarios, optionally topic-filtered. isAdmin widens
// the result to include drafts — every other caller only ever sees
// published content.
func (s *ScenarioService) List(ctx context.Context, req requests.ListScenarios, isAdmin bool) ([]*models.Scenario, int64, error) {
	page, pageSize := req.Page, req.PageSize
	if page < 1 {
		page = common.DefaultPage
	}
	if pageSize < 1 {
		pageSize = common.DefaultPageSize
	}
	offset := utils.Offset(page, pageSize)

	items, err := s.scenarios.List(ctx, req.TopicID, !isAdmin, pageSize, offset)
	if err != nil {
		appErr := common.FromDBError(err, "scenario")
		logger.Report(ctx, "list scenarios failed", appErr)
		return nil, 0, appErr
	}
	total, err := s.scenarios.Count(ctx, req.TopicID, !isAdmin)
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
// persona comes back nil (generic partner label) rather than failing. A
// draft scenario is invisible to non-admins (404, existence not disclosed),
// same rule as vocabulary items/decks and video exercises.
func (s *ScenarioService) GetDetail(ctx context.Context, id string, isAdmin bool) (*models.Scenario, error) {
	m, err := s.scenarios.GetDetail(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "scenario")
		logger.Report(ctx, "get scenario detail failed", appErr, "scenarioID", id)
		return nil, appErr
	}
	if m.Status != enums.PublishStatusPublished && !isAdmin {
		return nil, common.NotFound("scenario not found")
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

// toScenarioDetail hand-maps the wire input into the model's jsonb-backed
// detail struct. copier's DeepCopy does NOT recurse into a nested field
// when the source and destination are different named struct types, even
// with identical inner field names (verified: a blanket utils.Map left
// Details entirely zeroed) — so this one nested field needs an explicit
// line-up instead of the usual blanket utils.Map.
func toScenarioDetail(in requests.ScenarioDetailInput) models.ScenarioDetail {
	return models.ScenarioDetail{
		Role:             in.Role,
		InterlocutorRole: in.InterlocutorRole,
		GoalFormat:       in.GoalFormat,
		Context:          in.Context,
		Opening:          models.DetailOpening{Label: in.Opening.Label, Text: in.Opening.Text},
		Vocab:            in.Vocab,
	}
}

// Create adds an admin-curated scenario, starting as draft until explicitly
// published.
func (s *ScenarioService) Create(ctx context.Context, req requests.CreateScenario) (*models.Scenario, error) {
	m := &models.Scenario{}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map create scenario failed", err)
		return nil, common.Internal()
	}
	m.Details = toScenarioDetail(req.Details)
	m.Status = enums.PublishStatusDraft
	if err := s.scenarios.Create(ctx, m); err != nil {
		appErr := common.FromDBError(err, "scenario")
		logger.Report(ctx, "create scenario failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "scenario created", "scenarioID", m.ID)
	return m, nil
}

// Update edits a scenario; only non-nil fields apply. copier skips a nil
// pointer source rather than zeroing the destination (verified for the
// vocabulary deck/item Update methods, same pattern here).
func (s *ScenarioService) Update(ctx context.Context, id string, req requests.UpdateScenario) (*models.Scenario, error) {
	m, err := s.scenarios.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "scenario")
		logger.Report(ctx, "get scenario for update failed", appErr, "id", id)
		return nil, appErr
	}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map update scenario failed", err, "id", id)
		return nil, common.Internal()
	}
	if req.Details != nil {
		m.Details = toScenarioDetail(*req.Details)
	}
	if req.Status != nil {
		m.Status = *req.Status
	}
	if err := s.scenarios.Update(ctx, m); err != nil {
		appErr := common.FromDBError(err, "scenario")
		logger.Report(ctx, "update scenario failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "scenario updated", "scenarioID", id)
	return m, nil
}

func (s *ScenarioService) Delete(ctx context.Context, id string) error {
	if err := s.scenarios.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "scenario")
		logger.Report(ctx, "delete scenario failed", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "scenario deleted", "id", id)
	return nil
}

// CreateTopic adds a banner taxonomy entry. Topics have no draft concept —
// visible to everyone immediately (same as video categories).
func (s *ScenarioService) CreateTopic(ctx context.Context, req requests.CreateScenarioTopic) (*models.ScenarioTopic, error) {
	t := &models.ScenarioTopic{}
	if err := utils.Map(t, req); err != nil {
		logger.Report(ctx, "map create scenario topic failed", err)
		return nil, common.Internal()
	}
	if err := s.topics.Create(ctx, t); err != nil {
		appErr := common.FromDBError(err, "scenario topic")
		logger.Report(ctx, "create scenario topic failed", appErr)
		return nil, appErr
	}
	slog.InfoContext(ctx, "scenario topic created", "topicID", t.ID)
	return t, nil
}

func (s *ScenarioService) UpdateTopic(ctx context.Context, id string, req requests.UpdateScenarioTopic) (*models.ScenarioTopic, error) {
	t, err := s.topics.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "scenario topic")
		logger.Report(ctx, "get scenario topic for update failed", appErr, "id", id)
		return nil, appErr
	}
	if err := utils.Map(t, req); err != nil {
		logger.Report(ctx, "map update scenario topic failed", err, "id", id)
		return nil, common.Internal()
	}
	if err := s.topics.Update(ctx, t); err != nil {
		appErr := common.FromDBError(err, "scenario topic")
		logger.Report(ctx, "update scenario topic failed", appErr, "id", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "scenario topic updated", "topicID", id)
	return t, nil
}

func (s *ScenarioService) DeleteTopic(ctx context.Context, id string) error {
	if err := s.topics.Delete(ctx, id); err != nil {
		appErr := common.FromDBError(err, "scenario topic")
		logger.Report(ctx, "delete scenario topic failed", appErr, "id", id)
		return appErr
	}
	slog.InfoContext(ctx, "scenario topic deleted", "id", id)
	return nil
}
