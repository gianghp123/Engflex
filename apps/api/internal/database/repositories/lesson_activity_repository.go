package repositories

import (
	"context"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// LessonActivityRepository is the DB access layer for the lesson_activities table.
type LessonActivityRepository interface {
	ListByLesson(ctx context.Context, lessonID string) ([]*models.LessonActivity, error)
	GetByID(ctx context.Context, id string) (*models.LessonActivity, error)
	Create(ctx context.Context, a *models.LessonActivity) error
	Update(ctx context.Context, a *models.LessonActivity) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *gorm.DB) LessonActivityRepository
}

type lessonActivityRepository struct {
	db *gorm.DB
}

// NewLessonActivityRepository builds the repository over the shared DB handle.
func NewLessonActivityRepository(db *gorm.DB) LessonActivityRepository {
	return &lessonActivityRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *lessonActivityRepository) WithTx(tx *gorm.DB) LessonActivityRepository {
	return &lessonActivityRepository{db: tx}
}

func (r *lessonActivityRepository) ListByLesson(ctx context.Context, lessonID string) ([]*models.LessonActivity, error) {
	if _, err := parseID(lessonID, "lessonId"); err != nil {
		return nil, err
	}
	out := make([]*models.LessonActivity, 0)
	if err := r.db.WithContext(ctx).Where("lesson_id = ?", lessonID).Order("part_number asc").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *lessonActivityRepository) GetByID(ctx context.Context, id string) (*models.LessonActivity, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var m models.LessonActivity
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *lessonActivityRepository) Create(ctx context.Context, a *models.LessonActivity) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *lessonActivityRepository) Update(ctx context.Context, a *models.LessonActivity) error {
	result := r.db.WithContext(ctx).Save(a)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *lessonActivityRepository) Delete(ctx context.Context, id string) error {
	if _, err := parseID(id, "id"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.LessonActivity{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
