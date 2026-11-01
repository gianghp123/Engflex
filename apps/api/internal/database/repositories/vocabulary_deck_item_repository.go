package repositories

import (
	"context"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// VocabularyDeckItemRepository is the DB access layer for the vocabulary_deck_items table.
type VocabularyDeckItemRepository interface {
	ListByDeck(ctx context.Context, deckID string) ([]*models.VocabularyDeckItem, error)
	GetByID(ctx context.Context, id string) (*models.VocabularyDeckItem, error)
	Create(ctx context.Context, i *models.VocabularyDeckItem) error
	Update(ctx context.Context, i *models.VocabularyDeckItem) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *gorm.DB) VocabularyDeckItemRepository
}

type vocabularyDeckItemRepository struct {
	db *gorm.DB
}

// NewVocabularyDeckItemRepository builds the repository over the shared DB handle.
func NewVocabularyDeckItemRepository(db *gorm.DB) VocabularyDeckItemRepository {
	return &vocabularyDeckItemRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *vocabularyDeckItemRepository) WithTx(tx *gorm.DB) VocabularyDeckItemRepository {
	return &vocabularyDeckItemRepository{db: tx}
}

func (r *vocabularyDeckItemRepository) ListByDeck(ctx context.Context, deckID string) ([]*models.VocabularyDeckItem, error) {
	if _, err := parseID(deckID, "deckId"); err != nil {
		return nil, err
	}
	out := make([]*models.VocabularyDeckItem, 0)
	if err := r.db.WithContext(ctx).Where("deck_id = ?", deckID).Order("created_at asc").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *vocabularyDeckItemRepository) GetByID(ctx context.Context, id string) (*models.VocabularyDeckItem, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var m models.VocabularyDeckItem
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *vocabularyDeckItemRepository) Create(ctx context.Context, i *models.VocabularyDeckItem) error {
	if _, err := parseID(i.DeckID, "deckId"); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *vocabularyDeckItemRepository) Update(ctx context.Context, i *models.VocabularyDeckItem) error {
	result := r.db.WithContext(ctx).Save(i)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *vocabularyDeckItemRepository) Delete(ctx context.Context, id string) error {
	if _, err := parseID(id, "id"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.VocabularyDeckItem{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
