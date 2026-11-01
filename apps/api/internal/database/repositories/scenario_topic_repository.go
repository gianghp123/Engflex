package repositories

import (
	"context"

	"gorm.io/gorm"

	"engflex-api/internal/database/models"
)

// ScenarioTopicRepository is the DB access layer for the scenario_topics table.
// The scenarios themselves live in ScenarioRepository.
//
// ListTopicsWithPreview is the reason this file exists as its own repository:
// it is the topic list the catalog page reads, and it fills each topic's
// has-many Scenarios with a per-topic top-k cut. That cut is not expressible
// as a single Joins, so the method reads scenarios (with their personas) as a
// second constant query and groups in Go.
type ScenarioTopicRepository interface {
	GetByID(ctx context.Context, id string) (*models.ScenarioTopic, error)
	GetBySlug(ctx context.Context, slug string) (*models.ScenarioTopic, error)
	ListWithPreview(ctx context.Context, previewK int, difficulty *string, search string) ([]*models.ScenarioTopic, error)
	Create(ctx context.Context, t *models.ScenarioTopic) error
	Update(ctx context.Context, t *models.ScenarioTopic) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *gorm.DB) ScenarioTopicRepository
}

type scenarioTopicRepository struct {
	db *gorm.DB
}

// NewScenarioTopicRepository builds the repository over the shared DB handle.
func NewScenarioTopicRepository(db *gorm.DB) ScenarioTopicRepository {
	return &scenarioTopicRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *scenarioTopicRepository) WithTx(tx *gorm.DB) ScenarioTopicRepository {
	return &scenarioTopicRepository{db: tx}
}

func (r *scenarioTopicRepository) GetByID(ctx context.Context, id string) (*models.ScenarioTopic, error) {
	if _, err := parseID(id, "topicId"); err != nil {
		return nil, err
	}
	var t models.ScenarioTopic
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *scenarioTopicRepository) GetBySlug(ctx context.Context, slug string) (*models.ScenarioTopic, error) {
	var t models.ScenarioTopic
	if err := r.db.WithContext(ctx).First(&t, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// rankedIDs carries the per-topic window function: GORM has no builder for a
// top-N-per-group cut, so the ranking is the one piece of SQL here (aggregates
// belong in the repo per the layering rules).
//
// created_at alone is not a total order (one INSERT stamps every seed row with
// the same now()), so the ranking falls back to id to stay stable across
// requests.
const rankedIDs = `
SELECT ranked.id FROM (
    SELECT s2.id,
           ROW_NUMBER() OVER (PARTITION BY s2.topic_id ORDER BY s2.created_at ASC, s2.id ASC) AS rn
    FROM scenarios s2
    WHERE (CAST(? AS TEXT) IS NULL OR s2.cefr_level = CAST(? AS TEXT))
      AND (? = '' OR s2.title ILIKE '%' || ? || '%' OR s2.objective ILIKE '%' || ? || '%')
) AS ranked WHERE ranked.rn <= ?`

// ListWithPreview returns every topic in position order, each with up to
// previewK eligible scenarios attached — with their partner personas, in rank
// order. A nil difficulty matches any CEFR level and an empty search matches
// any title or objective; both filter which scenarios are eligible for the
// top-k cut. A topic with no eligible scenario keeps an empty Scenarios, so
// the filter chips never flicker.
//
// Deliberate exception to "one method = one query": a single GORM Joins cannot
// fill the has-many (it panics on slice relations), and a Preload cannot limit
// per topic. Two constant queries is not an N+1 — no query count here grows
// with the number of topics.
func (r *scenarioTopicRepository) ListWithPreview(ctx context.Context, previewK int, difficulty *string, search string) ([]*models.ScenarioTopic, error) {
	var dif any
	if difficulty != nil {
		dif = *difficulty
	}
	var topics []*models.ScenarioTopic
	if err := r.db.WithContext(ctx).
		Order("position ASC").
		Find(&topics).Error; err != nil {
		return nil, err
	}
	if len(topics) == 0 {
		return topics, nil
	}
	ids := make([]string, 0, len(topics))
	byID := make(map[string]*models.ScenarioTopic, len(topics))
	for _, t := range topics {
		ids = append(ids, t.ID)
		byID[t.ID] = t
	}
	var scenarios []*models.Scenario
	if err := r.db.WithContext(ctx).
		Model(&models.Scenario{}).
		Joins("Persona").
		Where("scenarios.topic_id IN ?", ids).
		Where("scenarios.id IN ("+rankedIDs+")", dif, dif, search, search, search, previewK).
		Order("scenarios.created_at ASC, scenarios.id ASC").
		Find(&scenarios).Error; err != nil {
		return nil, err
	}
	for _, s := range scenarios {
		t := byID[s.TopicID]
		if t == nil {
			continue
		}
		t.Scenarios = append(t.Scenarios, s)
	}
	return topics, nil
}

func (r *scenarioTopicRepository) Create(ctx context.Context, t *models.ScenarioTopic) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *scenarioTopicRepository) Update(ctx context.Context, t *models.ScenarioTopic) error {
	result := r.db.WithContext(ctx).Save(t)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *scenarioTopicRepository) Delete(ctx context.Context, id string) error {
	if _, err := parseID(id, "topicId"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.ScenarioTopic{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
