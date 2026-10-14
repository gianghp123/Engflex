package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

// Attempt is one recorded practice run (the progress source of truth).
// result stays loosely typed this phase; typed per-practice schemas come
// with the practice endpoints.
type Attempt struct {
	ID          string              `json:"id"`
	UserID      string              `json:"userId"`
	LessonID    *string             `json:"lessonId"`
	Type        enums.AttemptType   `json:"type"`
	Status      enums.AttemptStatus `json:"status"`
	Score       *float64            `json:"score"`
	Result      map[string]any      `json:"result"`
	DurationSec *int                `json:"durationSec"`
	CreatedAt   time.Time           `json:"createdAt"`
}

// Progress is the measurable-only dashboard rollup. Fluency, pronunciation,
// WPM and due-vocabulary are deliberately absent: no scoring source exists,
// and due-vocabulary waits on the vocabulary feature. They return when their
// sources do — never as fabricated numbers.
type Progress struct {
	StreakDays        int `json:"streakDays"`
	VoiceMinutesToday int `json:"voiceMinutesToday"`
	DailyGoalMinutes  int `json:"dailyGoalMinutes"`
	LessonsCompleted  int `json:"lessonsCompleted"`
}
