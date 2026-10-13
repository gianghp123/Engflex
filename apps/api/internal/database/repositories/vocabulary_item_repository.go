package repositories

import (
	"context"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// VocabularyItemRepository is the DB access layer for the vocabulary_items table.
//
// The read surface arrives with the module that needs it: the vocabulary
// module now queries this table, so List/Count/GetByID/Create are declared.
// WithTx is present because every repository in this package carries it.
// List/Count take onlyPublished so draft entries never leak into a non-admin
// page (the filter must happen before LIMIT/OFFSET, not after — otherwise
// pages and counts drift). GetByID stays unfiltered; the service decides
// visibility for a single fetch the same way it does for decks.
type VocabularyItemRepository interface {
	List(ctx context.Context, limit, offset int, onlyPublished bool) ([]*models.VocabularyItem, error)
	Count(ctx context.Context, onlyPublished bool) (int64, error)
	GetByID(ctx context.Context, id string) (*models.VocabularyItem, error)
	Create(ctx context.Context, i *models.VocabularyItem) error
	Update(ctx context.Context, i *models.VocabularyItem) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *gorm.DB) VocabularyItemRepository
}

type vocabularyItemRepository struct {
	db *gorm.DB
}

// NewVocabularyItemRepository builds the repository over the shared DB handle.
func NewVocabularyItemRepository(db *gorm.DB) VocabularyItemRepository {
	return &vocabularyItemRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *vocabularyItemRepository) WithTx(tx *gorm.DB) VocabularyItemRepository {
	return &vocabularyItemRepository{db: tx}
}

// visiblePublished scopes to published entries, plus every custom word
// (which is always published on create, but the filter stays explicit
// rather than relying on that invariant never drifting).
func visiblePublished(db *gorm.DB, onlyPublished bool) *gorm.DB {
	if !onlyPublished {
		return db
	}
	return db.Where("status = ?", enums.PublishStatusPublished)
}

func (r *vocabularyItemRepository) List(ctx context.Context, limit, offset int, onlyPublished bool) ([]*models.VocabularyItem, error) {
	var out []*models.VocabularyItem
	q := visiblePublished(r.db.WithContext(ctx), onlyPublished)
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *vocabularyItemRepository) Count(ctx context.Context, onlyPublished bool) (int64, error) {
	var n int64
	q := visiblePublished(r.db.WithContext(ctx).Model(&models.VocabularyItem{}), onlyPublished)
	if err := q.Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (r *vocabularyItemRepository) GetByID(ctx context.Context, id string) (*models.VocabularyItem, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var m models.VocabularyItem
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *vocabularyItemRepository) Create(ctx context.Context, i *models.VocabularyItem) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *vocabularyItemRepository) Update(ctx context.Context, i *models.VocabularyItem) error {
	result := r.db.WithContext(ctx).Save(i)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *vocabularyItemRepository) Delete(ctx context.Context, id string) error {
	if _, err := parseID(id, "id"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.VocabularyItem{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
