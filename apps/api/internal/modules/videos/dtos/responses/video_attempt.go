package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

type VideoProgressPayload struct {
	DictationVersion       int            `json:"dictationVersion,omitempty"`
	ActiveSequence         int            `json:"activeSequence"`
	ActiveTranscriptID     *string        `json:"activeTranscriptId,omitempty"`
	Level                  *string        `json:"level,omitempty"`
	CompletedTranscriptIDs []string       `json:"completedTranscriptIds"`
	Transcripts            map[string]any `json:"transcripts"`
}

type DictationLineResult struct {
	Draft     string   `json:"draft"`
	Answers   []string `json:"answers,omitempty"`
	Revealed  []int    `json:"revealed"`
	Completed bool     `json:"completed"`
	Earned    *int     `json:"earned,omitempty"`
	Total     *int     `json:"total,omitempty"`
}

type VideoAttemptResponse struct {
	ID              string                `json:"id"`
	UserID          string                `json:"userId"`
	VideoExerciseID *string               `json:"videoExerciseId,omitempty"`
	Type            enums.AttemptType     `json:"type"`
	Status          enums.AttemptStatus   `json:"status"`
	Score           *float64              `json:"score,omitempty"`
	Result          *VideoProgressPayload `json:"result,omitempty"`
	DurationSec     *int                  `json:"durationSec,omitempty"`
	CreatedAt       time.Time             `json:"createdAt"`
	UpdatedAt       time.Time             `json:"updatedAt"`
}
