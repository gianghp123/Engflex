package tests

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/modules/lessons/services"
)

func TestParsePronounceDetail(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want func(*testing.T, services.PronouncePartDetail)
	}{
		{
			name: "new shape round-trips score and errors",
			raw:  `{"items":{"0":{"score":88,"per":0.12,"wer":0.2,"errors":[{"word":"think","tip":"tongue between teeth"}],"totalErrors":2,"engineVersion":"openpronounce-v0.3.0","referenceHash":"abc"}}}`,
			want: func(t *testing.T, got services.PronouncePartDetail) {
				t.Helper()
				item := got.Items["0"]
				assert.Equal(t, 88.0, item.Score)
				assert.Equal(t, 0.12, item.PER)
				assert.Equal(t, 0.2, item.WER)
				assert.Equal(t, 2, item.TotalErrors)
				assert.Equal(t, "openpronounce-v0.3.0", item.EngineVersion)
				assert.Equal(t, "abc", item.ReferenceHash)
				require.Len(t, item.Errors, 1)
				assert.Equal(t, "tongue between teeth", item.Errors[0].Tip)
			},
		},
		{
			name: "old score map still parses",
			raw:  `{"items":{"0":95,"1":80}}`,
			want: func(t *testing.T, got services.PronouncePartDetail) {
				t.Helper()
				assert.Equal(t, 95.0, got.Items["0"].Score)
				assert.Equal(t, 80.0, got.Items["1"].Score)
				assert.Zero(t, got.Items["0"].TotalErrors)
			},
		},
		{
			name: "malformed blob degrades to empty map",
			raw:  `{"items":`,
			want: func(t *testing.T, got services.PronouncePartDetail) {
				t.Helper()
				require.NotNil(t, got.Items)
				assert.Empty(t, got.Items)
			},
		},
		{
			name: "null items degrades to empty map",
			raw:  `null`,
			want: func(t *testing.T, got services.PronouncePartDetail) {
				t.Helper()
				require.NotNil(t, got.Items)
				assert.Empty(t, got.Items)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := services.ParsePronounceDetail(json.RawMessage(tt.raw))
			tt.want(t, got)
		})
	}
}
