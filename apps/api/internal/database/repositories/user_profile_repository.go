package repositories

import (
	"context"

	"gorm.io/gorm"

	"engflex-api/internal/database/models"
)

// UserProfileRepository is the DB access layer for the user_profiles table.
//
// The read surface arrives with the module that needs it: admin user detail
// is the first caller, so GetByUserID is declared now. WithTx is present
// from the start because every repository in this package carries it, and a
// transaction-scoped handle has to exist before the first write does. Add
// methods when a caller appears, not before.
type UserProfileRepository interface {
	GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error)
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

func (r *userProfileRepository) GetByUserID(ctx context.Context, userID string) (*models.UserProfile, error) {
	var m models.UserProfile
	if err := r.db.WithContext(ctx).First(&m, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}
