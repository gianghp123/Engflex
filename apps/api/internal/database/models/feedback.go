package models

import (
	"database/sql/driver"
	"time"

	"engflex-api/internal/common/enums"
)

// Feedback is the global, per-subject coaching record: one row per
// (subject_type, subject_id), whatever the subject is. There is deliberately
// no `type` column -- the subject-to-product mapping is a product invariant,
// so the payload envelope carries the variation instead (spec D17/D18).
// user_id is denormalized because the polymorphic pair has no foreign key;
// it keeps "all my feedback" and "delete my data" answerable.
type Feedback struct {
	ID          string                    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      string                    `gorm:"not null" json:"userId"`
	SubjectType enums.FeedbackSubjectType `gorm:"not null" json:"subjectType"`
	SubjectID   string                    `gorm:"not null" json:"subjectId"`
	Payload     TurnFeedback              `gorm:"type:jsonb;not null" json:"payload"`
	CreatedAt   time.Time                 `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time                 `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (Feedback) TableName() string { return "feedbacks" }

// AnalysisSpan is one problematic span: the smallest meaningful unit,
// located by exact substring plus 1-based occurrence.
type AnalysisSpan struct {
	Text       string           `json:"text"`
	Occurrence int              `json:"occurrence"`
	Status     enums.SpanStatus `json:"status"`
	ErrorType  enums.ErrorType  `json:"errorType"`
	Correction string           `json:"correction"`
	Reason     string           `json:"reason"`
}

// Relevance judges whether a learner turn answers its conversation context.
type Relevance struct {
	Status enums.RelevanceStatus `json:"status"`
	Reason *string               `json:"reason"`
}

// Alternative is one better phrasing or answer with its reason.
type Alternative struct {
	Text   string `json:"text"`
	Reason string `json:"reason"`
}

// SpanAlternatives holds the singular alternatives: language preserves the
// learner's meaning, contextual answers the conversation better. Nil means
// not produced.
type SpanAlternatives struct {
	Language   *Alternative `json:"language"`
	Contextual *Alternative `json:"contextual"`
}

// TurnFeedback is the typed view over feedbacks.payload (jsonb) for
// conversation turns: the corrected text, the language spans, the relevance
// verdict, the alternatives, and the tip. It mirrors responses.TurnFeedback
// field for field, so utils.Map copies the whole document in one call.
//
// Its json names are camelCase, unlike models.ScenarioDetail's snake_case:
// the scenario blob was seeded with snake_case keys, while this one was
// always written from the wire DTO, whose keys are camelCase.
type TurnFeedback struct {
	Corrected    string           `json:"corrected"`
	Spans        []AnalysisSpan   `json:"spans"`
	Relevance    Relevance        `json:"relevance"`
	Alternatives SpanAlternatives `json:"alternatives"`
	Tip          string           `json:"tip"`
}

// GormDataType pins the column type for migrations handled outside GORM.
func (TurnFeedback) GormDataType() string { return "jsonb" }

// Value marshals the document for writes.
func (f TurnFeedback) Value() (driver.Value, error) { return jsonbValue(f) }

// Scan decodes the column on reads; a broken blob leaves the zero value.
func (f *TurnFeedback) Scan(value any) error {
	jsonbScan(value, f)
	return nil
}
