package repositories

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"engflex-api/internal/database/models"
)

type FeedbackRepository interface {
	Upsert(ctx context.Context, f *models.Feedback) error
	ListRecentByUser(ctx context.Context, userID string, limit int) ([]*models.Feedback, error)
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
			DoUpdates: clause.AssignmentColumns([]string{"payload", "user_id", "updated_at"}),
		}).
		Create(f).Error
}

// ListRecentByUser returns the caller's newest coaching rows for the error
// aggregation endpoints (query-light: capped at 200 by the caller).
func (r *feedbackRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]*models.Feedback, error) {
	if limit <= 0 {
		limit = 200
	}
	var rows []*models.Feedback
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
