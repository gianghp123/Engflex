package models

// ExamStatistics summarizes a learner's completed full TOEIC exams. Practice
// attempts are excluded because their question sets are not comparable.
// Nullable scores remain absent until a real score has been recorded.
type ExamStatistics struct {
	CompletedTests int64
	HighestScore   *float64
	AverageScore   *float64
	LatestScore    *float64
}
