package repositories

import (
	"context"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// UserTrialRepository records and reads the once-per-user trial row. The
// user_id primary key is the guard: Create on an existing row fails with a
// unique violation, which the service maps to 409 (idempotent grant).
type UserTrialRepository interface {
	Get(ctx context.Context, userID string) (*models.UserTrial, error)
	Create(ctx context.Context, m *models.UserTrial) error
	WithTx(tx *gorm.DB) UserTrialRepository
}

type userTrialRepository struct{ db *gorm.DB }

// NewUserTrialRepository builds the repository over the shared DB handle.
func NewUserTrialRepository(db *gorm.DB) UserTrialRepository {
	return &userTrialRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *userTrialRepository) WithTx(tx *gorm.DB) UserTrialRepository {
	return &userTrialRepository{db: tx}
}

func (r *userTrialRepository) Get(ctx context.Context, userID string) (*models.UserTrial, error) {
	var m models.UserTrial
	if err := r.db.WithContext(ctx).First(&m, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *userTrialRepository) Create(ctx context.Context, m *models.UserTrial) error {
	return r.db.WithContext(ctx).Create(m).Error
}
