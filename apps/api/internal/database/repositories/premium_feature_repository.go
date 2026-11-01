package repositories

import (
	"context"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// PremiumFeatureRepository is the DB access layer for the premium_features
// table: admin's registry of named app capabilities and whether each one
// currently requires premium.
type PremiumFeatureRepository interface {
	List(ctx context.Context, limit, offset int) ([]*models.PremiumFeature, error)
	Count(ctx context.Context) (int64, error)
	GetByID(ctx context.Context, id string) (*models.PremiumFeature, error)
	Create(ctx context.Context, f *models.PremiumFeature) error
	Update(ctx context.Context, f *models.PremiumFeature) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *gorm.DB) PremiumFeatureRepository
}

type premiumFeatureRepository struct {
	db *gorm.DB
}

// NewPremiumFeatureRepository builds the repository over the shared DB handle.
func NewPremiumFeatureRepository(db *gorm.DB) PremiumFeatureRepository {
	return &premiumFeatureRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *premiumFeatureRepository) WithTx(tx *gorm.DB) PremiumFeatureRepository {
	return &premiumFeatureRepository{db: tx}
}

func (r *premiumFeatureRepository) List(ctx context.Context, limit, offset int) ([]*models.PremiumFeature, error) {
	out := make([]*models.PremiumFeature, 0)
	if err := r.db.WithContext(ctx).Order("key ASC").Limit(limit).Offset(offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *premiumFeatureRepository) Count(ctx context.Context) (int64, error) {
	var n int64
	if err := r.db.WithContext(ctx).Model(&models.PremiumFeature{}).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (r *premiumFeatureRepository) GetByID(ctx context.Context, id string) (*models.PremiumFeature, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var m models.PremiumFeature
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *premiumFeatureRepository) Create(ctx context.Context, f *models.PremiumFeature) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *premiumFeatureRepository) Update(ctx context.Context, f *models.PremiumFeature) error {
	result := r.db.WithContext(ctx).Save(f)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *premiumFeatureRepository) Delete(ctx context.Context, id string) error {
	if _, err := parseID(id, "id"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.PremiumFeature{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
