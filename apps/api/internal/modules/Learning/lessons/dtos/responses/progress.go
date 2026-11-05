package responses

import (
	"engflex-api/internal/common/enums"
	"time"
)

// CheckedQuestion is one already-graded question, rehydrated for resume: the
// learner's pick plus the reveal (correct key + explanation) the check
// unlocked. Only ever served for the attempt owner's open run.
type CheckedQuestion struct {
	Index       int    `json:"index"`
	Key         string `json:"key"`
	Correct     bool   `json:"correct"`
	CorrectKey  string `json:"correctKey"`
	Explanation string `json:"explanation"`
}

// AttemptProgress is the resume payload: run state, per-activity graded
// questions keyed by activity id (checkable activities only), and the ids of
// every part with a recorded result — writing and speaking included.
type AttemptProgress struct {
	Attempt AttemptSummary `json:"attempt"`
	// Checks carries graded questions for checkable activities only.
	Checks map[string][]CheckedQuestion `json:"checks"`
	// CompletedParts contains only activities with every required outcome.
	CompletedParts []string `json:"completedParts"`
}

// LessonProgressSummary is one lesson's hub progress: the representative
// attempt's status and the ids of its attached parts (any activity type).
type LessonProgressSummary struct {
	LessonID             string              `json:"lessonId"`
	Status               enums.AttemptStatus `json:"status"`
	CompletedActivityIDs []string            `json:"completedActivityIds"`
}

type StartedAttempt struct {
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
