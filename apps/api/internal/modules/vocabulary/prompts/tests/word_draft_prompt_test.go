package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/modules/vocabulary/prompts"
)

func TestBuildWordDraftPrompt(t *testing.T) {
	tests := []struct {
		name     string
		term     string
		context  string
		language string
		contains []string
		absent   []string
	}{
		{
			name:    "term and context ride the prompt",
			term:    "punctual",
			context: "She is always punctual.",
			contains: []string{
				"She is always punctual.",
				"Write everything in English.",
				"one learner-friendly English sentence",
			},
		},
		{
			name: "no context line when none was given",
			term: "punctual",
			// The phrase also lives in the few-shot example; pin the dynamic
			// line instead — it can only appear for the term being asked about.
			absent: []string{"punctual\nThe learner met it"},
		},
		{
			name:     "english name keeps the default brief",
			term:     "punctual",
			language: "English",
			contains: []string{
				"Write everything in English.",
				"one learner-friendly English sentence",
			},
			absent: []string{
				"Write the explanations in",
				"The example below is in English only",
			},
		},
		{
			name:     "vietnamese localizes the rules and swaps in the calibrated example",
			term:     "reliable",
			language: "Vietnamese",
			contains: []string{
				"Write the explanations in Vietnamese",
				"one learner-friendly Vietnamese sentence",
				"Người học Việt Nam hay nhấn sai âm tiết đầu",
			},
			absent: []string{
				"Write everything in English.",
				"Able to be trusted to do what is expected",
			},
		},
		{
			name:     "another language keeps the english example with a structure note",
			term:     "punctual",
			language: "French",
			contains: []string{
				"Write the explanations in French",
				"one learner-friendly French sentence",
				"The example below is in English only to show the JSON shape",
			},
			absent: []string{"Write everything in English."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prompts.BuildWordDraftPrompt(tt.term, tt.context, tt.language)
			require.Contains(t, got, tt.term)
			if tt.context != "" {
				require.Contains(t, got, tt.context)
			}
			for _, want := range tt.contains {
				require.Contains(t, got, want)
			}
			for _, no := range tt.absent {
				require.NotContains(t, got, no)
			}
			require.Contains(t, got, "NEVER invent")
			// The few-shot block is part of the tested artifact — pin it.
			require.Contains(t, got, "Examples:")
			require.Contains(t, got, "reliable")
		})
	}
}
