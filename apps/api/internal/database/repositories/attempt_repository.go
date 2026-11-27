package repositories

import (
	"context"
	"time"

	"engflex-api/internal/common"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
)

type AttemptRepository interface {
	GetExamStatistics(ctx context.Context, userID string) (*models.ExamStatistics, error)
	GetExamAttemptForUpdate(context.Context, string, string) (*models.Attempt, error)
	SaveExamAttempt(context.Context, *models.Attempt, int) error
	Create(ctx context.Context, m *models.Attempt) error
	GetByID(ctx context.Context, id string) (*models.Attempt, error)
	GetOpenAttempt(ctx context.Context, userID, lessonID string) (*models.Attempt, error)
	Update(ctx context.Context, m *models.Attempt) error
	ListCompletedDays(ctx context.Context, userID string, since time.Time) ([]string, error)
	CountCompleted(ctx context.Context, userID string) (int64, error)
	ListLessonAttempts(ctx context.Context, userID string) ([]*models.Attempt, error)
	ListCompletedVideo(ctx context.Context, userID string, videoExerciseIDs []string) ([]*models.Attempt, error)
	GetVideoAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType) (*models.Attempt, error)
	GetVideoAttemptForUpdate(ctx context.Context, userID, videoExerciseID, attemptID string) (*models.Attempt, error)
	ListExamAttempts(ctx context.Context, userID string, examType enums.ExamType, limit, offset int) ([]*models.Attempt, error)
	CountExamAttempts(ctx context.Context, userID string, examType enums.ExamType) (int64, error)
	GetExamAttempt(ctx context.Context, userID, attemptID string) (*models.Attempt, error)
	ListRecentByUser(ctx context.Context, userID string, limit int) ([]*models.Attempt, error)
	WithTx(tx *gorm.DB) AttemptRepository
}

func (r *attemptRepository) GetExamStatistics(ctx context.Context, userID string) (*models.ExamStatistics, error) {
	var stats models.ExamStatistics
	err := r.db.WithContext(ctx).Raw(`SELECT count(*) AS completed_tests,
		max(a.score) AS highest_score, avg(a.score) AS average_score,
		(array_agg(a.score ORDER BY a.submitted_at DESC NULLS LAST, a.created_at DESC, a.id DESC))[1] AS latest_score
		FROM attempts a
		JOIN exam_tests t ON t.id = a.exam_test_id
		JOIN exam_categories c ON c.id = t.category_id
		WHERE a.user_id = ? AND a.type = ? AND a.status = ? AND a.mode = ? AND c.type = ?`,
		userID, enums.AttemptTypeExam, enums.AttemptStatusCompleted, enums.ExamModeExam, enums.ExamTypeTOEIC).Scan(&stats).Error
	return &stats, err
}

func (r *attemptRepository) ListExamAttempts(ctx context.Context, userID string, examType enums.ExamType, limit, offset int) ([]*models.Attempt, error) {
	var attempts []*models.Attempt
	query := r.db.WithContext(ctx).Joins("ExamTest.Category", r.db.Omit("Content")).Where("attempts.user_id = ? AND attempts.type = ?", userID, enums.AttemptTypeExam)
	if examType != "" {
		query = query.Where(`"ExamTest__Category".type = ?`, examType)
	}
	err := query.
		Order("attempts.created_at DESC").Order("attempts.id DESC").Limit(limit).Offset(offset).Find(&attempts).Error
	return attempts, err
}

func (r *attemptRepository) CountExamAttempts(ctx context.Context, userID string, examType enums.ExamType) (int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Attempt{}).Where("attempts.user_id = ? AND attempts.type = ?", userID, enums.AttemptTypeExam)
	if examType != "" {
		query = query.Joins("ExamTest.Category").Where(`"ExamTest__Category".type = ?`, examType)
	}
	err := query.Count(&total).Error
	return total, err
}

