package repositories

import (
	"context"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"

	"gorm.io/gorm"
)

type VideoExerciseRepository interface {
	List(ctx context.Context, categoryID *string, cefr *enums.CEFR, search string, limit, offset int) ([]*models.VideoExercise, error)
	GetByID(ctx context.Context, id string) (*models.VideoExercise, error)
	GetDetail(ctx context.Context, id string) (*models.VideoExercise, error)
	Create(ctx context.Context, lesson *models.VideoExercise) error
	Update(ctx context.Context, lesson *models.VideoExercise) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, categoryID *string, cefr *enums.CEFR, search string) (int64, error)
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

func filterVideoExercises(q *gorm.DB, categoryID *string, cefr *enums.CEFR, search string) *gorm.DB {
	if categoryID != nil && *categoryID != "" {
		q = q.Where("video_exercises.category_id = ?", *categoryID)
	}
	if cefr != nil && *cefr != "" {
		q = q.Where("video_exercises.cefr_level = ?", *cefr)
	}
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("(video_exercises.title ILIKE ? OR video_exercises.description ILIKE ?)", like, like)
	}
	return q
}

const exerciseCategorySelect = "video_exercises.*, \"Category\".id AS \"Category__id\", \"Category\".slug AS \"Category__slug\", \"Category\".name AS \"Category__name\", \"Category\".created_at AS \"Category__created_at\", \"Category\".updated_at AS \"Category__updated_at\""

func (r *videoExerciseRepository) List(ctx context.Context, categoryID *string, cefr *enums.CEFR, search string, limit, offset int) ([]*models.VideoExercise, error) {
	lessons := make([]*models.VideoExercise, 0)
	query := filterVideoExercises(r.db.WithContext(ctx).Joins("Category").Select(exerciseCategorySelect), categoryID, cefr, search).
		Order("video_exercises.created_at DESC, video_exercises.id ASC")
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

func (r *videoExerciseRepository) GetDetail(ctx context.Context, id string) (*models.VideoExercise, error) {
	if _, err := parseID(id, "id"); err != nil {
		return nil, err
	}
	var lesson models.VideoExercise
	if err := r.db.WithContext(ctx).Joins("Category").Select(exerciseCategorySelect).First(&lesson, "video_exercises.id = ?", id).Error; err != nil {
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

func (r *videoExerciseRepository) Count(ctx context.Context, categoryID *string, cefr *enums.CEFR, search string) (int64, error) {
	var count int64
	err := filterVideoExercises(r.db.WithContext(ctx).Model(&models.VideoExercise{}), categoryID, cefr, search).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
