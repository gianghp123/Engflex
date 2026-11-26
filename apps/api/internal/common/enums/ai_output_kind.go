package enums

// AIOutputKind names one cached paid AI output. The (subject, kind) pair is
// the caching boundary: generate once, read free forever.
type AIOutputKind string

const (
	AIOutputKindExamReview AIOutputKind = "exam_review"
	// WritingReview is a future Phase 2b output; allowed value only.
	AIOutputKindWritingReview AIOutputKind = "writing_review"
	AIOutputKindStudyPlan     AIOutputKind = "study_plan"
	AIOutputKindTurnAnalysis  AIOutputKind = "turn_analysis"
)
