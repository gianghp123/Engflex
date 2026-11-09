// Package pronunciation scores one spoken attempt against a reference text.
// The engine owns the acoustics; this package owns the Go-side policy (size
// cap, engine-error mapping, wire-shape mapping) so every feature that needs
// scoring — lesson speaking today, vocabulary/shadowing/retakes later — shares
// one implementation. Consumers resolve their own reference text.
package pronunciation

import (
	"context"
	"errors"
	"net/http"

	"engflex-api/internal/common"
	"engflex-api/internal/logger"
	"engflex-api/internal/utils"
	"engflex-api/internal/voice"
)

// MaxAudioBytes mirrors the engine's MAX_AUDIO_BYTES on /pronounce.
const MaxAudioBytes = 2 * 1024 * 1024

// PhoneDetail is one mispronounced phone: expected vs heard IPA with confidence.
type PhoneDetail struct {
	Expected   string
	Heard      string
	Confidence float64
}

// ReferencePhone mirrors one reference word's expected IPA from the engine.
type ReferencePhone struct {
	Word   string
	Phones string
}

// MispronouncedWord mirrors the engine's per-word finding. Go field names match
// the wire DTOs so utils.Map copies whole. Tip is Go-side coaching prose, set
// after the engine map; the engine never sends it.
type MispronouncedWord struct {
	Word       string
	Expected   string
	Heard      string
	Confidence float64
	Phones     []PhoneDetail
	Tip        string
}

// Prosody mirrors the engine's pitch and energy contours.
type Prosody struct {
	F0     []float64
	Energy []float64
}

// Result mirrors the engine's assessment of one attempt.
type Result struct {
	Score            float64
	Transcription    string
	PhonemeErrorRate float64
	WordErrorRate    float64
	AcousticDistance float64
	Errors           []MispronouncedWord
	Prosody          Prosody
	ModelCurve       []float64
	LearnerCurve     []float64
	ReferencePhones  []ReferencePhone
}

// Scorer scores one attempt. An interface so module services depend on it
// without the concrete client.
type Scorer interface {
	Score(ctx context.Context, expectedText string, audio []byte, mime string) (*Result, error)
}

type scorer struct{ engine voice.PronounceEngine }

// NewScorer builds the shared scorer over the pronounce engine client.
func NewScorer(engine voice.PronounceEngine) Scorer { return &scorer{engine: engine} }

// Score enforces the cap, calls the engine with lang="en" and maps the result.
func (s *scorer) Score(ctx context.Context, expectedText string, audio []byte, mime string) (*Result, error) {
	if len(audio) > MaxAudioBytes {
		return nil, common.New(http.StatusRequestEntityTooLarge, "audio too large")
	}
	res, err := s.engine.Pronounce(ctx, expectedText, "en", audio, mime)
	if err != nil || res == nil {
		return nil, mapEngineError(ctx, err)
	}
	var out Result
	if utils.Map(&out, res) != nil {
		return nil, common.Internal()
	}
	for i := range out.Errors {
		out.Errors[i].Tip = CoachingTip(out.Errors[i].Expected, out.Errors[i].Heard)
	}
	return &out, nil
}

// mapEngineError maps the refusals /pronounce can produce; anything else is
// unavailable. Mirrors conversations' transcriptError vocabulary.
func mapEngineError(ctx context.Context, err error) error {
	var status *voice.TranscriptStatusError
	if errors.As(err, &status) {
		switch status.Status {
		case http.StatusBadRequest, http.StatusUnprocessableEntity:
			logger.Report(ctx, "pronounce refused", common.UnprocessableEntity("unscorable audio"))
			return common.UnprocessableEntity("unscorable audio")
		case http.StatusRequestEntityTooLarge:
			logger.Report(ctx, "pronounce refused", common.New(http.StatusRequestEntityTooLarge, "audio too large"))
			return common.New(http.StatusRequestEntityTooLarge, "audio too large")
		}
	}
	logger.Report(ctx, "pronounce engine call failed", common.ServiceUnavailable("pronunciation unavailable"))
	return common.ServiceUnavailable("pronunciation unavailable")
}
