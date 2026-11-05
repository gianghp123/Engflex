// Package services owns ordered lesson parts and lesson attempts.
package services

import (
	"encoding/json"
	"strings"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/lessons/schemas"
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

// PronouncePartDetail keeps the latest score per item index.
type PronouncePartDetail struct {
	Items map[string]float64 `json:"items"`
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
