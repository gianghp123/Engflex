package tests

import (
	"encoding/json"
	"os"
	"testing"

	"engflex-api/internal/exams/toeic"
	"engflex-api/internal/modules/exams/dtos/responses"
	"engflex-api/internal/utils"
	"github.com/stretchr/testify/require"
)

func fixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/synthetic.json")
	require.NoError(t, err)
	return b
}

func TestContentValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*toeic.Content)
	}{
		{"valid nullable fields", nil},
		{"version", func(c *toeic.Content) { c.SchemaVersion = "0.2.0" }},
		{"wrong section", func(c *toeic.Content) { c.Parts[0].Section = "reading" }},
		{"duplicate question", func(c *toeic.Content) { c.Parts[0].Items[1].Questions[0].ID = c.Parts[0].Items[0].Questions[0].ID }},
		{"missing question", func(c *toeic.Content) { c.Parts[0].Items = c.Parts[0].Items[1:] }},
		{"part 2 option D", func(c *toeic.Content) {
			c.Parts[1].Items[0].Questions[0].Options = append(c.Parts[1].Items[0].Questions[0].Options, toeic.Option{Key: "D"})
		}},
		{"missing correct key", func(c *toeic.Content) { c.Parts[0].Items[0].Questions[0].CorrectKey = nil }},
		{"missing asset", func(c *toeic.Content) { c.Assets = c.Assets[:1] }},
		{"missing audio start", func(c *toeic.Content) { c.Parts[0].Items[0].ItemAudio.StartTime = nil }},
		{"invalid audio end", func(c *toeic.Content) {
			c.Parts[0].Items[0].ItemAudio.EndTime = c.Parts[0].Items[0].ItemAudio.StartTime
		}},
		{"broken blank reference", func(c *toeic.Content) { c.Parts[5].Items[0].Questions[0].TargetIDs[0] = "missing" }},
		{"filled blank", func(c *toeic.Content) {
			value := "answer"
			c.Parts[5].Items[0].Passages[0].Blocks[0].Inline[0].Text = &value
		}},
		{"unresolved source region", func(c *toeic.Content) { c.Parts[6].Items[0].Passages[0].Blocks[0].Type = "sourceRegion" }},
		{"part 7 missing passage", func(c *toeic.Content) { c.Parts[6].Items[10].Passages = c.Parts[6].Items[10].Passages[:1] }},
		{"missing reading option", func(c *toeic.Content) { c.Parts[4].Items[0].Questions[0].Options[0].Text = nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c toeic.Content
			require.NoError(t, json.Unmarshal(fixture(t), &c))
			if tt.mutate != nil {
				tt.mutate(&c)
			}
			raw, err := json.Marshal(c)
			require.NoError(t, err)
			_, err = toeic.ParseContent(raw)
			if tt.mutate == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMappingAndDisclosure(t *testing.T) {
	content, err := toeic.ParseContent(fixture(t))
	require.NoError(t, err)
	tests := []struct {
		name   string
		taking bool
	}{{"taking", true}, {"completed review", false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.taking {
				var dto responses.TakingContent
				require.NoError(t, utils.Map(&dto, content))
				require.Len(t, dto.Parts, 7)
				require.Equal(t, "q131", dto.Parts[5].Items[0].Questions[0].ID)
				require.Equal(t, "blank131", dto.Parts[5].Items[0].Passages[0].Blocks[0].Inline[0].ID)
				require.NotNil(t, dto.Parts[0].Items[0].ItemAudio)
				b, err := json.Marshal(dto)
				require.NoError(t, err)
				for _, secret := range []string{"correctKey", "explanation", "transcript", "vietnameseText", "vocabulary", "PRIVATE"} {
					require.NotContains(t, string(b), secret)
				}
			} else {
				var dto responses.ReviewContent
				require.NoError(t, utils.Map(&dto, content))
				require.Equal(t, "A", *dto.Parts[0].Items[0].Questions[0].CorrectKey)
				require.Equal(t, "PRIVATE TRANSCRIPT", dto.Parts[0].Items[0].Transcript[0].Text)
				require.Equal(t, "blank131", dto.Parts[5].Items[0].Questions[0].TargetIDs[0])
			}
		})
	}
}

func TestParseRejectsMalformed(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
	}{{"malformed", []byte("{")}, {"unknown field", []byte(`{"schemaVersion":"0.1.0","unexpected":true}`)}, {"trailing JSON", append(fixture(t), []byte(` {}`)...)}, {"null", []byte("null")}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { _, err := toeic.ParseContent(tt.raw); require.Error(t, err) })
	}
}

func TestResultParsing(t *testing.T) {
	tests := []struct {
		name, raw string
		valid     bool
	}{
		{"nullable answer and score", `{"revision":1,"answers":[{"questionId":"q1","selectedKey":null,"flagged":true}],"correctCount":null,"accuracyPct":null,"scaledScore":null}`, true},
		{"legacy scoring placeholders", `{"revision":1,"answers":[{"questionId":"q1","selectedKey":null,"flagged":true}],"sections":null,"scoringVersion":null}`, true},
		{"unknown result field", `{"revision":1,"answers":[],"unexpected":null}`, false},
		{"unknown answer field", `{"revision":1,"answers":[{"questionId":"q1","unexpected":null}]}`, false},
		{"unsupported legacy sections", `{"revision":1,"answers":[],"sections":[]}`, false},
		{"unsupported legacy scoring version", `{"revision":1,"answers":[],"scoringVersion":"unknown"}`, false},
		{"invalid revision", `{"revision":0,"answers":[]}`, false},
		{"missing answers", `{"revision":1}`, false},
		{"duplicate question", `{"revision":1,"answers":[{"questionId":"q1"},{"questionId":"q1"}]}`, false},
		{"invalid option", `{"revision":1,"answers":[{"questionId":"q1","selectedKey":"E"}]}`, false},
		{"invalid score", `{"revision":1,"answers":[],"accuracyPct":101}`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toeic.ParseResult([]byte(tt.raw))
			if !tt.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var dto responses.Result
			require.NoError(t, utils.Map(&dto, result))
			require.Nil(t, dto.Answers[0].SelectedKey)
			require.True(t, dto.Answers[0].Flagged)
			require.Nil(t, dto.ScaledScore)
		})
	}
}

func TestNullableCollectionsRemainCompatibleWithWireTypes(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*toeic.Content)
	}{
		{"absent item lists", func(c *toeic.Content) {
			c.Parts[1].Items[0].Passages = nil
			c.Parts[1].Items[0].Media = nil
			c.Parts[1].Items[0].Transcript = nil
		}},
		{"absent question lists", func(c *toeic.Content) {
			c.Parts[0].Items[0].Questions[0].Vocabulary = nil
			c.Parts[0].Items[0].Questions[0].TargetIDs = nil
		}},
		{"absent block inline", func(c *toeic.Content) { c.Parts[6].Items[0].Passages[0].Blocks[0].Inline = nil }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var original toeic.Content
			require.NoError(t, json.Unmarshal(fixture(t), &original))
			tt.mutate(&original)
			raw, err := json.Marshal(original)
			require.NoError(t, err)
			content, err := toeic.ParseContent(raw)
			require.NoError(t, err)
			var dto responses.ReviewContent
			require.NoError(t, utils.Map(&dto, content))
			encoded, err := json.Marshal(dto)
			require.NoError(t, err)
			for _, field := range []string{`"passages":null`, `"transcript":null`, `"media":null`, `"vocabulary":null`, `"targetIds":null`, `"inline":null`} {
				require.NotContains(t, string(encoded), field)
			}
		})
	}
}
