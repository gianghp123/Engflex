package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/modules/lessons/schemas"
)

func validWritingOutput() schemas.WritingFeedback {
	return schemas.WritingFeedback{
		Task: schemas.TaskCompletion{
			Verdicts: []schemas.TaskVerdict{
				{Item: "You greeted the team", Status: "covered", Evidence: "Hello everyone"},
				{Item: "You gave your name and role", Status: "covered", Evidence: "my name is Kenji"},
			},
			AnswersPrompt: true,
			AnswersNote:   "Answers the prompt directly.",
		},
		Grammar: []schemas.Span{
			{Text: "I is", Occurrence: 1, Correction: "I am", Reason: "Be verb agrees with I as am"},
		},
		Phrasing: []schemas.Span{
			{Text: "a sales assistant", Occurrence: 1, Correction: "a sales associate", Reason: "More natural job title"},
		},
		Expressions: []string{"my name is", "I work as", "nice to meet you"},
		Suggested:   "Hello everyone, my name is Kenji and I am a sales associate.",
		Tip:         "Remember the be verb: I am, not I is.",
	}
}

func TestNormalizeWritingFeedback(t *testing.T) {
	text := "Hello everyone, my name is Kenji and I is a sales assistant."
	tests := []struct {
		name         string
		mutate       func(*schemas.WritingFeedback)
		checklistLen int
		wantErr      string
	}{
		{name: "valid payload passes", mutate: func(o *schemas.WritingFeedback) {}, checklistLen: 2},
		{name: "missing suggested errors", mutate: func(o *schemas.WritingFeedback) { o.Suggested = "  " }, checklistLen: 2, wantErr: "suggested or tip"},
		{name: "missing tip errors", mutate: func(o *schemas.WritingFeedback) { o.Tip = "" }, checklistLen: 2, wantErr: "suggested or tip"},
		{name: "grammar span not in text errors", mutate: func(o *schemas.WritingFeedback) {
			o.Grammar[0].Text = "not in text"
		}, checklistLen: 2, wantErr: "grammar span 0 is not an exact substring"},
		{name: "phrasing span not in text errors", mutate: func(o *schemas.WritingFeedback) {
			o.Phrasing[0].Text = "not in text"
		}, checklistLen: 2, wantErr: "phrasing span 0 is not an exact substring"},
		{name: "verdict count mismatch errors", mutate: func(o *schemas.WritingFeedback) {
			o.Task.Verdicts = o.Task.Verdicts[:1]
		}, checklistLen: 2, wantErr: "verdicts 1 != checklist 2"},
		{name: "verdict unknown status errors", mutate: func(o *schemas.WritingFeedback) {
			o.Task.Verdicts[0].Status = "done"
		}, checklistLen: 2, wantErr: "unknown status"},
		{name: "evidence absent errors", mutate: func(o *schemas.WritingFeedback) {
			o.Task.Verdicts[0].Evidence = "absent quote"
		}, checklistLen: 2, wantErr: "evidence is not a substring"},
		{name: "zero expressions errors", mutate: func(o *schemas.WritingFeedback) {
			o.Expressions = nil
		}, checklistLen: 2, wantErr: "no expressions"},
		{name: "empty grammar and phrasing pass", mutate: func(o *schemas.WritingFeedback) {
			o.Grammar = nil
			o.Phrasing = nil
		}, checklistLen: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := validWritingOutput()
			tt.mutate(&out)
			got, err := schemas.NormalizeWritingFeedback(out, text, tt.checklistLen)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			want := out
			if want.Grammar == nil {
				want.Grammar = []schemas.Span{}
			}
			if want.Phrasing == nil {
				want.Phrasing = []schemas.Span{}
			}
			require.Equal(t, want, got)
			assert.NotNil(t, got.Grammar)
			assert.NotNil(t, got.Phrasing)
		})
	}
}

func assertStrictNode(t *testing.T, node map[string]any) {
	t.Helper()
	props, ok := node["properties"].(map[string]any)
	require.True(t, ok, "object level without properties")
	assert.Equal(t, false, node["additionalProperties"])
	required, ok := node["required"].([]any)
	require.True(t, ok, "object level without required")
	got := map[string]bool{}
	for _, r := range required {
		got[r.(string)] = true
	}
	for name, sub := range props {
		assert.True(t, got[name], "property %s missing from required", name)
		if child, ok := sub.(map[string]any); ok {
			if _, hasProps := child["properties"]; hasProps {
				assertStrictNode(t, child)
			}
			if items, ok := child["items"].(map[string]any); ok {
				if _, hasProps := items["properties"]; hasProps {
					assertStrictNode(t, items)
				}
			}
		}
	}
}

func TestWritingFeedbackSchemaIsStrictCompliant(t *testing.T) {
	assertStrictNode(t, schemas.WritingFeedbackSchema)
}
