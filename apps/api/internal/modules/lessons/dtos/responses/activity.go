package responses

import (
	"encoding/json"
	"time"

	"engflex-api/internal/common/enums"
)

// AdminActivity is the admin-authoring shape: the raw config verbatim (answer
// keys, voice ids, model audio keys included), unlike the learner-facing
// Activity above which strips those through ApplyPayload. Only ever
// returned from admin-gated endpoints.
type AdminActivity struct {
	ID          string             `json:"id"`
	LessonID    string             `json:"lessonId"`
	PartNumber  int                `json:"partNumber"`
	Type        enums.ActivityType `json:"type"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	SkillFocus  string             `json:"skillFocus"`
	DurationMin int                `json:"durationMin"`
	Config      json.RawMessage    `json:"config"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}

// QuestionOption is one multiple-choice option; correctness never crosses.
type QuestionOption struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// Question is one checkable reading/listening item (answers stay server-side).
type Question struct {
	Stem        string           `json:"stem"`
	Instruction string           `json:"instruction"`
	Options     []QuestionOption `json:"options"`
}

type ReadingPayload struct {
	Text      string     `json:"text"`
	Questions []Question `json:"questions"`
}

type ListeningPayload struct {
	AudioURL   string     `json:"audioUrl"`
	Transcript string     `json:"transcript"`
	Questions  []Question `json:"questions"`
}

type WritingPayload struct {
	Task         string   `json:"task"`
	Instructions string   `json:"instructions"`
	Stimulus     string   `json:"stimulus"`
	MinWords     int      `json:"minWords"`
	MaxWords     int      `json:"maxWords"`
	ModelAnswer  string   `json:"modelAnswer"`
	Checklist    []string `json:"checklist"`
}

type SpeakingItem struct {
	Text          string `json:"text"`
	ModelAudioURL string `json:"modelAudioUrl"`
}

type SpeakingPayload struct {
	Items []SpeakingItem `json:"items"`
}

// CheckResult is the grading outcome for one question. The explanation is
// revealed only after the learner commits an answer.
type CheckResult struct {
	Correct     bool   `json:"correct"`
	CorrectKey  string `json:"correctKey"`
	Explanation string `json:"explanation"`
}

// PhoneDetail is one mispronounced phone: expected vs heard IPA with confidence.
type PhoneDetail struct {
	Expected   string  `json:"expected"`
	Heard      string  `json:"heard"`
	Confidence float64 `json:"confidence"`
}

// ReferencePhone is one reference word's expected IPA, in sentence order.
// The panel renders it under every word chip, flagged or not.
type ReferencePhone struct {
	Word   string `json:"word"`
	Phones string `json:"phones"`
}

// MispronouncedWord is one engine-flagged word: expected vs heard IPA.
type MispronouncedWord struct {
	Word       string        `json:"word"`
	Expected   string        `json:"expected"`
	Heard      string        `json:"heard"`
	Confidence float64       `json:"confidence"`
	Phones     []PhoneDetail `json:"phones"`
}

// Prosody carries the learner's pitch and energy contours, downsampled by
// the engine to at most 120 points each.
type Prosody struct {
	F0     []float64 `json:"f0"`
	Energy []float64 `json:"energy"`
}

// PronunciationResult is the engine's assessment of one read-aloud attempt.
type PronunciationResult struct {
	Score            float64             `json:"score"`
	Transcription    string              `json:"transcription"`
	PhonemeErrorRate float64             `json:"phonemeErrorRate"`
	WordErrorRate    float64             `json:"wordErrorRate"`
	AcousticDistance float64             `json:"acousticDistance"`
	Errors           []MispronouncedWord `json:"errors"`
	Prosody          Prosody             `json:"prosody"`
	ModelCurve       []float64           `json:"modelCurve"`
	LearnerCurve     []float64           `json:"learnerCurve"`
	ReferencePhones  []ReferencePhone    `json:"referencePhones"`
}

// Activity is one ordered lesson part. Exactly one payload pointer is set,
// matching Type. DurationMin and SkillFocus are plain columns on the model.
type Activity struct {
	ID          string             `json:"id"`
	PartNumber  int                `json:"partNumber"`
	Type        enums.ActivityType `json:"type"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	DurationMin int                `json:"durationMin"`
	SkillFocus  string             `json:"skillFocus"`
	Reading     *ReadingPayload    `json:"reading"`
	Listening   *ListeningPayload  `json:"listening"`
	Writing     *WritingPayload    `json:"writing"`
	Speaking    *SpeakingPayload   `json:"speaking"`
}

// TaskVerdict is one checklist judgement with its evidence quote.
type TaskVerdict struct {
	Item     string `json:"item"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

// Span is one quoted problem with a minimal fix. The array it sits in is the
// label (grammar vs phrasing).
type Span struct {
	Text       string `json:"text"`
	Occurrence int    `json:"occurrence"`
	Correction string `json:"correction"`
	Reason     string `json:"reason"`
}

// TaskCompletion is section 1; counts live in ScoreSummary, never here.
type TaskCompletion struct {
	Verdicts      []TaskVerdict `json:"verdicts"`
	AnswersPrompt bool          `json:"answersPrompt"`
	AnswersNote   string        `json:"answersNote"`
}

// ScoreSummary is the server-computed header strip.
type ScoreSummary struct {
	PointsCovered int     `json:"pointsCovered"`
	PointsTotal   int     `json:"pointsTotal"`
	Score         float64 `json:"score"`
	WordCount     int     `json:"wordCount"`
	MinWords      int     `json:"minWords"`
	MaxWords      int     `json:"maxWords"`
}

// AttemptSummary lets the UI render run state with zero refetch.
type AttemptSummary struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Score  *float64 `json:"score"`
}

// WritingScoreResponse is the full wire response: feedback + numbers + run
// state in one round trip. No copy of the learner paragraph anywhere.
type WritingScoreResponse struct {
	Summary     ScoreSummary   `json:"summary"`
	Task        TaskCompletion `json:"task"`
	Grammar     []Span         `json:"grammar"`
	Phrasing    []Span         `json:"phrasing"`
	Expressions []string       `json:"expressions"`
	Suggested   string         `json:"suggested"`
	Tip         string         `json:"tip"`
	Attempt     AttemptSummary `json:"attempt"`
}
