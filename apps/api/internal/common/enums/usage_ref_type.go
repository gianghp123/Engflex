package enums

// UsageRefType names the entity a reservation is held against — the
// idempotency base key is derived from it.
type UsageRefType string

const (
	UsageRefTypeConversation    UsageRefType = "conversation"
	UsageRefTypeExamAttempt     UsageRefType = "exam_attempt"
	UsageRefTypeWritingActivity UsageRefType = "writing_activity"
	UsageRefTypeStudyPlan       UsageRefType = "study_plan"
)
