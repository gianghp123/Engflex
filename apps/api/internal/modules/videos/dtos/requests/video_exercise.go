package requests

import (
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
)

// ListVideoExercises pages the video catalog. Pagination and the title /
// description search come from the embedded common.ListParams; categoryId
// and cefrLevel narrow the rows and mirror the response field names.
type ListVideoExercises struct {
	common.ListParams `tstype:",extends"`
	CategoryID        *string     `form:"categoryId" json:"categoryId" binding:"omitempty,uuid"`
	CEFRLevel         *enums.CEFR `form:"cefrLevel" json:"cefrLevel" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
}

type CreateVideoExercise struct {
	CategoryID   *string     `json:"categoryId" binding:"omitempty,uuid"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	VideoURL     string      `json:"videoUrl"`
	YoutubeURL   string      `json:"youtubeUrl"`
	ThumbnailURL string      `json:"thumbnailUrl"`
	CEFRLevel    *enums.CEFR `json:"cefrLevel" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Duration     float64     `json:"duration"`
}

type UpdateVideoExercise struct {
	CategoryID   *string     `json:"categoryId" binding:"omitempty,uuid"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	VideoURL     string      `json:"videoUrl"`
	YoutubeURL   string      `json:"youtubeUrl"`
	ThumbnailURL string      `json:"thumbnailUrl"`
	CEFRLevel    *enums.CEFR `json:"cefrLevel" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Duration     float64     `json:"duration"`
}
