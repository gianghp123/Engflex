package repositories

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"engflex-api/internal/database/models"
)

type FeedbackRepository interface {
	Upsert(ctx context.Context, f *models.Feedback) error
	WithTx(tx *gorm.DB) FeedbackRepository
}

type feedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) FeedbackRepository {
	return &feedbackRepository{db: db}
}

func (r *feedbackRepository) WithTx(tx *gorm.DB) FeedbackRepository {
	return &feedbackRepository{db: tx}
}

func (r *feedbackRepository) Upsert(ctx context.Context, f *models.Feedback) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "subject_type"}, {Name: "subject_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"payload", "input_fingerprint", "user_id", "updated_at"}),
		}).
		Create(f).Error
}
