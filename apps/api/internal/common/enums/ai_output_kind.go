package enums

// AIOutputKind names one cached paid AI output. The (subject, kind) pair is
// the caching boundary: generate once, read free forever.
type AIOutputKind string

const (
	AIOutputKindExamReview    AIOutputKind = "exam_review"
	AIOutputKindWritingReview AIOutputKind = "writing_review" // future Phase 2b; allowed value only
	AIOutputKindStudyPlan     AIOutputKind = "study_plan"
	AIOutputKindTurnAnalysis  AIOutputKind = "turn_analysis"
)
