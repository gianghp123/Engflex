package repositories

import (
	"context"
	"time"

	"engflex-api/internal/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserVocabularyRepository is the DB access layer for the user_vocabulary table.
//
// The personal library is keyed (user_id, item_id): Upsert saves state without
// resetting the review schedule, and UpdateSchedule advances it after a review.
type UserVocabularyRepository interface {
	Get(ctx context.Context, userID, itemID string) (*models.UserVocabulary, error)
	Upsert(ctx context.Context, u *models.UserVocabulary, updateMastered bool) error
	Patch(ctx context.Context, userID, itemID string, mastered *bool, note *string) error
	UpdateSchedule(ctx context.Context, u *models.UserVocabulary) error
	Delete(ctx context.Context, userID, itemID string) error
	CountByUser(ctx context.Context, userID string) (int64, error)
	CountMastered(ctx context.Context, userID string) (int64, error)
	CountDue(ctx context.Context, userID string, now time.Time) (int64, error)
	// CountSavedBetween counts enrolments created in [from, to) — the caller's
	// local day, so "saved today" is not UTC's today.
	CountSavedBetween(ctx context.Context, userID string, from, to time.Time) (int64, error)
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

// Upsert saves per-user state over a shared item. The INSERT carries the fresh
// schedule (first save sets srs_due_at = now); the conflict UPDATE deliberately
// excludes srs_due_at, interval_days, review_count, last_reviewed_at, so a
// re-save never resets progress. mastered is only updated when the caller
// supplied it.
func (r *userVocabularyRepository) Upsert(ctx context.Context, u *models.UserVocabulary, updateMastered bool) error {
	if _, err := parseID(u.ItemID, "itemId"); err != nil {
		return err
	}
	cols := []string{"source_type", "source_id", "note", "updated_at"}
	if updateMastered {
		cols = append(cols, "mastered")
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "item_id"}},
		DoUpdates: clause.AssignmentColumns(cols),
	}).Create(u).Error
}

// Patch updates only the supplied fields. A map is used (not a struct) so an
// empty-string note is written as a clear rather than skipped as a zero value.
// It never touches the schedule columns.
func (r *userVocabularyRepository) Patch(ctx context.Context, userID, itemID string, mastered *bool, note *string) error {
	if _, err := parseID(itemID, "itemId"); err != nil {
		return err
	}
	updates := map[string]any{}
	if mastered != nil {
		updates["mastered"] = *mastered
	}
	if note != nil {
		updates["note"] = *note
	}
	if len(updates) == 0 {
		// The service rejects an empty patch before reaching here; this is a
		// guard so a future caller cannot issue a no-op UPDATE.
		return nil
	}
	res := r.db.WithContext(ctx).Model(&models.UserVocabulary{}).
		Where("user_id = ? AND item_id = ?", userID, itemID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *userVocabularyRepository) UpdateSchedule(ctx context.Context, u *models.UserVocabulary) error {
	if _, err := parseID(u.ItemID, "itemId"); err != nil {
		return err
	}
	res := r.db.WithContext(ctx).Model(&models.UserVocabulary{}).
		Where("user_id = ? AND item_id = ?", u.UserID, u.ItemID).
		Updates(map[string]any{
			"srs_due_at":       u.SRSDueAt,
			"interval_days":    u.IntervalDays,
			"review_count":     u.ReviewCount,
			"last_reviewed_at": u.LastReviewedAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

func (r *userVocabularyRepository) CountByUser(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.UserVocabulary{}).Where("user_id = ?", userID).Count(&n).Error
	return n, err
}

func (r *userVocabularyRepository) CountMastered(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.UserVocabulary{}).
		Where("user_id = ? AND mastered = true", userID).Count(&n).Error
	return n, err
}

func (r *userVocabularyRepository) CountDue(ctx context.Context, userID string, now time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.UserVocabulary{}).
		Where("user_id = ? AND mastered = false AND srs_due_at IS NOT NULL AND srs_due_at <= ?", userID, now).
		Count(&n).Error
	return n, err
}

// CountSavedBetween counts the caller's enrolments created in [from, to). The
// web sends its local midnight as `from`, so the window is the reader's day.
func (r *userVocabularyRepository) CountSavedBetween(ctx context.Context, userID string, from, to time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.UserVocabulary{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, from, to).
		Count(&n).Error
	return n, err
}
