package ielts

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"engflex-api/internal/modules/exams/dtos/responses"
	"engflex-api/internal/utils"
)

func TestRichMediaRoundTripAndResponseMapping(t *testing.T) {
	content := sampleIeltsContent()
	audio, image, value := "ielts/academic/test/audio/section-1.mp3", "ielts/academic/test/images/diagram.png", "Keep this cell"
	content.Assets = []Asset{{SourceID: audio, Type: "audio"}, {SourceID: image, Type: "image"}}
	content.Sections[0].AudioAssetID = &audio
	content.Sections[0].Transcript = []Paragraph{{Text: "PRIVATE TRANSCRIPT"}}
	content.Sections[0].QuestionGroups[0].Context = []ContextBlock{
		{Tag: "table", Children: []ContextBlock{{Tag: "tr", Children: []ContextBlock{{Tag: "td", ColSpan: 2, RowSpan: 3, Children: []ContextBlock{{Tag: "text", Text: &value}}}}}}},
		{Tag: "img", AssetID: &image},
	}
	answer := "PRIVATE SOURCE ANSWER"
	content.Sections[0].QuestionGroups[0].Questions[0].SourceAnswer = &answer
	raw, err := json.Marshal(content)
	require.NoError(t, err)
	parsed, err := ParseContent(raw)
	require.NoError(t, err)
	require.Equal(t, 2, parsed.Sections[0].QuestionCount)
	content.Sections[0].QuestionCount = 2
	require.Equal(t, content, *parsed)

	var taking responses.IeltsTakingContent
	require.NoError(t, utils.Map(&taking, parsed))
	takingJSON, err := json.Marshal(taking)
	require.NoError(t, err)
	require.NotContains(t, string(takingJSON), "PRIVATE")
	require.NotContains(t, string(takingJSON), "correctKey")
	require.NotContains(t, string(takingJSON), "explanation")
	require.Equal(t, &audio, taking.Sections[0].AudioAssetID)
	require.Equal(t, 2, taking.Sections[0].QuestionGroups[0].Context[0].Children[0].Children[0].ColSpan)
	require.Equal(t, "image", taking.Assets[1].Type)

	var review responses.IeltsReviewContent
	require.NoError(t, utils.Map(&review, parsed))
	require.Equal(t, "PRIVATE TRANSCRIPT", review.Sections[0].Transcript[0].Text)
	require.Equal(t, &answer, review.Sections[0].QuestionGroups[0].Questions[0].SourceAnswer)
}

func TestRejectInvalidRichMedia(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*Content)
	}{
		{"unresolved section audio", func(c *Content) { id := "missing"; c.Sections[0].AudioAssetID = &id }},
		{"unresolved image", func(c *Content) {
			id := "missing"
			c.Sections[0].QuestionGroups[0].Context = []ContextBlock{{Tag: "img", AssetID: &id}}
		}},
		{"script tag", func(c *Content) { c.Sections[0].QuestionGroups[0].Context = []ContextBlock{{Tag: "script"}} }},
		{"invalid span", func(c *Content) { c.Sections[0].QuestionGroups[0].Context = []ContextBlock{{Tag: "td", ColSpan: -1}} }},
		{"duplicate asset", func(c *Content) { c.Assets = []Asset{{SourceID: "same"}, {SourceID: "same"}} }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := sampleIeltsContent()
			tc.edit(&content)
			raw, err := json.Marshal(content)
			require.NoError(t, err)
			_, err = ParseContent(raw)
			require.Error(t, err)
		})
	}
	raw, err := json.Marshal(sampleIeltsContent())
	require.NoError(t, err)
	_, err = ParseContent(append(raw, []byte(` {"extra": true}`)...))
	require.ErrorContains(t, err, "trailing")
}
