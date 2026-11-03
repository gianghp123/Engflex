package responses

import "engflex-api/internal/common/enums"

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
// learner's meaning, contextual answers the conversation better. Null means
// not produced.
type SpanAlternatives struct {
	Language   *Alternative `json:"language"`
	Contextual *Alternative `json:"contextual"`
}

// TurnFeedback is the conversation-aware English feedback payload:
// language spans, conversational relevance, and alternatives.
type TurnFeedback struct {
	Corrected    string           `json:"corrected"`
	Spans        []AnalysisSpan   `json:"spans"`
	Relevance    Relevance        `json:"relevance"`
	Alternatives SpanAlternatives `json:"alternatives"`
	Tip          string           `json:"tip"`
}
