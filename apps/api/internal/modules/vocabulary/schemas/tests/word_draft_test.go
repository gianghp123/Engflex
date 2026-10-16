package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/modules/vocabulary/schemas"
)

func TestNormalizeWordDraft(t *testing.T) {
	valid := func() schemas.WordDraft {
		return schemas.WordDraft{
			Definition:   "  Arriving exactly on time. ",
			CEFR:         "B2",
			PartOfSpeech: "adjective",
			Details: schemas.WordDraftDetails{
				IPA: "ˈpʌŋk.tʃu.əl",
				Contexts: []schemas.WordDraftContext{
					{Label: "At work", Quote: "She is always punctual."},
					{Label: "", Quote: "   "},
				},
				Collocations: []schemas.WordDraftCollocation{
					{Phrase: "a punctual start", Pattern: "adj + noun", Example: "We had a punctual start."},
					{Phrase: ""},
				},
				WordForms: []schemas.WordDraftWordForm{
					{FormType: "noun", Forms: []string{"punctuality"}},
					{FormType: "", Forms: nil},
				},
			},
		}
	}

	tests := []struct {
		name    string
		in      func() schemas.WordDraft
		wantErr string
		check   func(t *testing.T, out schemas.WordDraft)
	}{
		{
			name: "valid draft is trimmed and empty rows dropped",
			in:   valid,
			check: func(t *testing.T, out schemas.WordDraft) {
				require.Equal(t, "Arriving exactly on time.", out.Definition)
				require.Len(t, out.Details.Contexts, 1)
				require.Len(t, out.Details.Collocations, 1)
				require.Len(t, out.Details.WordForms, 1)
			},
		},
		{
			name:    "empty definition fails",
			in:      func() schemas.WordDraft { d := valid(); d.Definition = "   "; return d },
			wantErr: "definition",
		},
		{
			name:    "half-step cefr fails",
			in:      func() schemas.WordDraft { d := valid(); d.CEFR = "B1+"; return d },
			wantErr: "cefr",
		},
		{
			name:  "unknown part of speech is dropped",
			in:    func() schemas.WordDraft { d := valid(); d.PartOfSpeech = "preposition-ish"; return d },
			check: func(t *testing.T, out schemas.WordDraft) { require.Empty(t, out.PartOfSpeech) },
		},
		{
			name: "verbose arrays are capped",
			in: func() schemas.WordDraft {
				d := valid()
				for i := 0; i < 9; i++ {
					d.Details.Contexts = append(d.Details.Contexts, schemas.WordDraftContext{Label: "x", Quote: "A sentence."})
				}
				return d
			},
			check: func(t *testing.T, out schemas.WordDraft) {
				require.Len(t, out.Details.Contexts, 4)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := schemas.NormalizeWordDraft(tt.in())
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			tt.check(t, out)
		})
	}
}
