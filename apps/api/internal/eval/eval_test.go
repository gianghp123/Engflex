package eval

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/conversations/schemas"
	"engflex-api/internal/modules/conversations/services"
	lessonsschemas "engflex-api/internal/modules/lessons/schemas"
)

// goldenCase is one regression row: learner text with the expected taxonomy
// tags. The harness checksNormalize* accept the fixture and every span
// quotes a substring — never tuning prompts on this set (dev.json is the
// iteration set; test.json would be CI-only).
type goldenCase struct {
	Text     string   `json:"text"`
	Spans    []spanIn `json:"spans"`
	Expected []string `json:"expectedErrorTypes"`
}

type spanIn struct {
	Text       string `json:"text"`
	Occurrence int    `json:"occurrence"`
	ErrorType  string `json:"errorType"`
	Correction string `json:"correction"`
	Reason     string `json:"reason"`
}

func loadGolden(t *testing.T, name string) []goldenCase {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err)
	var out []goldenCase
	require.NoError(t, json.Unmarshal(raw, &out))
	return out
}

func TestWritingGoldenNormalize(t *testing.T) {
	for _, g := range loadGolden(t, "writing_dev.json") {
		spans := make([]lessonsschemas.Span, 0, len(g.Spans))
		for _, s := range g.Spans {
			spans = append(spans, lessonsschemas.Span{
				Text: s.Text, Occurrence: s.Occurrence,
				ErrorType:  enums.ErrorType(s.ErrorType),
				Correction: s.Correction, Reason: s.Reason,
			})
		}
		fb := lessonsschemas.WritingFeedback{
			Task:        lessonsschemas.TaskCompletion{Verdicts: []lessonsschemas.TaskVerdict{{Item: "task", Status: "covered", Evidence: g.Text}}, AnswersPrompt: true, AnswersNote: "ok"},
			Grammar:     spans,
			Expressions: []string{"pattern one", "pattern two", "pattern three"},
			Suggested:   "suggested rewrite",
			Tip:         "one tip",
		}
		got, err := lessonsschemas.NormalizeWritingFeedback(fb, g.Text, 1)
		require.NoError(t, err, "golden text: %q", g.Text)
		require.Len(t, got.Grammar, len(g.Expected))
		for i, want := range g.Expected {
			require.Equal(t, enums.ErrorType(want), got.Grammar[i].ErrorType)
		}
	}
}

func TestFeedbackGoldenNormalize(t *testing.T) {
	for _, g := range loadGolden(t, "feedback_dev.json") {
		spans := make([]models.AnalysisSpan, 0, len(g.Spans))
		for _, s := range g.Spans {
			spans = append(spans, models.AnalysisSpan{
				Text: s.Text, Occurrence: s.Occurrence, Status: enums.SpanStatusIncorrect,
				ErrorType:  enums.ErrorType(s.ErrorType),
				Correction: s.Correction, Reason: s.Reason,
			})
		}
		fb := models.TurnFeedback{Corrected: "corrected", Spans: spans, Tip: "tip"}
		raw, err := json.Marshal(fb)
		require.NoError(t, err)
		got, err := services.NormalizeLanguageFeedback(raw)
		require.NoError(t, err, "golden text: %q", g.Text)
		require.Len(t, got.Spans, len(g.Expected))
		for i, want := range g.Expected {
			require.Equal(t, enums.ErrorType(want), got.Spans[i].ErrorType)
		}
	}
}

func TestSessionSummaryGoldenNormalize(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "session_dev.json"))
	require.NoError(t, err)
	var cases []struct {
		Transcript string                     `json:"transcript"`
		Objectives []models.ScenarioObjective `json:"objectives"`
		LLM        schemas.SessionSummaryLLM  `json:"llm"`
		Want       []models.ObjectiveResult   `json:"want"`
	}
	require.NoError(t, json.Unmarshal(raw, &cases))
	for _, c := range cases {
		got, err := schemas.NormalizeSessionSummary(c.LLM, c.Objectives, c.Transcript)
		require.NoError(t, err)
		require.Equal(t, c.Want, got.Objectives)
	}
}
