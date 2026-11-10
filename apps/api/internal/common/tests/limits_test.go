package tests

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
)

func TestSavedNoteBindingLimit(t *testing.T) {
	tests := []struct {
		name     string
		payload  map[string]string
		valid    bool
		wantNote bool
	}{
		{name: "missing field"},
		{name: "empty clears note", payload: map[string]string{"note": ""}, valid: true},
		{name: "whitespace clears note", payload: map[string]string{"note": "  "}, valid: true},
		{name: "maximum length", payload: map[string]string{"note": strings.Repeat("ơ", common.NoteMaxLength)}, valid: true, wantNote: true},
		{name: "over maximum", payload: map[string]string{"note": strings.Repeat("ơ", common.NoteMaxLength+1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)
			var req common.UpdateSavedNote
			err = binding.JSON.BindBody(body, &req)
			if !tt.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantNote, req.Normalized() != nil)
		})
	}
}
