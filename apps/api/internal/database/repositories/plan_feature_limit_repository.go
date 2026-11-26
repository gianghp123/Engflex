package repositories

import (
	"context"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PlanFeatureLimitRepository reads and batch-upserts the entitlement
// catalog. The policy snapshot reads ListAll on a TTL; admin limit edits
// write through UpsertBatch and invalidate the writer's snapshot.
type PlanFeatureLimitRepository interface {
	ListAll(ctx context.Context) ([]*models.PlanFeatureLimit, error)
	ListByPlan(ctx context.Context, plan enums.PlanKey) ([]*models.PlanFeatureLimit, error)
	UpsertBatch(ctx context.Context, rows []*models.PlanFeatureLimit) error
	WithTx(tx *gorm.DB) PlanFeatureLimitRepository
}

type planFeatureLimitRepository struct{ db *gorm.DB }

// NewPlanFeatureLimitRepository builds the repository over the shared DB handle.
func NewPlanFeatureLimitRepository(db *gorm.DB) PlanFeatureLimitRepository {
	return &planFeatureLimitRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *planFeatureLimitRepository) WithTx(tx *gorm.DB) PlanFeatureLimitRepository {
	return &planFeatureLimitRepository{db: tx}
}

func (r *planFeatureLimitRepository) ListAll(ctx context.Context) ([]*models.PlanFeatureLimit, error) {
	out := make([]*models.PlanFeatureLimit, 0)
	if err := r.db.WithContext(ctx).Order("plan_key ASC, counter_key ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *planFeatureLimitRepository) ListByPlan(ctx context.Context, plan enums.PlanKey) ([]*models.PlanFeatureLimit, error) {
	out := make([]*models.PlanFeatureLimit, 0)
	if err := r.db.WithContext(ctx).Where("plan_key = ?", plan).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// UpsertBatch writes a whole plan's limits (or all plans) with
// created_at/updated_at maintained by GORM. The composite PK makes each
// (plan, counter) an upsert target; the caller validates values first.
func (r *planFeatureLimitRepository) UpsertBatch(ctx context.Context, rows []*models.PlanFeatureLimit) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&rows).Error
}
