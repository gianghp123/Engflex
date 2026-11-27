package repositories

import (
	"context"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ActionCreditCostRepository reads and batch-upserts the admin-editable
// per-action credit prices. The seeded three rows are updated, never created
// or deleted by the app.
type ActionCreditCostRepository interface {
	List(ctx context.Context) ([]*models.ActionCreditCost, error)
	UpsertBatch(ctx context.Context, rows []*models.ActionCreditCost) error
	WithTx(tx *gorm.DB) ActionCreditCostRepository
}

type actionCreditCostRepository struct{ db *gorm.DB }

func NewActionCreditCostRepository(db *gorm.DB) ActionCreditCostRepository {
	return &actionCreditCostRepository{db: db}
}

func (r *actionCreditCostRepository) WithTx(tx *gorm.DB) ActionCreditCostRepository {
	return &actionCreditCostRepository{db: tx}
}

func (r *actionCreditCostRepository) List(ctx context.Context) ([]*models.ActionCreditCost, error) {
	out := make([]*models.ActionCreditCost, 0)
	if err := r.db.WithContext(ctx).Order("action ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *actionCreditCostRepository) UpsertBatch(ctx context.Context, rows []*models.ActionCreditCost) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&rows).Error
}
