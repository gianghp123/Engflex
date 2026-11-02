package repositories

import (
	"context"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// CreateOpen lets the partial unique indexes arbitrate concurrent starts.
// The constant second query returns the winning row when another start won.
func (r *attemptRepository) CreateOpen(ctx context.Context, m *models.Attempt) error {
	now := time.Now().UTC()
	m.StartedAt = &now
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(m)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		return nil
	}
	q := r.db.WithContext(ctx).Where("user_id = ? AND type = ? AND status = ?", m.UserID, m.Type, enums.AttemptStatusInProgress)
	if m.LessonID != nil {
		q = q.Where("lesson_id = ?", *m.LessonID)
	} else if m.VideoExerciseID != nil {
		q = q.Where("video_exercise_id = ?", *m.VideoExerciseID)
	} else {
		return common.BadRequest("attempt target required")
	}
	m.ID = ""
	return q.First(m).Error
}

func (r *attemptRepository) GetForUpdate(ctx context.Context, id string) (*models.Attempt, error) {
	if _, err := parseID(id, "attemptId"); err != nil {
		return nil, err
	}
	var m models.Attempt
	if err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *attemptRepository) SaveVideo(ctx context.Context, m *models.Attempt, revision int) error {
	result := r.db.WithContext(ctx).Model(&models.Attempt{}).
		Where("id = ? AND user_id = ? AND video_exercise_id = ? AND revision = ? AND status = ?", m.ID, m.UserID, m.VideoExerciseID, revision, enums.AttemptStatusInProgress).
		Updates(map[string]any{"result": m.Result, "status": m.Status, "score": m.Score, "duration_sec": m.DurationSec, "submitted_at": m.SubmittedAt, "revision": gorm.Expr("revision + 1")})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return common.Conflict("attempt changed; reload before saving")
	}
	m.Revision = revision + 1
	return nil
}
