package voice

import (
	"encoding/json"
	"math"
)

// maxPronounceCurvePoints caps every contour the panel draws. The container
// returns full-resolution vectors; the wire carries at most this many points.
const maxPronounceCurvePoints = 120

// containerPhone is one misheard phone inside a flagged word. Keys follow the
// container's raw shape (expected/heard), not the wire DTOs.
type containerPhone struct {
	Expected   string  `json:"expected"`
	Heard      string  `json:"heard"`
	Confidence float64 `json:"confidence"`
}

// containerWordError is one entry of the raw differences.errors list.
// Actual is any: the phone path carries a string, the word path carries null
// (or omits the key) — both map to "".
type containerWordError struct {
	Word       string           `json:"word"`
	Expected   string           `json:"expected"`
	Actual     any              `json:"actual"`
	Confidence float64          `json:"confidence"`
	Phones     []containerPhone `json:"phones"`
}

type containerDifferences struct {
	PhonemeErrorRate  float64              `json:"phoneme_error_rate"`
	WordErrorRate     float64              `json:"word_error_rate"`
	Errors            []containerWordError `json:"errors"`
	ExpectedVector    []float64            `json:"expected_vector"`
	TranscribedVector []float64            `json:"transcribed_vector"`
}

type containerProsody struct {
	F0     []float64 `json:"f0"`
	Energy []float64 `json:"energy"`
}

// containerAssessment is the raw POST /pronunciation reply: no envelope.
type containerAssessment struct {
	AcousticDistance float64              `json:"acoustic_distance"`
	Transcribe       string               `json:"transcribe"`
	Differences      containerDifferences `json:"differences"`
	Prosody          containerProsody     `json:"prosody"`
}

// containerPhonemes is the raw POST /phonemes reply: the token each word owns.
type containerPhonemes struct {
	Phonemes []string `json:"phonemes"`
	Words    []string `json:"words"`
}

// learningScore grades pronunciation only: the container's blended score gives
// 30% to an acoustic term that measures voice-timbre similarity to a TTS
// reference, not pronunciation, so Go recomputes from the error rates.
// acoustic_distance stays on the wire as a diagnostic, excluded from the grade.
func learningScore(phonemeErrorRate, wordErrorRate float64) float64 {
	term := func(rate float64) float64 {
		return min(max(100*(1-rate), 0), 100)
	}
	return math.Round((0.6*term(phonemeErrorRate)+0.4*term(wordErrorRate))*100) / 100
}

// downsampleCurve keeps short contours whole and strides long ones to at most
// maxPronounceCurvePoints, first frame kept. Never returns nil: the contracts
// say arrays, so JSON must emit [] never null.
func downsampleCurve(values []float64) []float64 {
	out := append([]float64{}, values...)
	if len(out) <= maxPronounceCurvePoints {
		return out
	}
	step := (len(out) + maxPronounceCurvePoints - 1) / maxPronounceCurvePoints
	kept := make([]float64, 0, maxPronounceCurvePoints)
	for i := 0; i < len(out); i += step {
		kept = append(kept, out[i])
	}
	return kept
}

// groupReferencePhones folds /phonemes tokens into one entry per word:
// consecutive runs of the same word merge (same word, same phones is
// display-equivalent), multi-token words concatenate. Never returns nil.
func groupReferencePhones(phonemes, words []string) []ReferencePhone {
	out := []ReferencePhone{}
	for i, word := range words {
		token := ""
		if i < len(phonemes) {
			token = phonemes[i]
		}
		if n := len(out); n > 0 && out[n-1].Word == word {
			out[n-1].Phones += token
			continue
		}
		out = append(out, ReferencePhone{Word: word, Phones: token})
	}
	return out
}

func heardString(v any) string {
	s, _ := v.(string)
	return s
}

// mapContainerAssessment ports the old Python mapper onto the wire DTO. A
// missing differences object yields zero values, never a panic; every list
// stays non-nil so JSON emits [].
func mapContainerAssessment(assessmentRaw, phonemesRaw []byte) (*PronounceResult, error) {
	var assessment containerAssessment
	if err := json.Unmarshal(assessmentRaw, &assessment); err != nil {
		return nil, err
	}
	var phonemes containerPhonemes
	if err := json.Unmarshal(phonemesRaw, &phonemes); err != nil {
		return nil, err
	}

	diffs := assessment.Differences
	errors := []MispronouncedWord{}
	for _, entry := range diffs.Errors {
		phones := []PhoneDetail{}
		for _, p := range entry.Phones {
			phones = append(phones, PhoneDetail{
				Expected:   p.Expected,
				Heard:      p.Heard,
				Confidence: p.Confidence,
			})
		}
		errors = append(errors, MispronouncedWord{
			Word:       entry.Word,
			Expected:   entry.Expected,
			Heard:      heardString(entry.Actual),
			Confidence: entry.Confidence,
			Phones:     phones,
		})
	}

	return &PronounceResult{
		Score:            learningScore(diffs.PhonemeErrorRate, diffs.WordErrorRate),
		Transcription:    assessment.Transcribe,
		PhonemeErrorRate: diffs.PhonemeErrorRate,
		WordErrorRate:    diffs.WordErrorRate,
		AcousticDistance: assessment.AcousticDistance,
		Errors:           errors,
		Prosody: Prosody{
			F0:     downsampleCurve(assessment.Prosody.F0),
			Energy: downsampleCurve(assessment.Prosody.Energy),
		},
		ModelCurve:      downsampleCurve(diffs.ExpectedVector),
		LearnerCurve:    downsampleCurve(diffs.TranscribedVector),
		ReferencePhones: groupReferencePhones(phonemes.Phonemes, phonemes.Words),
	}, nil
}
