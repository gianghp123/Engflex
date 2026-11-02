package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/exams/toeic"
)

// Attempt is the single source of truth for learner runs across lesson,
// video exercise, and full exam attempts. Per-part/answers results accumulate in
// result (jsonb); target content is identified by lesson_id, video_exercise_id,
// or exam_test_id.
type Attempt struct {
	Revision        int                 `gorm:"not null;default:1" json:"revision"`
	ID              string              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID          string              `gorm:"not null" json:"userId"`
	LessonID        *string             `gorm:"type:uuid" json:"lessonId"`
	VideoExerciseID *string             `gorm:"type:uuid" json:"videoExerciseId"`
	ExamTestID      *string             `gorm:"type:uuid" json:"examTestId"`
	Type            enums.AttemptType   `gorm:"not null" json:"type"`
	Status          enums.AttemptStatus `gorm:"not null;default:in_progress" json:"status"`
	Mode            *enums.ExamMode     `gorm:"type:text" json:"mode"`
	SelectedParts   pq.Int64Array       `gorm:"type:smallint[]" json:"selectedParts,omitempty"`
	Score           *float64            `gorm:"type:numeric(5,2)" json:"score"`
	Result          datatypes.JSON      `gorm:"type:jsonb;not null;default:'{}'" json:"result"`
	ParsedResult    *toeic.Result       `gorm:"-" json:"-"`
	DurationSec     *int                `json:"durationSec"`
	StartedAt       *time.Time          `json:"startedAt"`
	SubmittedAt     *time.Time          `json:"submittedAt"`
	DeadlineAt      *time.Time          `json:"deadlineAt"`
	CreatedAt       time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`

	Lesson        *Lesson        `gorm:"->;foreignKey:LessonID;references:ID" json:"lesson,omitempty"`
	VideoExercise *VideoExercise `gorm:"->;foreignKey:VideoExerciseID;references:ID" json:"videoExercise,omitempty"`
	ExamTest      *ExamTest      `gorm:"->;foreignKey:ExamTestID;references:ID" json:"examTest,omitempty"`
}

func (Attempt) TableName() string { return "attempts" }
