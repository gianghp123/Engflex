package responses

import "time"

type VideoCategoryResponse struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type VideoCategoryPreview struct {
	ID         string                  `json:"id"`
	Slug       string                  `json:"slug"`
	Name       string                  `json:"name"`
	VideoCount int64                   `json:"videoCount"`
	Videos     []VideoExerciseResponse `json:"videos"`
}
