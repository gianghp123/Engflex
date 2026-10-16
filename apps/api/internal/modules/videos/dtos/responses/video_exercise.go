package responses

import (
	"time"

	"engflex-api/internal/common/enums"
)

type VideoExerciseResponse struct {
	ID           string              `json:"id"`
	CategoryID   *string             `json:"categoryId"`
	Title        string              `json:"title"`
	Description  string              `json:"description"`
	VideoURL     string              `json:"videoUrl"`
	ThumbnailURL string              `json:"thumbnailUrl"`
	CEFRLevel    *enums.CEFR         `json:"cefrLevel"`
	Duration     float64             `json:"duration"`
	Status       enums.PublishStatus `json:"status"`
	CreatedAt    time.Time           `json:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt"`

	Category    *VideoCategoryResponse    `json:"category,omitempty"`
	Transcripts []VideoTranscriptResponse `json:"transcripts,omitempty"`
}
