package repositories

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"engflex-api/internal/database/models"
)

// UserProfileRepository is the DB access layer for the user_profiles table.
//
// The read surface arrives with the module that needs it. GetByUserID is the
// first read: the progress rollup reads preferences.dailyCommitmentMin. Add
// methods when a caller appears, not before.
type UserProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error)
	Upsert(ctx context.Context, m *models.UserProfile) error
	WithTx(tx *gorm.DB) UserProfileRepository
}

type userProfileRepository struct {
	db *gorm.DB
}

// NewUserProfileRepository builds the repository over the shared DB handle.
func NewUserProfileRepository(db *gorm.DB) UserProfileRepository {
	return &userProfileRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *userProfileRepository) WithTx(tx *gorm.DB) UserProfileRepository {
	return &userProfileRepository{db: tx}
}

// GetByUserID returns the caller's profile, or gorm.ErrRecordNotFound when
// the caller has not onboarded yet (the service falls back to a default).
func (r *userProfileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	var m models.UserProfile
	if err := r.db.WithContext(ctx).First(&m, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// Upsert inserts the profile or updates it in place, keyed by user_id. Only
// preferences, onboarding_completed_at, and updated_at change on conflict, so
// a re-submit is idempotent and never duplicates a row or resets created_at.
func (r *userProfileRepository) Upsert(ctx context.Context, m *models.UserProfile) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"preferences", "onboarding_completed_at", "updated_at"}),
	}).Create(m).Error
}
