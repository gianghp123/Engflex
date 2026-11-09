package ielts

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  Hello World  ", "hello world"},
		{"KITCHEN   AREA.", "kitchen area"},
		{" 15% , ", "15%"},
		{"word...", "word"},
		{"First, Second; Third!", "first, second; third"},
	}

	for _, tt := range tests {
		require.Equal(t, tt.expected, NormalizeText(tt.input))
	}
}

func TestGradeIeltsAttempt(t *testing.T) {
	keyA := "A"
	exp := "explanation"
	content := &Content{
		SchemaVersion: "1.0.0",
		TestID:        "test-grade",
		Title:         "IELTS Test Grade",
		Skill:         enums.ExamSkillReading,
		SubType:       enums.ExamSubTypeGeneralTraining,
		Sections: []Section{
			{
				SectionNumber: 1,
				QuestionGroups: []QuestionGroup{
					{
						ID: "qg1",
						Questions: []Question{
							// Q1: Single choice
							{ID: "q1", QuestionNumber: 1, CorrectKey: &keyA, Explanation: &exp},
							// Q2: Multi-select
							{ID: "q2", QuestionNumber: 2, CorrectKeys: []string{"B", "D"}},
							// Q3: Text fill
							{ID: "q3", QuestionNumber: 3, CorrectAnswers: []string{"kitchen area", "the kitchen area"}},
							// Q4: Text fill (unanswered)
							{ID: "q4", QuestionNumber: 4, CorrectAnswers: []string{"takeaway"}},
						},
					},
				},
			},
		},
	}

	ans1 := "a" // lowercase matching "A"
	ans2 := []string{"d", "B"} // reverse order, different case
	ans3 := "  The   Kitchen Area.  " // dirty whitespace + punctuation + casing
	wrongAns := "wrong"

	answers := []Answer{
		{QuestionID: "q1", SelectedKey: &ans1},
		{QuestionID: "q2", SelectedKeys: ans2},
		{QuestionID: "q3", TextValue: &ans3},
		{QuestionID: "q4", TextValue: &wrongAns},
	}

	res, err := GradeIeltsAttempt(content, answers)
	require.NoError(t, err)
	require.Equal(t, 4, res.RawTotal)
	require.Equal(t, 3, res.RawCorrect)
	require.Equal(t, 75.0, res.AccuracyPct)
	require.Equal(t, 2.0, res.BandScore) // GT 3 correct = 2.0
	require.Equal(t, 3, res.SectionScores[1].RawCorrect)
	require.Equal(t, 4, res.SectionScores[1].RawTotal)
}

func TestParseResult(t *testing.T) {
	t.Run("valid result", func(t *testing.T) {
		res := Result{
			Revision:   1,
			RawCorrect: 30,
			RawTotal:   40,
			BandScore:  6.0,
		}
		raw, err := json.Marshal(res)
		require.NoError(t, err)

		parsed, err := ParseResult(raw)
		require.NoError(t, err)
		require.Equal(t, 30, parsed.RawCorrect)
		require.Equal(t, 6.0, parsed.BandScore)
	})

	t.Run("empty bytes", func(t *testing.T) {
		_, err := ParseResult(nil)
		require.Error(t, err)
	})
}
