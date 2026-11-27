package repositories

import (
	"context"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

type StoryFilter struct {
	CEFR          string
	Topic         string
	Search        string
	OnlyPublished bool
	Limit         int
	Offset        int
}

type StoryRepository interface {
	List(ctx context.Context, f StoryFilter) ([]*models.Story, error)
	Count(ctx context.Context, f StoryFilter) (int64, error)
	GetByID(ctx context.Context, id string) (*models.Story, error)
	GetBySlug(ctx context.Context, slug string) (*models.Story, error)
	Create(ctx context.Context, s *models.Story) error
	Update(ctx context.Context, s *models.Story) error
	Delete(ctx context.Context, id string) error
	ListTopics(ctx context.Context) ([]string, error)
	// ReplaceSentences swaps a story's entire sentence timeline: admin-pasted
	// JSON has no concept of "which rows changed", so every save is a full
	// replace rather than a diff. Call through WithTx — the delete and the
	// batch insert must commit together or a mid-save failure could leave a
	// story with no sentences at all.
	ReplaceSentences(ctx context.Context, storyID string, sentences []*models.StorySentence) error
	WithTx(tx *gorm.DB) StoryRepository
}

type storyRepository struct {
	db *gorm.DB
}

func NewStoryRepository(db *gorm.DB) StoryRepository {
	return &storyRepository{db: db}
}

func (r *storyRepository) WithTx(tx *gorm.DB) StoryRepository {
	return &storyRepository{db: tx}
}

func applyStoryFilters(q *gorm.DB, f StoryFilter) *gorm.DB {
	if f.CEFR != "" {
		q = q.Where("cefr_level = ?", f.CEFR)
	}
	if f.Topic != "" && f.Topic != "all" {
		q = q.Where("topic = ?", f.Topic)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("title ILIKE ? OR title_vietnamese ILIKE ?", like, like)
	}
	if f.OnlyPublished {
		q = q.Where("status = ?", "published")
	}
	return q
}

func (r *storyRepository) List(ctx context.Context, f StoryFilter) ([]*models.Story, error) {
	out := make([]*models.Story, 0)
	q := applyStoryFilters(r.db.WithContext(ctx).Model(&models.Story{}), f)
	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	if err := q.Order("created_at DESC").Limit(limit).Offset(f.Offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *storyRepository) Count(ctx context.Context, f StoryFilter) (int64, error) {
	var count int64
	q := applyStoryFilters(r.db.WithContext(ctx).Model(&models.Story{}), f)
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *storyRepository) GetByID(ctx context.Context, id string) (*models.Story, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var m models.Story
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	sentences := make([]*models.StorySentence, 0)
	if err := r.db.WithContext(ctx).Where("story_id = ?", m.ID).Order("sequence asc").Find(&sentences).Error; err != nil {
		return nil, err
	}
	m.Sentences = sentences
	return &m, nil
}

func (r *storyRepository) GetBySlug(ctx context.Context, slug string) (*models.Story, error) {
	var m models.Story
	if err := r.db.WithContext(ctx).First(&m, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	sentences := make([]*models.StorySentence, 0)
	if err := r.db.WithContext(ctx).Where("story_id = ?", m.ID).Order("sequence asc").Find(&sentences).Error; err != nil {
		return nil, err
	}
	m.Sentences = sentences
	return &m, nil
}

func (r *storyRepository) Create(ctx context.Context, s *models.Story) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *storyRepository) Update(ctx context.Context, s *models.Story) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *storyRepository) Delete(ctx context.Context, id string) error {
	if _, err := parseID(id, "id"); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Delete(&models.Story{}, "id = ?", id).Error
}

func (r *storyRepository) ReplaceSentences(ctx context.Context, storyID string, sentences []*models.StorySentence) error {
	if err := r.db.WithContext(ctx).Where("story_id = ?", storyID).Delete(&models.StorySentence{}).Error; err != nil {
		return err
	}
	if len(sentences) == 0 {
		return nil
	}
	for _, sentence := range sentences {
		sentence.StoryID = storyID
	}
	return r.db.WithContext(ctx).Create(&sentences).Error
}

func (r *storyRepository) ListTopics(ctx context.Context) ([]string, error) {
	var topics []string
	if err := r.db.WithContext(ctx).Model(&models.Story{}).Where("topic != ''").Distinct().Order("topic ASC").Pluck("topic", &topics).Error; err != nil {
		return nil, err
	}
	return topics, nil
}
