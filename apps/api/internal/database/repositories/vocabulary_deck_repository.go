package repositories

import (
	"context"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// VocabularyDeckRepository is the DB access layer for the vocabulary_decks table.
// GetDetail joins the category in one round trip. Items page separately via
// VocabularyDeckItemRepository.ListByDeck (a deck with many cards must not
// cartesian-duplicate the parent into a join).
type VocabularyDeckRepository interface {
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.VocabularyDeck, error)
	CountByUser(ctx context.Context, userID string) (int64, error)
	GetByID(ctx context.Context, id string) (*models.VocabularyDeck, error)
	GetDetail(ctx context.Context, id string) (*models.VocabularyDeck, error)
	Create(ctx context.Context, d *models.VocabularyDeck) error
	Update(ctx context.Context, d *models.VocabularyDeck) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *gorm.DB) VocabularyDeckRepository
}

type vocabularyDeckRepository struct {
	db *gorm.DB
}

// NewVocabularyDeckRepository builds the repository over the shared DB handle.
func NewVocabularyDeckRepository(db *gorm.DB) VocabularyDeckRepository {
	return &vocabularyDeckRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *vocabularyDeckRepository) WithTx(tx *gorm.DB) VocabularyDeckRepository {
	return &vocabularyDeckRepository{db: tx}
}

// visibleToUser scopes decks the caller may see: everything they own, plus
// published system decks (IsDefault) curated by admin for every learner.
func visibleToUser(db *gorm.DB, userID string) *gorm.DB {
	return db.Where("user_id = ? OR (is_default = true AND status = ?)", userID, enums.PublishStatusPublished)
}

func (r *vocabularyDeckRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.VocabularyDeck, error) {
	var out []*models.VocabularyDeck
	if err := visibleToUser(r.db.WithContext(ctx), userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *vocabularyDeckRepository) CountByUser(ctx context.Context, userID string) (int64, error) {
	var n int64
	if err := visibleToUser(r.db.WithContext(ctx).Model(&models.VocabularyDeck{}), userID).Count(&n).Error; err != nil {
		return 0, err
	}
	return n, nil
}

func (r *vocabularyDeckRepository) GetByID(ctx context.Context, id string) (*models.VocabularyDeck, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var m models.VocabularyDeck
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *vocabularyDeckRepository) GetDetail(ctx context.Context, id string) (*models.VocabularyDeck, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var m models.VocabularyDeck
	if err := r.db.WithContext(ctx).Joins("Category").First(&m, "vocabulary_decks.id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *vocabularyDeckRepository) Create(ctx context.Context, d *models.VocabularyDeck) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *vocabularyDeckRepository) Update(ctx context.Context, d *models.VocabularyDeck) error {
	result := r.db.WithContext(ctx).Save(d)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *vocabularyDeckRepository) Delete(ctx context.Context, id string) error {
	if _, err := parseID(id, "id"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.VocabularyDeck{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
