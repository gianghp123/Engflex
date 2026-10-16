package requests

import (
	"encoding/json"

	"engflex-api/internal/common/enums"
)

// LessonDetailsInput mirrors models.LessonDetails field for field so the
// service's hand-written mapper (not a blanket utils.Map — see
// toLessonDetails) can line the two up directly.
type LessonDetailsInput struct {
	EstimatedDurationMin *int                `json:"estimatedDurationMin"`
	AcousticTargetPct    *int                `json:"acousticTargetPct"`
	CoverImageURL        string              `json:"coverImageUrl"`
	Skill                *enums.ActivityType `json:"skill" binding:"omitempty,oneof=reading listening writing speaking"`
}

// CreateLesson adds an admin-curated curriculum unit, starting as draft
// until explicitly published.
type CreateLesson struct {
	Slug        string             `json:"slug" binding:"required"`
	Title       string             `json:"title" binding:"required"`
	SectionID   string             `json:"sectionId" binding:"required,uuid"`
	CEFRLevel   enums.CEFR         `json:"cefrLevel" binding:"required,oneof=A1 A2 B1 B2 C1 C2"`
	Description string             `json:"description"`
	Details     LessonDetailsInput `json:"details"`
}

// UpdateLesson edits a unit. Status only takes effect for the explicit
// publish/unpublish action (service-enforced, matching the other modules).
type UpdateLesson struct {
	Slug        *string              `json:"slug"`
	Title       *string              `json:"title"`
	SectionID   *string              `json:"sectionId" binding:"omitempty,uuid"`
	CEFRLevel   *enums.CEFR          `json:"cefrLevel" binding:"omitempty,oneof=A1 A2 B1 B2 C1 C2"`
	Description *string              `json:"description"`
	Details     *LessonDetailsInput  `json:"details"`
	Status      *enums.PublishStatus `json:"status" binding:"omitempty,oneof=draft published"`
}

// CreateLessonSection / UpdateLessonSection manage the path's ordered
// sections. Full replace, like video/vocabulary categories — sections have
// no draft concept of their own.
type CreateLessonSection struct {
	Slug        string     `json:"slug" binding:"required"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	CEFRBand    enums.CEFR `json:"cefrBand" binding:"required,oneof=A1 A2 B1 B2 C1 C2"`
	Position    int        `json:"position"`
}

type UpdateLessonSection struct {
	Slug        string     `json:"slug" binding:"required"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	CEFRBand    enums.CEFR `json:"cefrBand" binding:"required,oneof=A1 A2 B1 B2 C1 C2"`
	Position    int        `json:"position"`
}

// CreateLessonActivity / UpdateLessonActivity manage one ordered part of a
// lesson. Config arrives as a raw JSON document whose shape depends on
// Type (see activityconfig.Reading/Listening/Writing/SpeakingConfig) — the
// service parses and validates it against that shape before saving, so a
// malformed config never reaches the column.
type CreateLessonActivity struct {
	LessonID    string             `json:"lessonId" binding:"required,uuid"`
	PartNumber  int                `json:"partNumber" binding:"required,min=1"`
	Type        enums.ActivityType `json:"type" binding:"required,oneof=reading listening writing speaking"`
	Title       string             `json:"title" binding:"required"`
	Description string             `json:"description"`
	SkillFocus  string             `json:"skillFocus"`
	DurationMin int                `json:"durationMin"`
	Config      json.RawMessage    `json:"config" binding:"required"`
}

type UpdateLessonActivity struct {
	PartNumber  *int                `json:"partNumber" binding:"omitempty,min=1"`
	Type        *enums.ActivityType `json:"type" binding:"omitempty,oneof=reading listening writing speaking"`
	Title       *string             `json:"title"`
	Description *string             `json:"description"`
	SkillFocus  *string             `json:"skillFocus"`
	DurationMin *int                `json:"durationMin"`
	Config      json.RawMessage     `json:"config"`
}
