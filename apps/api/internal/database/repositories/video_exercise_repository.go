package repositories

import (
	"context"
	"engflex-api/internal/database/models"

	"engflex-api/internal/common/enums"
	"gorm.io/gorm"
)

// List/Count take onlyPublished so draft exercises never leak into a
// non-admin page (filtered before LIMIT/OFFSET so pages and counts agree).
// GetByID/GetDetail stay unfiltered; the service decides visibility for a
// single fetch, same as vocabulary items/decks.
type VideoExerciseRepository interface {
	List(ctx context.Context, limit int, offset int, onlyPublished bool) ([]*models.VideoExercise, error)
	GetByID(ctx context.Context, id string) (*models.VideoExercise, error)
	// GetDetail returns the exercise with its category (one join) and its
	// transcripts (second query, ordered by sequence) attached.
	GetDetail(ctx context.Context, id string) (*models.VideoExercise, error)
	Create(ctx context.Context, lesson *models.VideoExercise) error
	Update(ctx context.Context, lesson *models.VideoExercise) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, onlyPublished bool) (int64, error)
	WithTx(tx *gorm.DB) VideoExerciseRepository
}

type videoExerciseRepository struct {
	db *gorm.DB
}

func NewVideoExerciseRepository(db *gorm.DB) VideoExerciseRepository {
	return &videoExerciseRepository{db: db}
}
func (r *videoExerciseRepository) WithTx(tx *gorm.DB) VideoExerciseRepository {
	return &videoExerciseRepository{db: tx}
}

func visiblePublishedVideos(db *gorm.DB, onlyPublished bool) *gorm.DB {
	if !onlyPublished {
		return db
	}
	return db.Where("status = ?", enums.PublishStatusPublished)
}

func (r *videoExerciseRepository) List(ctx context.Context, limit int, offset int, onlyPublished bool) ([]*models.VideoExercise, error) {
	lessons := make([]*models.VideoExercise, 0)
	query := visiblePublishedVideos(r.db.WithContext(ctx), onlyPublished).Order("id asc")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}
	err := query.Find(&lessons).Error
	if err != nil {
		return nil, err
	}
	return lessons, nil
}

func (r *videoExerciseRepository) GetByID(ctx context.Context, id string) (*models.VideoExercise, error) {
	var lesson models.VideoExercise
	_, err := parseID(id, "id")
	if err != nil {
		return nil, err
	}
	err = r.db.WithContext(ctx).First(&lesson, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &lesson, nil
}

// GetDetail returns one exercise with its category joined in a single
// round trip and its transcripts attached from a second query ordered by
// sequence. Transcripts stay out of the join: a has-many join would
// cartesian-duplicate the parent and break pagination.
func (r *videoExerciseRepository) GetDetail(ctx context.Context, id string) (*models.VideoExercise, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var lesson models.VideoExercise
	if err := r.db.WithContext(ctx).Joins("Category").First(&lesson, "video_exercises.id = ?", id).Error; err != nil {
		return nil, err
	}
	transcripts := make([]*models.VideoTranscript, 0)
	if err := r.db.WithContext(ctx).
		Where("video_exercise_id = ?", id).
		Order("sequence asc").
		Find(&transcripts).Error; err != nil {
		return nil, err
	}
	lesson.Transcripts = transcripts
	return &lesson, nil
}

func (r *videoExerciseRepository) Create(ctx context.Context, lesson *models.VideoExercise) error {
	return r.db.WithContext(ctx).Create(lesson).Error
}

func (r *videoExerciseRepository) Update(ctx context.Context, lesson *models.VideoExercise) error {
	result := r.db.WithContext(ctx).Save(lesson)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *videoExerciseRepository) Delete(ctx context.Context, id string) error {
	_, err := parseID(id, "id")
	if err != nil {
		return err
	}
	result := r.db.WithContext(ctx).Delete(&models.VideoExercise{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *videoExerciseRepository) Count(ctx context.Context, onlyPublished bool) (int64, error) {
	var count int64
	query := visiblePublishedVideos(r.db.WithContext(ctx).Model(&models.VideoExercise{}), onlyPublished)
	err := query.Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
