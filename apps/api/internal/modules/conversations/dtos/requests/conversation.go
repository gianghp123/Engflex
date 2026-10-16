package requests

import (
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
)

// StartConversation provisions a session. scenarioId is required for
// roleplay and omitted for free talk.
type StartConversation struct {
	Mode       enums.ConversationMode `json:"mode" binding:"required,oneof=free_talk roleplay"`
	ScenarioID *string                `json:"scenarioId" binding:"omitempty,uuid"`
}

// ListConversations pages the caller's finished conversations, optionally
// scoped to one scenario.
type ListConversations struct {
	common.ListParams `tstype:",extends"`
	ScenarioID        *string `form:"scenarioId" json:"scenarioId" binding:"omitempty,uuid"`
}

// AnalysisSpanInput is one problematic span in an ingested analysis.
type AnalysisSpanInput struct {
	Text       string           `json:"text"`
	Occurrence int              `json:"occurrence"`
	Status     enums.SpanStatus `json:"status"`
	Correction string           `json:"correction"`
	Reason     string           `json:"reason"`
}

// RelevanceInput judges whether a learner turn answers its context.
type RelevanceInput struct {
	Status enums.RelevanceStatus `json:"status"`
	Reason *string               `json:"reason"`
}

// AlternativeInput is one better phrasing or answer with its reason.
type AlternativeInput struct {
	Text   string `json:"text"`
	Reason string `json:"reason"`
}

// SpanAlternativesInput holds the singular alternatives; null means not produced.
type SpanAlternativesInput struct {
	Language   *AlternativeInput `json:"language"`
	Contextual *AlternativeInput `json:"contextual"`
}

// TurnFeedbackInput carries per-turn analysis (spans, relevance,
// alternatives, coaching tip).
type TurnFeedbackInput struct {
	Corrected    string                `json:"corrected"`
	Spans        []AnalysisSpanInput   `json:"spans"`
	Relevance    RelevanceInput        `json:"relevance"`
	Alternatives SpanAlternativesInput `json:"alternatives"`
	Tip          string                `json:"tip"`
}

// CreateTurn appends one transcript line (used by the voice result
// ingestion flow). position keeps transcript order deterministic.
type CreateTurn struct {
	Position int                `json:"position" binding:"required,min=0"`
	Role     enums.TurnRole     `json:"role" binding:"required,oneof=user ai"`
	Text     string             `json:"text" binding:"required,max=10000"`
	Feedback *TurnFeedbackInput `json:"feedback"`
}
