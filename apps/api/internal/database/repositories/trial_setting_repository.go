package repositories

import (
	"context"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TrialSettingRepository reads/writes the single-row trial duration setting.
type TrialSettingRepository interface {
	Get(ctx context.Context) (*models.TrialSetting, error)
	Upsert(ctx context.Context, m *models.TrialSetting) error
	WithTx(tx *gorm.DB) TrialSettingRepository
}

type trialSettingRepository struct{ db *gorm.DB }

func NewTrialSettingRepository(db *gorm.DB) TrialSettingRepository {
	return &trialSettingRepository{db: db}
}

func (r *trialSettingRepository) WithTx(tx *gorm.DB) TrialSettingRepository {
	return &trialSettingRepository{db: tx}
}

func (r *trialSettingRepository) Get(ctx context.Context) (*models.TrialSetting, error) {
	var m models.TrialSetting
	if err := r.db.WithContext(ctx).First(&m, "id = true").Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *trialSettingRepository) Upsert(ctx context.Context, m *models.TrialSetting) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(m).Error
}
