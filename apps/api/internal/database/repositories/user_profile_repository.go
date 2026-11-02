package repositories

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"engflex-api/internal/database/models"
)

type UserProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error)
	Upsert(ctx context.Context, m *models.UserProfile) error
	WithTx(tx *gorm.DB) UserProfileRepository
}

type userProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) UserProfileRepository {
	return &userProfileRepository{db: db}
}

func (r *userProfileRepository) WithTx(tx *gorm.DB) UserProfileRepository {
	return &userProfileRepository{db: tx}
}

func (r *userProfileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	var m models.UserProfile
	if err := r.db.WithContext(ctx).First(&m, "clerk_user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *userProfileRepository) Upsert(ctx context.Context, m *models.UserProfile) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "clerk_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"level", "goals", "topics", "daily_commitment_min", "onboarding_completed_at", "updated_at"}),
	}).Create(m).Error
}
