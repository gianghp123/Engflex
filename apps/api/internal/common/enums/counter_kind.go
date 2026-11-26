package enums

// CounterKind names a metered resource. Two exist at launch: ai_credit is
// the single user-facing unit; voice_session is the cost sub-cap.
type CounterKind string

const (
	CounterKindAICredit     CounterKind = "ai_credit"
	CounterKindVoiceSession CounterKind = "voice_session"
)
