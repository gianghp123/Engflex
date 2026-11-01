package tests

import (
	"encoding/json"
	"fmt"
	"testing"

	"engflex-api/internal/exams/toeic"
	"github.com/stretchr/testify/require"
)

func TestEstimateScore(t *testing.T) {
	// Independent transcription of the supplied table, including its floor.
	values := []float64{5, 5, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55, 60, 65, 70, 75, 80, 85, 90, 95, 100, 105, 110, 115, 120, 125, 130, 135, 140, 145, 150, 155, 160, 165, 170, 175, 180, 185, 190, 195, 200, 205, 210, 215, 220, 225, 230, 235, 240, 245, 250, 255, 260, 265, 270, 275, 280, 285, 290, 295, 300, 305, 310, 315, 320, 325, 330, 335, 340, 345, 350, 355, 360, 365, 370, 375, 380, 385, 390, 395, 400, 405, 410, 415, 420, 425, 430, 435, 440, 445, 450, 455, 460, 465, 470, 475, 480, 485, 490, 495}
	tests := []struct {
		name    string
		raw     int
		want    float64
		invalid bool
	}{{"negative", -1, 0, true}, {"too many", 101, 0, true}}
	for raw, want := range values {
		tests = append(tests, struct {
			name    string
			raw     int
			want    float64
			invalid bool
		}{fmt.Sprint(raw), raw, want, false})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toeic.EstimateScore(tt.raw)
			if tt.invalid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestScoringSnapshotValidation(t *testing.T) {
	tests := []struct {
		name   string
		change func(map[string]any)
		valid  bool
	}{
		{"valid subset", func(map[string]any) {}, true},
		{"legacy", func(m map[string]any) { delete(m, "sections"); delete(m, "scoringVersion") }, true},
		{"missing version", func(m map[string]any) { delete(m, "scoringVersion") }, false},
		{"unknown policy", func(m map[string]any) { m["scoringVersion"] = "other" }, false},
		{"missing sections", func(m map[string]any) { delete(m, "sections") }, false},
		{"partial scaled score", func(m map[string]any) {
			m["sections"].(map[string]any)["listening"].(map[string]any)["scaledScore"] = 5
		}, false},
		{"mismatched correct", func(m map[string]any) { m["correctCount"] = 1 }, false},
		{"mismatched accuracy", func(m map[string]any) { m["accuracyPct"] = 50 }, false},
		{"duplicate total", func(m map[string]any) { m["scaledScore"] = 10 }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m map[string]any
			require.NoError(t, json.Unmarshal([]byte(`{"revision":2,"answers":[{"questionId":"q1"}],"correctCount":0,"accuracyPct":0,"scoringVersion":"toeic-estimate-v1","sections":{"listening":{"questionCount":1,"correctCount":0,"scaledScore":null},"reading":null}}`), &m))
			tt.change(m)
			raw, err := json.Marshal(m)
			require.NoError(t, err)
			_, err = toeic.ParseResult(raw)
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
