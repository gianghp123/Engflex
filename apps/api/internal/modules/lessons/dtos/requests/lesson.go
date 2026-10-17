package requests

import (
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/llm"
)

// ListLessons filters the unit hub (level, skill tab, completion status).
// Pagination/search come from the embedded common.ListParams
// (tstype:",extends" -> `interface ListLessons extends ListParams`); json tags
// mirror the query keys so the generated TS field names match the wire.
type ListLessons struct {
	common.ListParams `tstype:",extends"`
	Level             enums.CEFR         `form:"level" json:"level" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Skill             enums.ActivityType `form:"skill" json:"skill" binding:"omitempty,oneof=reading listening writing speaking"`
	Status            enums.LessonStatus `form:"status" json:"status" binding:"omitempty,oneof=unstarted in_progress completed"`
}

// ListActivities filters GET /lesson-activities (?lessonId= required).
type ListActivities struct {
	LessonID string `form:"lessonId" json:"lessonId" binding:"required,uuid"`
}

// SaveBookmark is the POST /lesson-bookmarks body.
type SaveBookmark struct {
	LessonID string `json:"lessonId" binding:"required,uuid"`
}

// CheckAnswer grades one reading/listening question. min=0, not required:
// `required` rejects the zero value, and question 0 is valid.
type CheckAnswer struct {
	AttemptID     string `json:"attemptId" binding:"required,uuid"`
	QuestionIndex int    `json:"questionIndex" binding:"min=0"`
	Key           string `json:"key" binding:"required"`
}

// ScoreWriting grades one writing part into the open lesson attempt.
type ScoreWriting struct {
	llm.LanguageParam `tstype:",extends"`
	AttemptID         string `json:"attemptId" binding:"required,uuid"`
	Text              string `json:"text" binding:"required"`
}