func (r *attemptRepository) GetExamAttempt(ctx context.Context, userID, attemptID string) (*models.Attempt, error) {
	id, err := parseID(attemptID, "attemptId")
	if err != nil {
		return nil, err
	}
	var attempt models.Attempt
	err = r.db.WithContext(ctx).Joins("ExamTest.Category").Where("attempts.user_id = ? AND attempts.type = ?", userID, enums.AttemptTypeExam).First(&attempt, "attempts.id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

type attemptRepository struct {
	db *gorm.DB
}

func NewAttemptRepository(db *gorm.DB) AttemptRepository {
	return &attemptRepository{db: db}
}

func (r *attemptRepository) Create(ctx context.Context, m *models.Attempt) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *attemptRepository) GetByID(ctx context.Context, id string) (*models.Attempt, error) {
	var m models.Attempt
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *attemptRepository) GetOpenAttempt(ctx context.Context, userID, lessonID string) (*models.Attempt, error) {
	var m models.Attempt
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND lesson_id = ? AND status = ?", userID, lessonID, enums.AttemptStatusInProgress).
		Order("created_at DESC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *attemptRepository) Update(ctx context.Context, m *models.Attempt) error {
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *attemptRepository) ListCompletedDays(ctx context.Context, userID string, since time.Time) ([]string, error) {
	var days []string
	err := r.db.WithContext(ctx).Model(&models.Attempt{}).
		Select("DISTINCT to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')").
		Where("user_id = ? AND status = ? AND created_at >= ?", userID, enums.AttemptStatusCompleted, since).
		Scan(&days).Error
	return days, err
}

// CountCompleted counts completed lesson runs for the LessonsCompleted metric.
func (r *attemptRepository) CountCompleted(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Attempt{}).
		Where("user_id = ? AND type = ? AND status = ?", userID, enums.AttemptTypeLesson, enums.AttemptStatusCompleted).
		Count(&n).Error
	return n, err
}

func (r *attemptRepository) ListLessonAttempts(ctx context.Context, userID string) ([]*models.Attempt, error) {
	var rows []*models.Attempt
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND lesson_id IS NOT NULL AND status IN ?", userID,
			[]enums.AttemptStatus{enums.AttemptStatusInProgress, enums.AttemptStatusCompleted}).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *attemptRepository) WithTx(tx *gorm.DB) AttemptRepository {
	return &attemptRepository{db: tx}
}

// ListRecentByUser returns the caller's newest attempts for the error
// aggregation endpoints (query-light: capped by the caller).
func (r *attemptRepository) ListRecentByUser(ctx context.Context, userID string, limit int) ([]*models.Attempt, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []*models.Attempt
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *attemptRepository) ListCompletedVideo(ctx context.Context, userID string, videoExerciseIDs []string) ([]*models.Attempt, error) {
	if len(videoExerciseIDs) == 0 {
		return nil, nil
	}
	var attempts []*models.Attempt
	err := r.db.WithContext(ctx).Where("user_id = ? AND video_exercise_id IN ? AND status = ?", userID, videoExerciseIDs, enums.AttemptStatusCompleted).Find(&attempts).Error
	if err != nil {
		return nil, err
	}
	return attempts, err
}

func (r *attemptRepository) GetVideoAttempt(ctx context.Context, userID, videoExerciseID string, attemptType enums.AttemptType) (*models.Attempt, error) {
	var m models.Attempt
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND video_exercise_id = ? AND type = ? AND status = ?",
			userID, videoExerciseID, attemptType, enums.AttemptStatusInProgress).
		Order("created_at DESC").First(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *attemptRepository) GetVideoAttemptForUpdate(ctx context.Context, userID, videoExerciseID, attemptID string) (*models.Attempt, error) {
	var attempt models.Attempt
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND user_id = ? AND video_exercise_id = ? AND type IN ?", attemptID, userID, videoExerciseID,
			[]enums.AttemptType{enums.AttemptTypeDictation, enums.AttemptTypeShadowing}).
		First(&attempt).Error
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *attemptRepository) GetExamAttemptForUpdate(ctx context.Context, userID, attemptID string) (*models.Attempt, error) {
	id, err := parseID(attemptID, "attemptId")
	if err != nil {
		return nil, err
	}
	var m models.Attempt
	err = r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE", Table: clause.Table{Name: clause.CurrentTable}}).Joins("ExamTest.Category").Where("attempts.user_id = ? AND attempts.type = ?", userID, enums.AttemptTypeExam).First(&m, "attempts.id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}
func (r *attemptRepository) SaveExamAttempt(ctx context.Context, m *models.Attempt, revision int) error {
	result := r.db.WithContext(ctx).Model(&models.Attempt{}).Where("id = ? AND user_id = ? AND type = ? AND status = ? AND (result->>'revision')::integer = ?", m.ID, m.UserID, enums.AttemptTypeExam, enums.AttemptStatusInProgress, revision).Updates(map[string]any{"result": m.Result, "status": m.Status, "score": m.Score, "submitted_at": m.SubmittedAt, "duration_sec": m.DurationSec})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return common.Conflict("attempt revision changed")
	}
	return nil
}
