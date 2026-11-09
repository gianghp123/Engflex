package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/voice/pronunciation"
)

func TestGateErrors(t *testing.T) {
	word := func(name string, confidence float64, phones ...pronunciation.PhoneDetail) pronunciation.MispronouncedWord {
		return pronunciation.MispronouncedWord{Word: name, Confidence: confidence, Phones: phones}
	}
	phone := func(expected, heard string, confidence float64) pronunciation.PhoneDetail {
		return pronunciation.PhoneDetail{Expected: expected, Heard: heard, Confidence: confidence}
	}

	tests := []struct {
		name  string
		input []pronunciation.MispronouncedWord
		want  []string
	}{
		{
			name:  "empty stays empty",
			input: nil,
			want:  []string{},
		},
		{
			name: "drops words below the word floor",
			input: []pronunciation.MispronouncedWord{
				word("think", pronunciation.MinWordConfidence-0.01),
				word("she", pronunciation.MinWordConfidence),
			},
			want: []string{"she"},
		},
		{
			name: "prunes phones below the phone floor",
			input: []pronunciation.MispronouncedWord{
				word("think", 0.9,
					phone("θ", "t", pronunciation.MinPhoneConfidence-0.01),
					phone("ɪ", "i", pronunciation.MinPhoneConfidence),
				),
			},
			want: []string{"think"},
		},
		{
			name: "orders by confidence descending and truncates",
			input: []pronunciation.MispronouncedWord{
				word("a", 0.70),
				word("b", 0.95),
				word("c", 0.80),
				word("d", 0.90),
			},
			want: []string{"b", "d", "c"},
		},
		{
			name: "stable order for equal confidence",
			input: []pronunciation.MispronouncedWord{
				word("first", 0.9),
				word("second", 0.9),
				word("third", 0.9),
				word("fourth", 0.9),
			},
			want: []string{"first", "second", "third"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pronunciation.GateErrors(tt.input)
			require.NotNil(t, got)
			names := make([]string, 0, len(got))
			for _, e := range got {
				names = append(names, e.Word)
			}
			assert.Equal(t, tt.want, names)
			if tt.name == "prunes phones below the phone floor" {
				require.Len(t, got[0].Phones, 1)
				assert.Equal(t, "ɪ", got[0].Phones[0].Expected)
			}
		})
	}
}
