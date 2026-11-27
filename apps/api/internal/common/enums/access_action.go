package enums

// AccessAction is a user-facing metered or free capability. The closed set
// is shared by the policy evaluator, the usage ledger, and the admin pricing
// API; only the three paid members have an action_credit_costs row.
type AccessAction string

const (
	AccessActionVoiceStart    AccessAction = "voice_session_start"
	AccessActionExamReview    AccessAction = "exam_ai_review"
	AccessActionStudyPlan     AccessAction = "study_plan"
	AccessActionWritingScore  AccessAction = "writing_score"
	AccessActionPronunciation AccessAction = "pronunciation"
)
