package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"engflex-api/internal/llm"
)

// TestExplanationLanguage pins the shared resolver every LLM-triggering
// request runs its code through: assigned codes resolve to the English name
// the prompts interpolate, and everything a model cannot write in is refused
// so the service answers 400 instead of shipping an uninterpretable brief.
func TestExplanationLanguage(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		want   string
		wantOK bool
	}{
		{name: "empty selects the default brief", code: "", want: "", wantOK: true},
		{name: "whitespace is trimmed to the default", code: "  ", want: "", wantOK: true},
		{name: "en resolves to english", code: "en", want: "English", wantOK: true},
		{name: "vi resolves to vietnamese", code: "vi", want: "Vietnamese", wantOK: true},
		{name: "a regional tag keeps the language", code: "vi-VN", want: "Vietnamese", wantOK: true},
		{name: "an unassigned tag is refused", code: "zz", wantOK: false},
		{name: "a malformed tag is refused", code: "not a tag", wantOK: false},
		{name: "und names no language", code: "und", wantOK: false},
		{name: "mul names no language", code: "mul", wantOK: false},
		{name: "zxx names no language", code: "zxx", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := llm.ExplanationLanguage(tt.code)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}
