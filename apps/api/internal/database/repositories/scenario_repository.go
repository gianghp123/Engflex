package repositories

import (
	"context"

	"gorm.io/gorm"

	"engflex-api/internal/database/models"
)

// ScenarioRepository is the DB access layer for the scenarios table.
// Topics live in ScenarioTopicRepository.
//
// GetDetail reads two more tables (Topic, Persona) in one query. It lives here
// because it returns a scenario; the file is named for what comes back, not for
// every table the SQL happens to touch.
type ScenarioRepository interface {
	// List returns scenarios newest-first, or in random() order when random is
	// true (the home card's two surprises).
	List(ctx context.Context, topicID *string, random bool, limit, offset int) ([]*models.Scenario, error)
	Count(ctx context.Context, topicID *string) (int64, error)
	GetByID(ctx context.Context, id string) (*models.Scenario, error)
	GetDetail(ctx context.Context, id string) (*models.Scenario, error)
	WithTx(tx *gorm.DB) ScenarioRepository
}

type scenarioRepository struct {
	db *gorm.DB
}

// NewScenarioRepository builds the repository over the shared DB handle.
func NewScenarioRepository(db *gorm.DB) ScenarioRepository {
	return &scenarioRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *scenarioRepository) WithTx(tx *gorm.DB) ScenarioRepository {
	return &scenarioRepository{db: tx}
}

func (r *scenarioRepository) List(ctx context.Context, topicID *string, random bool, limit, offset int) ([]*models.Scenario, error) {
	q := r.db.WithContext(ctx).Model(&models.Scenario{})
	if topicID != nil && *topicID != "" {
		q = q.Where("topic_id = ?", *topicID)
	}
	if random {
		q = q.Order("random()")
	} else {
		q = q.Order("created_at DESC")
	}
	var out []*models.Scenario
	if err := q.Limit(limit).Offset(offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *scenarioRepository) Count(ctx context.Context, topicID *string) (int64, error) {
	q := r.db.WithContext(ctx).Model(&models.Scenario{})
	if topicID != nil && *topicID != "" {
		q = q.Where("topic_id = ?", *topicID)
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (r *scenarioRepository) GetByID(ctx context.Context, id string) (*models.Scenario, error) {
	if _, err := parseID(id, "scenarioId"); err != nil {
		return nil, err
	}
	var m models.Scenario
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// GetDetail returns the scenario with its topic and partner persona in a
// single query. Joins (not Preload) issues one round trip: Preload fans out
// to a query per association, which is exactly the N+1 this replaces. Both
// joins are LEFT, so a scenario without a persona comes back with nil
// Persona instead of vanishing.
func (r *scenarioRepository) GetDetail(ctx context.Context, id string) (*models.Scenario, error) {
	if _, err := parseID(id, "scenarioId"); err != nil {
		return nil, err
	}
	var m models.Scenario
	if err := r.db.WithContext(ctx).
		Joins("Topic").
		Joins("Persona").
		Where("scenarios.id = ?", id).
		First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
