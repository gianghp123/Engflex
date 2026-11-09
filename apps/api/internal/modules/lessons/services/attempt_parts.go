// Package services owns ordered lesson parts and lesson attempts.
package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/lessons/dtos/responses"
	"engflex-api/internal/modules/lessons/schemas"
	"engflex-api/internal/utils"
	"engflex-api/internal/voice/pronunciation"
)

// PartEntry is one scored part attached to a lesson attempt: normalized
// 0-100 score plus the typed detail blob of that part type.
type PartEntry struct {
	Type   enums.AttemptType `json:"type"`
	Score  float64           `json:"score"`
	Detail any               `json:"detail"`
}

// CheckPartDetail accumulates per-question outcomes; repeats overwrite.
type CheckPartDetail struct {
	Outcomes []QuestionOutcome `json:"outcomes"`
}

// QuestionOutcome is one graded question.
type QuestionOutcome struct {
	Index   int    `json:"index"`
	Key     string `json:"key"`
	Correct bool   `json:"correct"`
}

// PronounceItemDetail is one scored read-aloud item: the score plus the
// gated error list and enough provenance to explain it later.
type PronounceItemDetail struct {
	Score         float64                       `json:"score"`
	PER           float64                       `json:"per"`
	WER           float64                       `json:"wer"`
	Errors        []responses.MispronouncedWord `json:"errors"`
	TotalErrors   int                           `json:"totalErrors"`
	EngineVersion string                        `json:"engineVersion"`
	ReferenceHash string                        `json:"referenceHash"`
}

// PronouncePartDetail keeps the latest detail per item index.
type PronouncePartDetail struct {
	Items map[string]PronounceItemDetail `json:"items"`
}

// newPronounceItemDetail maps one scorer result into the stored shape. The
// scorer's Errors are already gated by the caller; totalErrors is the
// pre-gate count so the UI can say "showing 3 of 7".
func newPronounceItemDetail(res *pronunciation.Result, referenceText string, totalErrors int) PronounceItemDetail {
	h := sha256.Sum256([]byte(referenceText))
	var errs []responses.MispronouncedWord
	_ = utils.MapSlice(&errs, res.Errors)
	return PronounceItemDetail{
		Score:         res.Score,
		PER:           res.PhonemeErrorRate,
		WER:           res.WordErrorRate,
		Errors:        errs,
		TotalErrors:   totalErrors,
		EngineVersion: "openpronounce-v0.3.0",
		ReferenceHash: hex.EncodeToString(h[:])[:16],
	}
}

// ParsePronounceDetail decodes the current shape, then falls back to the
// pre-Plan-03 map[string]float64 blob. Old rows keep their scores; TotalErrors
// stays 0 so the UI never claims an "N of M" it cannot know.
func ParsePronounceDetail(raw []byte) PronouncePartDetail {
	var next PronouncePartDetail
	if err := json.Unmarshal(raw, &next); err == nil && next.Items != nil {
		return next
	}
	var old struct {
		Items map[string]float64 `json:"items"`
	}
	out := PronouncePartDetail{Items: map[string]PronounceItemDetail{}}
	if err := json.Unmarshal(raw, &old); err == nil {
		for k, v := range old.Items {
			out.Items[k] = PronounceItemDetail{Score: v}
		}
	}
	return out
}

// SpeakingScores returns the part-level scores of every speaking part in one
// attempt. The part Score is already the mean of its item scores, so old and
// new detail shapes both work without parsing the detail blob.
func SpeakingScores(m *models.Attempt) []float64 {
	parts := decodePartMap(m)
	keys := make([]string, 0, len(parts))
	for key := range parts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]float64, 0, len(keys))
	for _, key := range keys {
		if parts[key].Type == enums.AttemptTypeSpeaking {
			out = append(out, parts[key].Score)
		}
	}
	return out
}

// WritingPartDetail persists one scored writing part: the normalized
// feedback plus the raw-text receipt. RawText is stored ONLY here, never in
// logs, and is cleared by RedactWriting; RawHash proves what was scored.
type WritingPartDetail struct {
	Feedback  schemas.WritingFeedback `json:"feedback"`
	RawHash   string                  `json:"rawHash"`
	WordCount int                     `json:"wordCount"`
	RawText   string                  `json:"rawText,omitempty"`
}

// decodePartMap reads the attempt result map; a broken blob self-heals to
// empty (last-writer-wins) rather than failing the attach.
func decodePartMap(m *models.Attempt) map[string]PartEntry {
	out := map[string]PartEntry{}
	if len(m.Result) == 0 {
		return out
	}
	_ = json.Unmarshal(m.Result, &out)
	return out
}

func encodePartMap(parts map[string]PartEntry) ([]byte, error) {
	return json.Marshal(parts)
}

// maxWritingWords caps one submission (cost/abuse guard).
const maxWritingWords = 500

func countWords(s string) int { return len(strings.Fields(s)) }

// coverageScore maps verdicts to 0-100: covered 1, partial 0.5, missing 0.
// Spans coach; they never move the number (TOEIC-style task completion).
func coverageScore(v []schemas.TaskVerdict) float64 {
	if len(v) == 0 {
		return 0
	}
	sum := 0.0
	for _, t := range v {
		switch t.Status {
		case "covered":
			sum++
		case "partial":
			sum += 0.5
		}
	}
	return sum / float64(len(v)) * 100
}

// countCovered counts fully covered checklist items for the header strip.
func countCovered(v []schemas.TaskVerdict) int {
	n := 0
	for _, t := range v {
		if t.Status == "covered" {
			n++
		}
	}
	return n
}
