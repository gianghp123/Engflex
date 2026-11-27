package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/voice/pronunciation"
)

func TestCoachingTip(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		heard    string
		wantWord string
		want     bool
	}{
		{name: "theta becomes t", expected: "θɪŋk", heard: "tɪŋk", wantWord: "θ", want: true},
		{name: "eth becomes d", expected: "ðeɪ", heard: "deɪ", wantWord: "ð", want: true},
		{name: "esh becomes s", expected: "ʃiː", heard: "siː", wantWord: "ʃ", want: true},
		{name: "affricate becomes tr", expected: "tʃuː", heard: "truː", wantWord: "tʃ", want: true},
		{name: "l becomes n", expected: "laɪt", heard: "naɪt", wantWord: "/l/", want: true},
		{name: "final consonant dropped", expected: "kæt", heard: "kæ", wantWord: "final", want: true},
		{name: "final IPA consonant dropped", expected: "briːð", heard: "briː", wantWord: "final", want: true},
		{name: "clean pair has no tip", expected: "kæt", heard: "kæt", want: false},
		{name: "unknown substitution has no tip", expected: "kɔːl", heard: "kɑːl", want: false},
		{name: "empty input has no tip", expected: "", heard: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pronunciation.CoachingTip(tt.expected, tt.heard)
			if !tt.want {
				assert.Empty(t, got)
				return
			}
			require.NotEmpty(t, got)
			assert.Contains(t, got, tt.wantWord)
		})
	}
}
