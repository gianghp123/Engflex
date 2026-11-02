package models

type CheckResult struct {
	Correct     bool   `json:"correct"`
	CorrectKey  string `json:"correctKey"`
	Explanation string `json:"explanation"`
}

type ScoreSummary struct {
	PointsCovered int     `json:"pointsCovered"`
	PointsTotal   int     `json:"pointsTotal"`
	Score         float64 `json:"score"`
	WordCount     int     `json:"wordCount"`
	MinWords      int     `json:"minWords"`
	MaxWords      int     `json:"maxWords"`
}

type LessonWritingResult struct {
	WritingFeedback
	Summary ScoreSummary
	Attempt *Attempt
}
type CheckedQuestion struct {
	Index       int
	Key         string
	Correct     bool
	CorrectKey  string
	Explanation string
}
type LessonAttemptProgress struct {
	Attempt        *Attempt
	Checks         map[string][]CheckedQuestion
	CompletedParts []string
}
type LessonProgressSummary struct {
	LessonID             *string
	Status               string
	CompletedActivityIDs []string
}
