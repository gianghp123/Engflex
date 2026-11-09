package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

type Category struct {
	TestsCount    int64              `json:"testsCount"`
	AttemptsCount int64              `json:"attemptsCount"`
	ID            string             `json:"id"`
	Slug          string             `json:"slug"`
	Title         string             `json:"title"`
	Type          enums.ExamType     `json:"type"`
	SubType       *enums.ExamSubType `json:"subType,omitempty"`
}

// CategoryReference is the category relation on a test or attempt. Live catalog
// counters are returned by the categories list, not by these identity joins.
type CategoryReference struct {
	ID      string             `json:"id"`
	Slug    string             `json:"slug"`
	Title   string             `json:"title"`
	Type    enums.ExamType     `json:"type"`
	SubType *enums.ExamSubType `json:"subType,omitempty"`
}

type ExamStatistics struct {
	CompletedTests int64    `json:"completedTests"`
	HighestScore   *float64 `json:"highestScore"`
	AverageScore   *float64 `json:"averageScore"`
	LatestScore    *float64 `json:"latestScore"`
}

type ExamAttemptSummary struct {
	ParsedResult      *ResultSummary      `json:"result,omitempty"`
	ParsedIeltsResult *ResultSummary      `json:"ieltsResult,omitempty"`
	ID                string              `json:"id"`
	Status            enums.AttemptStatus `json:"status"`
	Mode              *enums.ExamMode     `json:"mode"`
	SelectedParts     []int64             `json:"selectedParts"`
	Score             *float64            `json:"score"`
	DurationSec       *int                `json:"durationSec"`
	StartedAt         *time.Time          `json:"startedAt"`
	SubmittedAt       *time.Time          `json:"submittedAt"`
	DeadlineAt        *time.Time          `json:"deadlineAt"`
	CreatedAt         time.Time           `json:"createdAt"`
	ExamTest          *AttemptTestSummary `json:"examTest"`
}
type ResultSummary struct {
	CorrectCount  *int     `json:"correctCount"`
	AccuracyPct   *float64 `json:"accuracyPct"`
	QuestionCount int      `json:"questionCount"`
}
type AttemptTestSummary struct {
	ID         string             `json:"id"`
	CategoryID string             `json:"categoryId"`
	Slug       string             `json:"slug"`
	Title      string             `json:"title"`
	Category   *CategoryReference `json:"category"`
}

type ExamAttempt struct {
	ID                string              `json:"id"`
	Status            enums.AttemptStatus `json:"status"`
	Mode              *enums.ExamMode     `json:"mode"`
	SelectedParts     []int64             `json:"selectedParts"`
	Score             *float64            `json:"score"`
	DurationSec       *int                `json:"durationSec"`
	StartedAt         *time.Time          `json:"startedAt"`
	SubmittedAt       *time.Time          `json:"submittedAt"`
	DeadlineAt        *time.Time          `json:"deadlineAt"`
	CreatedAt         time.Time           `json:"createdAt"`
	ExamTest          *TakingTest         `json:"examTest"`
	ParsedResult      *Result             `json:"result,omitempty"`
	ParsedIeltsResult *Result             `json:"ieltsResult,omitempty"`
}
type TakingTest struct {
	ID                 string                `json:"id"`
	CategoryID         string                `json:"categoryId"`
	Slug               string                `json:"slug"`
	Title              string                `json:"title"`
	Skill              *enums.ExamSkill      `json:"skill,omitempty"`
	SubType            *enums.ExamSubType    `json:"subType,omitempty"`
	Category           *CategoryReference    `json:"category"`
	ParsedContent      *TakingContent        `json:"content,omitempty"`
	ParsedIeltsContent *IeltsTakingContent   `json:"ieltsContent,omitempty"`
}

type ReviewAttempt struct {
	ID                string              `json:"id"`
	Status            enums.AttemptStatus `json:"status"`
	Mode              *enums.ExamMode     `json:"mode"`
	SelectedParts     []int64             `json:"selectedParts"`
	Score             *float64            `json:"score"`
	DurationSec       *int                `json:"durationSec"`
	StartedAt         *time.Time          `json:"startedAt"`
	SubmittedAt       *time.Time          `json:"submittedAt"`
	DeadlineAt        *time.Time          `json:"deadlineAt"`
	CreatedAt         time.Time           `json:"createdAt"`
	ExamTest          *ReviewTest         `json:"examTest"`
	ParsedResult      *Result             `json:"result,omitempty"`
	ParsedIeltsResult *Result             `json:"ieltsResult,omitempty"`
}
type ReviewTest struct {
	ID                 string                `json:"id"`
	CategoryID         string                `json:"categoryId"`
	Slug               string                `json:"slug"`
	Title              string                `json:"title"`
	Skill              *enums.ExamSkill      `json:"skill,omitempty"`
	SubType            *enums.ExamSubType    `json:"subType,omitempty"`
	Category           *CategoryReference    `json:"category"`
	ParsedContent      *ReviewContent        `json:"content,omitempty"`
	ParsedIeltsContent *IeltsReviewContent   `json:"ieltsContent,omitempty"`
}
