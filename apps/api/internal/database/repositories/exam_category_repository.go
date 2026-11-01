package repositories

import (
	"context"

	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
)

type ExamCategoryRepository interface {
	List(ctx context.Context, examType enums.ExamType) ([]*models.ExamCategoryCatalog, error)
	GetByID(ctx context.Context, id string) (*models.ExamCategory, error)
}

type examCategoryRepository struct{ db *gorm.DB }

func NewExamCategoryRepository(db *gorm.DB) ExamCategoryRepository {
	return &examCategoryRepository{db: db}
}

func (r *examCategoryRepository) List(ctx context.Context, examType enums.ExamType) ([]*models.ExamCategoryCatalog, error) {
	var categories []*models.ExamCategoryCatalog
	count := r.db.Model(&models.ExamTest{}).Select("COUNT(*)").Where("exam_tests.category_id = exam_categories.id")
	query := r.db.WithContext(ctx).Table("exam_categories").Select("exam_categories.*, (?) AS tests_count", count)
	if examType != "" {
		query = query.Where("exam_categories.type = ?", examType)
	}
	err := query.Order("exam_categories.title ASC").Order("exam_categories.id ASC").Find(&categories).Error
	return categories, err
}

func (r *examCategoryRepository) GetByID(ctx context.Context, id string) (*models.ExamCategory, error) {
	parsed, err := parseID(id, "categoryId")
	if err != nil {
		return nil, err
	}
	var category models.ExamCategory
	if err := r.db.WithContext(ctx).First(&category, "id = ?", parsed).Error; err != nil {
		return nil, err
	}
	return &category, nil
}
