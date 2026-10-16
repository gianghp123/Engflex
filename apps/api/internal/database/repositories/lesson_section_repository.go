package repositories

import (
	"context"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// LessonSectionRepository is the DB access layer for the lesson_sections table.
type LessonSectionRepository interface {
	List(ctx context.Context) ([]*models.LessonSection, error)
	GetByID(ctx context.Context, id string) (*models.LessonSection, error)
	ListWithUnits(ctx context.Context, level enums.CEFR, isAdmin bool) ([]*models.LessonSection, error)
	Create(ctx context.Context, s *models.LessonSection) error
	Update(ctx context.Context, s *models.LessonSection) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *gorm.DB) LessonSectionRepository
}

type lessonSectionRepository struct {
	db *gorm.DB
}

// NewLessonSectionRepository builds the repository over the shared DB handle.
func NewLessonSectionRepository(db *gorm.DB) LessonSectionRepository {
	return &lessonSectionRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *lessonSectionRepository) WithTx(tx *gorm.DB) LessonSectionRepository {
	return &lessonSectionRepository{db: tx}
}

func (r *lessonSectionRepository) List(ctx context.Context) ([]*models.LessonSection, error) {
	var out []*models.LessonSection
	if err := r.db.WithContext(ctx).Order("position ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *lessonSectionRepository) GetByID(ctx context.Context, id string) (*models.LessonSection, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var m models.LessonSection
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// ListWithUnits returns sections with their units attached, in path order.
// Two constant queries grouped in Go: a has-many join would cartesian
// the parents and break counts, and the join builder cannot limit per parent.
// Sections left empty by a level filter are dropped for learners — the hub
// renders only sections that have units; an admin keeps every section
// (including a brand-new, still-lesson-less one) and sees draft units too.
func (r *lessonSectionRepository) ListWithUnits(ctx context.Context, level enums.CEFR, isAdmin bool) ([]*models.LessonSection, error) {
	var sections []*models.LessonSection
	if err := r.db.WithContext(ctx).Order("position ASC").Find(&sections).Error; err != nil {
		return nil, err
	}
	lq := r.db.WithContext(ctx).Model(&models.Lesson{}).Joins("Section")
	if level != "" {
		lq = lq.Where("lessons.cefr_level = ?", string(level))
	}
	if !isAdmin {
		lq = lq.Where("lessons.status = ?", enums.PublishStatusPublished)
	}
	var units []*models.Lesson
	if err := lq.Order("lessons.created_at ASC").Find(&units).Error; err != nil {
		return nil, err
	}
	bySection := map[string][]*models.Lesson{}
	for _, u := range units {
		bySection[u.SectionID] = append(bySection[u.SectionID], u)
	}
	out := make([]*models.LessonSection, 0, len(sections))
	for _, s := range sections {
		s.Units = bySection[s.ID]
		if len(s.Units) == 0 && !isAdmin {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

func (r *lessonSectionRepository) Create(ctx context.Context, s *models.LessonSection) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *lessonSectionRepository) Update(ctx context.Context, s *models.LessonSection) error {
	result := r.db.WithContext(ctx).Save(s)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *lessonSectionRepository) Delete(ctx context.Context, id string) error {
	if _, err := parseID(id, "id"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.LessonSection{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
