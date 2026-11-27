package responses

import "time"

// PhoneDetail is one mispronounced phone: expected vs heard IPA with confidence.
type PhoneDetail struct {
	Expected   string  `json:"expected"`
	Heard      string  `json:"heard"`
	Confidence float64 `json:"confidence"`
}

// ReferencePhone is one reference word's expected IPA, in sentence order.
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
	Tip        string        `json:"tip"`
}

// Prosody carries the learner's pitch and energy contours, downsampled by
// the engine to at most 120 points each.
type Prosody struct {
	F0     []float64 `json:"f0"`
	Energy []float64 `json:"energy"`
}

// VideoPronunciationResult is the engine's assessment of one shadowing take
// against its transcript line. Named distinctly from lessons'
// PronunciationResult (otherwise structurally identical) because swag
// resolves @Success comment types by package name alone — two "responses"
// packages both exporting the same type name is ambiguous to it even though
// Go itself scopes them fine.
type VideoPronunciationResult struct {
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
	Feedback         string              `json:"feedback"`
}

// ShadowingLineResult is the persisted summary of one scored take, stored
// under VideoProgressPayload.Transcripts[transcriptId] alongside dictation's
// DictationLineResult — only the summary survives (score/feedback/what the
// engine heard), not the full chip/curve detail the live panel renders once
// and never needs again.
type ShadowingLineResult struct {
	Score         float64   `json:"score"`
	Feedback      string    `json:"feedback"`
	Transcription string    `json:"transcription"`
	AttemptedAt   time.Time `json:"attemptedAt"`
}
