package repositories

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"engflex-api/internal/database/models"
)

type ExamTestRepository interface {
	GetByID(ctx context.Context, id string) (*models.ExamTestCatalog, error)
	ListByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*models.ExamTestCatalog, error)
	CountByCategory(ctx context.Context, categoryID string) (int64, error)
	GetForUpdate(ctx context.Context, id string) (*models.ExamTest, error)
	WithTx(tx *gorm.DB) ExamTestRepository
}

type examTestRepository struct{ db *gorm.DB }

func NewExamTestRepository(db *gorm.DB) ExamTestRepository { return &examTestRepository{db: db} }
func (r *examTestRepository) WithTx(tx *gorm.DB) ExamTestRepository {
	return &examTestRepository{db: tx}
}

func (r *examTestRepository) ListByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*models.ExamTestCatalog, error) {
	id, err := parseID(categoryID, "categoryId")
	if err != nil {
		return nil, err
	}
	var tests []*models.ExamTestCatalog
	err = r.catalogQuery(ctx).Where("exam_tests.category_id = ?", id).Order("exam_tests.title ASC").Order("exam_tests.id ASC").Limit(limit).Offset(offset).Find(&tests).Error
	return tests, err
}

func (r *examTestRepository) CountByCategory(ctx context.Context, categoryID string) (int64, error) {
	id, err := parseID(categoryID, "categoryId")
	if err != nil {
		return 0, err
	}
	var count int64
	err = r.db.WithContext(ctx).Model(&models.ExamTest{}).Where("category_id = ?", id).Count(&count).Error
	return count, err
}

func (r *examTestRepository) GetByID(ctx context.Context, id string) (*models.ExamTestCatalog, error) {
	parsed, err := parseID(id, "testId")
	if err != nil {
		return nil, err
	}
	var test models.ExamTestCatalog
	if err := r.catalogQuery(ctx).Joins("Category").First(&test, "exam_tests.id = ?", parsed).Error; err != nil {
		return nil, err
	}
	return &test, nil
}

func (r *examTestRepository) catalogQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Select(`exam_tests.*,
		(SELECT count(*) FROM attempts a WHERE a.exam_test_id = exam_tests.id AND a.type = 'exam') AS attempts_count`)
}

// GetForUpdate is called inside the start transaction; lock only the test row.
func (r *examTestRepository) GetForUpdate(ctx context.Context, id string) (*models.ExamTest, error) {
	parsed, err := parseID(id, "testId")
	if err != nil {
		return nil, err
	}
	var test models.ExamTest
	err = r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: clause.CurrentTable}}).Joins("Category").First(&test, "exam_tests.id = ?", parsed).Error
	if err != nil {
		return nil, err
	}
	return &test, nil
}
