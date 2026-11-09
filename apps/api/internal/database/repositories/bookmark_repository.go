package repositories

import (
	"context"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookmarkRepository interface {
	ListByUser(ctx context.Context, userID, contentType string) ([]*models.Bookmark, error)
	Get(ctx context.Context, userID, contentType, contentID string) (*models.Bookmark, error)
	Create(ctx context.Context, b *models.Bookmark) error
	UpdateNote(ctx context.Context, userID, contentType, contentID string, note *string) error
	Delete(ctx context.Context, userID, contentType, contentID string) error
	WithTx(tx *gorm.DB) BookmarkRepository
}

type bookmarkRepository struct {
	db *gorm.DB
}

func NewBookmarkRepository(db *gorm.DB) BookmarkRepository {
	return &bookmarkRepository{db: db}
}

func (r *bookmarkRepository) WithTx(tx *gorm.DB) BookmarkRepository {
	return &bookmarkRepository{db: tx}
}

func (r *bookmarkRepository) ListByUser(ctx context.Context, userID, contentType string) ([]*models.Bookmark, error) {
	out := make([]*models.Bookmark, 0)
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if contentType != "" {
		q = q.Where("content_type = ?", contentType)
	}
	if err := q.Order("created_at DESC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *bookmarkRepository) Get(ctx context.Context, userID, contentType, contentID string) (*models.Bookmark, error) {
	if _, err := parseID(contentID, "contentId"); err != nil {
		return nil, err
	}
	var m models.Bookmark
	if err := r.db.WithContext(ctx).First(&m, "user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *bookmarkRepository) Create(ctx context.Context, b *models.Bookmark) error {
	if _, err := parseID(b.ContentID, "contentId"); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(b).Error
}

func (r *bookmarkRepository) Delete(ctx context.Context, userID, contentType, contentID string) error {
	if _, err := parseID(contentID, "contentId"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.Bookmark{}, "user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *bookmarkRepository) UpdateNote(ctx context.Context, userID, contentType, contentID string, note *string) error {
	if _, err := parseID(contentID, "contentId"); err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Model(&models.Bookmark{}).
		Where("user_id = ? AND content_type = ? AND content_id = ?", userID, contentType, contentID).
		Update("note", note)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
