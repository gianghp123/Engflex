package repositories

import (
	"context"
	"strings"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

type SavedItemRepository interface {
	ListLibrary(ctx context.Context, userID, contentType, search string, isAdmin, isPremium bool, limit, offset int) ([]*models.SavedItem, error)
	CountLibrary(ctx context.Context, userID, contentType, search string, isAdmin, isPremium bool) (int64, error)
	LibraryStats(ctx context.Context, userID string) (*models.SavedLibrary, error)
	GetTarget(ctx context.Context, contentType, contentID string) (*models.SavedItem, error)
}

type savedItemRepository struct {
	db *gorm.DB
}

func NewSavedItemRepository(db *gorm.DB) SavedItemRepository {
	return &savedItemRepository{db: db}
}

func (r *savedItemRepository) library(ctx context.Context, userID string) *gorm.DB {
	return r.db.WithContext(ctx).Model(&models.SavedItem{}).Table(`(
		SELECT content_type, content_id, note, created_at FROM bookmarks WHERE user_id = ?
		UNION ALL
		SELECT 'vocab', item_id, note, created_at FROM user_vocabulary WHERE user_id = ?
	) AS saved_items`, userID, userID)
}

func joinSavedContent(q *gorm.DB, userID string, isAdmin, isPremium bool) *gorm.DB {
	return q.
		Joins(`LEFT JOIN lessons AS "Lesson" ON saved_items.content_type = 'lesson' AND "Lesson".id = saved_items.content_id
			AND (? OR "Lesson".status = 'published') AND (? OR NOT "Lesson".is_premium)`, isAdmin, isAdmin || isPremium).
		Joins(`LEFT JOIN video_exercises AS "VideoExercise" ON saved_items.content_type = 'video' AND "VideoExercise".id = saved_items.content_id
			AND (? OR "VideoExercise".status = 'published') AND (? OR NOT "VideoExercise".is_premium)`, isAdmin, isAdmin || isPremium).
		Joins(`LEFT JOIN stories AS "Story" ON saved_items.content_type = 'story' AND "Story".id = saved_items.content_id`).
		Joins(`LEFT JOIN vocabulary_items AS "VocabularyItem" ON saved_items.content_type = 'vocab' AND "VocabularyItem".id = saved_items.content_id
			AND (? OR "VocabularyItem".status = 'published') AND (? OR NOT "VocabularyItem".is_premium)
			AND ("VocabularyItem".created_by_user_id IS NULL OR "VocabularyItem".created_by_user_id = ?)`, isAdmin, isAdmin || isPremium, userID)
}

func (r *savedItemRepository) filteredLibrary(ctx context.Context, userID, contentType, search string, isAdmin, isPremium bool) *gorm.DB {
	q := joinSavedContent(r.library(ctx, userID), userID, isAdmin, isPremium)
	if contentType != "" {
		q = q.Where("saved_items.content_type = ?", contentType)
	}
	if search = strings.TrimSpace(search); search != "" {
		q = q.Where(`strpos(lower(concat_ws(' ', "Lesson".title, "VideoExercise".title, "Story".title, "VocabularyItem".term, saved_items.note)), lower(?)) > 0`, search)
	}
	return q
}

func (r *savedItemRepository) GetTarget(ctx context.Context, contentType, contentID string) (*models.SavedItem, error) {
	if _, err := parseID(contentID, "contentId"); err != nil {
		return nil, err
	}
	var item models.SavedItem
	source := r.db.WithContext(ctx).Model(&models.SavedItem{}).
		Table("(SELECT ?::text AS content_type, ?::uuid AS content_id) AS saved_items", contentType, contentID)
	err := joinSavedContent(source, "", true, true).
		Select(`saved_items.*, "Story".id AS "Story__id",
			"Lesson".id AS "Lesson__id", "Lesson".status AS "Lesson__status", "Lesson".is_premium AS "Lesson__is_premium",
			"VideoExercise".id AS "VideoExercise__id", "VideoExercise".status AS "VideoExercise__status", "VideoExercise".is_premium AS "VideoExercise__is_premium"`).
		Take(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *savedItemRepository) ListLibrary(ctx context.Context, userID, contentType, search string, isAdmin, isPremium bool, limit, offset int) ([]*models.SavedItem, error) {
	items := make([]*models.SavedItem, 0)
	err := r.filteredLibrary(ctx, userID, contentType, search, isAdmin, isPremium).
		Select(`saved_items.*,
			"Lesson".id AS "Lesson__id", "Lesson".title AS "Lesson__title", "Lesson".slug AS "Lesson__slug", "Lesson".description AS "Lesson__description",
			"VideoExercise".id AS "VideoExercise__id", "VideoExercise".title AS "VideoExercise__title", "VideoExercise".description AS "VideoExercise__description",
			"Story".id AS "Story__id", "Story".title AS "Story__title", "Story".slug AS "Story__slug",
			"VocabularyItem".id AS "VocabularyItem__id", "VocabularyItem".term AS "VocabularyItem__term", "VocabularyItem".definition AS "VocabularyItem__definition"`).
		Order("saved_items.created_at DESC, saved_items.content_type, saved_items.content_id").
		Limit(limit).Offset(offset).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *savedItemRepository) CountLibrary(ctx context.Context, userID, contentType, search string, isAdmin, isPremium bool) (int64, error) {
	var total int64
	err := r.filteredLibrary(ctx, userID, contentType, search, isAdmin, isPremium).Count(&total).Error
	return total, err
}

func (r *savedItemRepository) LibraryStats(ctx context.Context, userID string) (*models.SavedLibrary, error) {
	var library models.SavedLibrary
	if err := r.library(ctx, userID).Select(`count(*) AS total,
		count(*) FILTER (WHERE btrim(note) <> '') AS with_notes,
		count(*) FILTER (WHERE content_type = 'vocab') AS vocab,
		count(*) FILTER (WHERE content_type = 'lesson') AS lesson,
		count(*) FILTER (WHERE content_type = 'video') AS video,
		count(*) FILTER (WHERE content_type = 'story') AS story`).Scan(&library).Error; err != nil {
		return nil, err
	}
	return &library, nil
}
