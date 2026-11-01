package models

// ExamAttemptStats is a read-only aggregate, not a persisted table.
type ExamAttemptStats struct {
	CompletedTests int64
	HighestScore   *float64
	AverageScore   *float64
	LatestScore    *float64
}
