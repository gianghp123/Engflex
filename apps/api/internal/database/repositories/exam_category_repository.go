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
	query := r.catalogQuery(ctx)
	if examType != "" {
		query = query.Where("type = ?", examType)
	}
	err := query.Order("title ASC").Order("id ASC").Find(&categories).Error
	return categories, err
}

func (r *examCategoryRepository) GetByID(ctx context.Context, id string) (*models.ExamCategory, error) {
	parsed, err := parseID(id, "categoryId")
	if err != nil {
		return nil, err
	}
	var category models.ExamCategory
	if err := r.db.WithContext(ctx).First(&category, "exam_categories.id = ?", parsed).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

// Correlated aggregates preserve one category per row, including empty ones.
func (r *examCategoryRepository) catalogQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Select(`exam_categories.*,
		(SELECT count(*) FROM exam_tests t WHERE t.category_id = exam_categories.id) AS tests_count,
		(SELECT count(*) FROM attempts a JOIN exam_tests t ON t.id = a.exam_test_id
		 WHERE t.category_id = exam_categories.id AND a.type = 'exam') AS attempts_count`)
}
