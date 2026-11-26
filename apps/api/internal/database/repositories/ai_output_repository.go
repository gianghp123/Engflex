package repositories

import (
	"context"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// AIOutputRepository reads and stores cached paid AI outputs. The unique
// (subject_type, subject_id, kind) makes Create-on-existing a 409 the
// service maps to the free cached read.
type AIOutputRepository interface {
	GetBySubject(ctx context.Context, subjectType enums.UsageRefType, subjectID string, kind enums.AIOutputKind) (*models.AIOutput, error)
	Create(ctx context.Context, m *models.AIOutput) error
	WithTx(tx *gorm.DB) AIOutputRepository
}

type aiOutputRepository struct{ db *gorm.DB }

func NewAIOutputRepository(db *gorm.DB) AIOutputRepository {
	return &aiOutputRepository{db: db}
}

func (r *aiOutputRepository) WithTx(tx *gorm.DB) AIOutputRepository {
	return &aiOutputRepository{db: tx}
}

func (r *aiOutputRepository) GetBySubject(ctx context.Context, subjectType enums.UsageRefType, subjectID string, kind enums.AIOutputKind) (*models.AIOutput, error) {
	var m models.AIOutput
	if err := r.db.WithContext(ctx).
		Where("subject_type = ? AND subject_id = ? AND kind = ?", subjectType, subjectID, kind).
		First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *aiOutputRepository) Create(ctx context.Context, m *models.AIOutput) error {
	return r.db.WithContext(ctx).Create(m).Error
}
