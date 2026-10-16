package responses

import "engflex-api/internal/common/enums"

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
	// CompletedParts lists the activity ids present in the run's result map:
	// the server's own finalize rule (any recorded entry covers a part).
	CompletedParts []string `json:"completedParts"`
}

// LessonProgressSummary is one lesson's hub progress: the representative
// attempt's status and the ids of its attached parts (any activity type).
type LessonProgressSummary struct {
	LessonID             string              `json:"lessonId"`
	Status               enums.AttemptStatus `json:"status"`
	CompletedActivityIDs []string            `json:"completedActivityIds"`
}
