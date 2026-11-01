package repositories

import (
	"context"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserVocabularyRepository is the DB access layer for the user_vocabulary table.
//
// The read surface arrives with the module that needs it: the vocabulary
// module now queries this table, so the methods below are declared.
type UserVocabularyRepository interface {
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.UserVocabulary, error)
	Get(ctx context.Context, userID, itemID string) (*models.UserVocabulary, error)
	Upsert(ctx context.Context, u *models.UserVocabulary) error
	Delete(ctx context.Context, userID, itemID string) error
	WithTx(tx *gorm.DB) UserVocabularyRepository
}

type userVocabularyRepository struct {
	db *gorm.DB
}

// NewUserVocabularyRepository builds the repository over the shared DB handle.
func NewUserVocabularyRepository(db *gorm.DB) UserVocabularyRepository {
	return &userVocabularyRepository{db: db}
}

// WithTx returns the repository bound to the given transaction handle.
func (r *userVocabularyRepository) WithTx(tx *gorm.DB) UserVocabularyRepository {
	return &userVocabularyRepository{db: tx}
}

func (r *userVocabularyRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*models.UserVocabulary, error) {
	var out []*models.UserVocabulary
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *userVocabularyRepository) Get(ctx context.Context, userID, itemID string) (*models.UserVocabulary, error) {
	if _, err := parseID(itemID, "itemId"); err != nil {
		return nil, err
	}
	var m models.UserVocabulary
	if err := r.db.WithContext(ctx).First(&m, "user_id = ? AND item_id = ?", userID, itemID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

// Upsert saves per-user state over a shared item, replacing any previous
// row. The (user_id, item_id) PK makes re-saving idempotent.
func (r *userVocabularyRepository) Upsert(ctx context.Context, u *models.UserVocabulary) error {
	if _, err := parseID(u.ItemID, "itemId"); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "item_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"source_type", "source_id", "note", "mastered", "srs_due_at", "updated_at"}),
	}).Create(u).Error
}

func (r *userVocabularyRepository) Delete(ctx context.Context, userID, itemID string) error {
	if _, err := parseID(itemID, "itemId"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.UserVocabulary{}, "user_id = ? AND item_id = ?", userID, itemID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
