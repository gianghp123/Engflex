package voice

// IceServer is one STUN/TURN entry handed to the browser.
type IceServer struct {
	URLs       []string `json:"urls"`
	Username   *string  `json:"username,omitempty"`
	Credential *string  `json:"credential,omitempty"`
}

// IceConfig is the engine-supplied ICE configuration for a session.
type IceConfig struct {
	IceServers []IceServer `json:"iceServers"`
}

// TranscriptCommand is the web-facing request for the correction modal.
//
// The three actions share the reviewed turn on the engine, so they are one
// endpoint. The client never names a turn: its row ordinal and the engine's
// collector position are separate bookkeeping, so the engine resolves the target
// and hands back its own copy of the text.
type TranscriptCommand struct {
	// Action is one of review, send, dismiss.
	Action string `json:"action"`
	// Text is the learner's correction, used only by send.
	Text string `json:"text,omitempty"`
}

// TranscriptResult is the engine's reply: where the review is now, and the
// transcript text when the action produced any.
type TranscriptResult struct {
	State string `json:"state"`
	Text  string `json:"text,omitempty"`
}

// TranscribeResult is the engine's reply to a retake upload: the sentence it heard.
type TranscribeResult struct {
	Text string `json:"text"`
}

// PhoneDetail is one mispronounced phone: expected vs heard IPA with confidence.
type PhoneDetail struct {
	Expected   string  `json:"expected"`
	Heard      string  `json:"heard"`
	Confidence float64 `json:"confidence"`
}

// ReferencePhone is one reference word with its expected IPA, in sentence
// order — the panel renders it under every word chip, flagged or not.
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

// PronounceResult is the container's assessment of one exercise attempt: the
// score plus per-word errors. It is constructed by the pronounce engine
// client, not parsed from the voice service.
//
// The container speaks snake_case, so the metric and curve keys match it
// exactly — encoding/json drops silently on a tag mismatch, which is how the
// error rates used to parse as zero.
type PronounceResult struct {
	Score            float64             `json:"score"`
	Transcription    string              `json:"transcription"`
	PhonemeErrorRate float64             `json:"phoneme_error_rate"`
	WordErrorRate    float64             `json:"word_error_rate"`
	AcousticDistance float64             `json:"acoustic_distance"`
	Errors           []MispronouncedWord `json:"errors"`
	Prosody          Prosody             `json:"prosody"`
	ModelCurve       []float64           `json:"model_curve"`
	LearnerCurve     []float64           `json:"learner_curve"`
	ReferencePhones  []ReferencePhone    `json:"reference_phones"`
}
