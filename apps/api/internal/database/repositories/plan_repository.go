package repositories

import (
	"context"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// PlanRepository is the DB access layer for the plans table: the current
// price of each fixed-duration package. Rows are fixed to the seeded
// enums.SubscriptionPlan values — only List/GetByPlan/Update are needed,
// no Create/Delete.
type PlanRepository interface {
	List(ctx context.Context) ([]*models.Plan, error)
	GetByPlan(ctx context.Context, plan enums.SubscriptionPlan) (*models.Plan, error)
	Update(ctx context.Context, p *models.Plan) error
	WithTx(tx *gorm.DB) PlanRepository
}

type planRepository struct {
	db *gorm.DB
}

// NewPlanRepository builds the repository over the shared DB handle.
func NewPlanRepository(db *gorm.DB) PlanRepository {
	return &planRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *planRepository) WithTx(tx *gorm.DB) PlanRepository {
	return &planRepository{db: tx}
}

func (r *planRepository) List(ctx context.Context) ([]*models.Plan, error) {
	var out []*models.Plan
	if err := r.db.WithContext(ctx).Order("months ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *planRepository) GetByPlan(ctx context.Context, plan enums.SubscriptionPlan) (*models.Plan, error) {
	var m models.Plan
	if err := r.db.WithContext(ctx).First(&m, "plan = ?", plan).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *planRepository) Update(ctx context.Context, p *models.Plan) error {
	result := r.db.WithContext(ctx).Model(&models.Plan{}).
		Where("plan = ?", p.Plan).
		Updates(map[string]any{
			"months":    p.Months,
			"amount":    p.Amount,
			"currency":  p.Currency,
			"is_active": p.IsActive,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
