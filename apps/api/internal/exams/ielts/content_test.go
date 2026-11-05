package ielts

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
)

func sampleIeltsContent() Content {
	choice := "TRUE"
	exp := "Located in Section 1 paragraph A"
	return Content{
		SchemaVersion: "1.0.0",
		TestID:        "gt-test-1",
		Title:         "GT IELTS Reading Test 1",
		Skill:         enums.ExamSkillReading,
		SubType:       enums.ExamSubTypeGeneralTraining,
		Sections: []Section{
			{
				SectionNumber: 1,
				Title:         "Section 1",
				Passages: []Passage{
					{
						ID: "p1",
						Content: []Paragraph{
							{Label: nil, Text: "Welcome to Community Centre..."},
						},
					},
				},
				QuestionGroups: []QuestionGroup{
					{
						ID:   "qg1",
						Type: "true_false_not_given",
						Questions: []Question{
							{
								ID:             "q1",
								QuestionNumber: 1,
								CorrectKey:     &choice,
								Explanation:    &exp,
							},
							{
								ID:             "q2",
								QuestionNumber: 2,
								CorrectAnswers: []string{"kitchen area", "the kitchen area"},
							},
						},
					},
				},
			},
		},
	}
}

func TestParseContent_Valid(t *testing.T) {
	c := sampleIeltsContent()
	raw, err := json.Marshal(c)
	require.NoError(t, err)

	parsed, err := ParseContent(raw)
	require.NoError(t, err)
	require.Equal(t, "gt-test-1", parsed.TestID)
	require.Equal(t, enums.ExamSkillReading, parsed.Skill)
	require.Equal(t, enums.ExamSubTypeGeneralTraining, parsed.SubType)
	require.Len(t, parsed.Sections, 1)
	require.Len(t, parsed.Sections[0].QuestionGroups[0].Questions, 2)
}

func TestParseContent_Invalid(t *testing.T) {
	t.Run("empty bytes", func(t *testing.T) {
		_, err := ParseContent(nil)
		require.Error(t, err)
	})

	t.Run("missing schemaVersion", func(t *testing.T) {
		c := sampleIeltsContent()
		c.SchemaVersion = ""
		raw, _ := json.Marshal(c)
		_, err := ParseContent(raw)
		require.ErrorContains(t, err, "schemaVersion")
	})

	t.Run("missing sections", func(t *testing.T) {
		c := sampleIeltsContent()
		c.Sections = nil
		raw, _ := json.Marshal(c)
		_, err := ParseContent(raw)
		require.ErrorContains(t, err, "at least one section")
	})

	t.Run("duplicate question id", func(t *testing.T) {
		c := sampleIeltsContent()
		c.Sections[0].QuestionGroups[0].Questions[1].ID = "q1"
		raw, _ := json.Marshal(c)
		_, err := ParseContent(raw)
		require.ErrorContains(t, err, "duplicate question id")
	})

	t.Run("duplicate question number", func(t *testing.T) {
		c := sampleIeltsContent()
		c.Sections[0].QuestionGroups[0].Questions[1].QuestionNumber = 1
		raw, _ := json.Marshal(c)
		_, err := ParseContent(raw)
		require.ErrorContains(t, err, "duplicate question number")
	})

	t.Run("missing correct answer definition", func(t *testing.T) {
		c := sampleIeltsContent()
		c.Sections[0].QuestionGroups[0].Questions[0].CorrectKey = nil
		c.Sections[0].QuestionGroups[0].Questions[0].CorrectKeys = nil
		c.Sections[0].QuestionGroups[0].Questions[0].CorrectAnswers = nil
		raw, _ := json.Marshal(c)
		_, err := ParseContent(raw)
		require.ErrorContains(t, err, "no correct answer defined")
	})
}
