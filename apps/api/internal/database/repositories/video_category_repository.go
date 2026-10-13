package repositories

import (
	"context"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

type VideoCategoryRepository interface {
	List(ctx context.Context, limit, offset int) ([]*models.VideoCategory, error)
	ListWithPreview(ctx context.Context, previewSize int) ([]*models.VideoCategory, error)
	GetByID(ctx context.Context, id string) (*models.VideoCategory, error)
	Create(ctx context.Context, category *models.VideoCategory) error
	Update(ctx context.Context, category *models.VideoCategory) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
	WithTx(tx *gorm.DB) VideoCategoryRepository
}

type videoCategoryRepository struct {
	db *gorm.DB
}

func NewVideoCategoryRepository(db *gorm.DB) VideoCategoryRepository {
	return &videoCategoryRepository{db: db}
}
func (r *videoCategoryRepository) WithTx(tx *gorm.DB) VideoCategoryRepository {
	return &videoCategoryRepository{db: tx}
}

func (r *videoCategoryRepository) Create(ctx context.Context, category *models.VideoCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *videoCategoryRepository) Update(ctx context.Context, category *models.VideoCategory) error {
	result := r.db.WithContext(ctx).Save(category)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *videoCategoryRepository) Delete(ctx context.Context, id string) error {
	_, err := parseID(id, "id")
	if err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.VideoCategory{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *videoCategoryRepository) GetByID(ctx context.Context, id string) (*models.VideoCategory, error) {
	var category models.VideoCategory
	_, err := parseID(id, "id")
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).First(&category, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *videoCategoryRepository) List(ctx context.Context, limit int, offset int) ([]*models.VideoCategory, error) {
	categories := make([]*models.VideoCategory, 0)
	query := r.db.WithContext(ctx).Order("id asc")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	err := query.Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *videoCategoryRepository) ListWithPreview(ctx context.Context, previewSize int) ([]*models.VideoCategory, error) {
		categories := make([]*models.VideoCategory, 0)
	query := r.db.WithContext(ctx).Select("video_categories.*, (SELECT COUNT(*) FROM video_exercises e WHERE e.category_id = video_categories.id) AS video_count").
	Order("video_categories.name ASC")
	if previewSize > 0 {
		query = query.Limit(previewSize)
	}
	err := query.Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *videoCategoryRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.VideoCategory{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
