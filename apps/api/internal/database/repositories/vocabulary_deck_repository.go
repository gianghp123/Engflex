package repositories

import (
	"context"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

// VocabularyDeckRepository is the DB access layer for the vocabulary_decks table.
// Decks are curated (user_id NULL) or private; ListVisible returns both for a
// caller. GetDetail joins the category in one round trip and attaches Items by
// a second query (a deck with many cards must not cartesian-duplicate the
// parent into a join).
type VocabularyDeckRepository interface {
	ListVisible(ctx context.Context, userID string, limit, offset int) ([]*models.VocabularyDeck, error)
	CountVisible(ctx context.Context, userID string) (int64, error)
	// CountByUser counts the caller's own decks only (curated decks are not
	// theirs and never count toward the per-user cap).
	CountByUser(ctx context.Context, userID string) (int64, error)
	GetByID(ctx context.Context, id string) (*models.VocabularyDeck, error)
	GetDetail(ctx context.Context, userID, id string) (*models.VocabularyDeck, error)
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

func (r *vocabularyDeckRepository) ListVisible(ctx context.Context, userID string, limit, offset int) ([]*models.VocabularyDeck, error) {
	out := make([]*models.VocabularyDeck, 0)
	if err := r.db.WithContext(ctx).
		Where("user_id IS NULL OR user_id = ?", userID).
		Order("user_id IS NULL desc, created_at DESC").
		Limit(limit).Offset(offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *vocabularyDeckRepository) CountVisible(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.VocabularyDeck{}).
		Where("user_id IS NULL OR user_id = ?", userID).Count(&n).Error
	return n, err
}

// CountByUser counts the caller's own decks (user_id = ?), never curated ones.
func (r *vocabularyDeckRepository) CountByUser(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.VocabularyDeck{}).
		Where("user_id = ?", userID).Count(&n).Error
	return n, err
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

func (r *vocabularyDeckRepository) GetDetail(ctx context.Context, userID, id string) (*models.VocabularyDeck, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var m models.VocabularyDeck
	if err := r.db.WithContext(ctx).Joins("Category").First(&m, "vocabulary_decks.id = ?", id).Error; err != nil {
		return nil, err
	}
	// Has-many: a Joins("Items") would cartesian-duplicate the deck row.
	// Second constant query, attached in Go — same shape as
	// VideoExerciseRepository.GetDetail attaching Transcripts. The caller's
	// review state rides on each item (attachUserState), so the deck page can
	// show mastery and due counts instead of zeros.
	items := make([]*models.VocabularyDeckItem, 0)
	if err := r.db.WithContext(ctx).
		Joins("Item").
		Where("vocabulary_deck_items.deck_id = ?", id).
		Order("vocabulary_deck_items.sequence asc, vocabulary_deck_items.created_at asc").
		Find(&items).Error; err != nil {
		return nil, err
	}
	if err := attachUserState(ctx, r.db, userID, items); err != nil {
		return nil, err
	}
	m.Items = items
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
