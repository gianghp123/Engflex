package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
)

// storedTurnFeedback is the exact blob shape the wire DTO used to write:
// camelCase keys, written by the removed datatypes.JSON path and still seeded
// by apps/web/e2e/helpers.ts. Renaming any model tag would make every stored
// row decode to the zero value, so this fixture is the contract. Old rows
// predate the errorType tag and decode it as "" (treated as other on read).
const storedTurnFeedback = `{
  "corrected": "I have went to the office.",
  "spans": [{"text": "have went", "occurrence": 1, "status": "incorrect", "errorType": "", "correction": "have gone", "reason": "present perfect"}],
  "relevance": {"status": "relevant", "reason": "answers the question"},
  "alternatives": {"language": {"text": "I went to the office.", "reason": "simpler"}, "contextual": null},
  "tip": "Prefer the present perfect for an unfinished time."
}`

func TestTurnFeedbackScan(t *testing.T) {
	tests := []struct {
		name          string
		value         any
		wantCorrected string
		wantSpans     int
		wantFix       string
		wantStatus    enums.RelevanceStatus
		wantLanguage  string
		wantTip       string
	}{
		{
			name:          "stored camelCase blob decodes every field",
			value:         []byte(storedTurnFeedback),
			wantCorrected: "I have went to the office.",
			wantSpans:     1,
			wantFix:       "have gone",
			wantStatus:    enums.RelevanceRelevant,
			wantLanguage:  "I went to the office.",
			wantTip:       "Prefer the present perfect for an unfinished time.",
		},
		{
			name:          "string driver value decodes identically",
			value:         storedTurnFeedback,
			wantCorrected: "I have went to the office.",
			wantSpans:     1,
			wantFix:       "have gone",
			wantStatus:    enums.RelevanceRelevant,
			wantLanguage:  "I went to the office.",
			wantTip:       "Prefer the present perfect for an unfinished time.",
		},
		{
			name:  "NULL leaves the zero value",
			value: nil,
		},
		{
			name:  "empty blob leaves the zero value",
			value: []byte{},
		},
		{
			name:  "malformed JSON leaves the zero value",
			value: []byte("{not json"),
		},
		{
			name:  "unsupported driver type leaves the zero value",
			value: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got models.TurnFeedback
			require.NoError(t, got.Scan(tt.value))

			assert.Equal(t, tt.wantCorrected, got.Corrected)
			assert.Equal(t, tt.wantTip, got.Tip)
			assert.Len(t, got.Spans, tt.wantSpans)
			if tt.wantSpans > 0 {
				assert.Equal(t, "have went", got.Spans[0].Text)
				assert.Equal(t, 1, got.Spans[0].Occurrence)
				assert.Equal(t, enums.SpanStatusIncorrect, got.Spans[0].Status)
				assert.Equal(t, tt.wantFix, got.Spans[0].Correction)
			}
			assert.Equal(t, tt.wantStatus, got.Relevance.Status)
			if tt.wantLanguage == "" {
				assert.Nil(t, got.Alternatives.Language)
			} else if got.Alternatives.Language != nil {
				assert.Equal(t, tt.wantLanguage, got.Alternatives.Language.Text)
			}
		})
	}
}

func TestTurnFeedbackValue(t *testing.T) {
	reason := "answers the question"
	in := models.TurnFeedback{
		Corrected: "I have went to the office.",
		Spans: []models.AnalysisSpan{
			{
				Text:       "have went",
				Occurrence: 1,
				Status:     enums.SpanStatusIncorrect,
				Correction: "have gone",
				Reason:     "present perfect",
			},
		},
		Relevance: models.Relevance{Status: enums.RelevanceRelevant, Reason: &reason},
		Alternatives: models.SpanAlternatives{
			Language: &models.Alternative{Text: "I went to the office.", Reason: "simpler"},
		},
		Tip: "Prefer the present perfect for an unfinished time.",
	}

	raw, err := in.Value()
	require.NoError(t, err)
	encoded, ok := raw.([]byte)
	require.True(t, ok, "Value must return JSON bytes for GORM")

	assert.JSONEq(t, storedTurnFeedback, string(encoded))

	var back models.TurnFeedback
	require.NoError(t, back.Scan(encoded))
	assert.Equal(t, in, back)
}
