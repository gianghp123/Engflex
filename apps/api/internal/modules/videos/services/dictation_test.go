package services

import (
	"encoding/json"
	"os"
	"testing"

	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/modules/videos/dtos/responses"

	"github.com/stretchr/testify/require"
)

func TestDictationSharedCases(t *testing.T) {
	data, err := os.ReadFile("testdata/dictation.json")
	require.NoError(t, err)
	var cases []struct {
		Name, Content, Level, Draft string
		Answers                     []string
		Positions                   []int
		Complete                    bool
	}
	require.NoError(t, json.Unmarshal(data, &cases))
	for _, item := range cases {
		t.Run(item.Name, func(t *testing.T) {
			require.Equal(t, item.Positions, dictationPositions(len(dictationWords(item.Content)), item.Level))
			result, err := gradeDictationLine(item.Content, item.Level, requests.DictationLineInput{
				Draft: item.Draft, Answers: item.Answers, Completed: true,
			}, responses.DictationLineResult{})
			if item.Complete {
				require.NoError(t, err)
				require.True(t, result.Completed)
				require.Equal(t, len(item.Positions), *result.Earned)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDictationHintStillCostsWhenCorrected(t *testing.T) {
	result, err := gradeDictationLine("Go now", "hard", requests.DictationLineInput{
		Draft: "go now", Completed: true,
	}, responses.DictationLineResult{Revealed: []int{0}})
	require.NoError(t, err)
	require.Equal(t, 1, *result.Earned)
	require.Equal(t, 2, *result.Total)
}

func TestDictationCheckpointFreezesSubmittedLines(t *testing.T) {
	lines := []*models.VideoTranscript{
		{ID: "first", Content: "Go now"},
		{ID: "second", Content: "Yes"},
		{ID: "punctuation", Content: "—"},
	}
	hard := "hard"
	first := requests.SaveVideoProgress{
		Level:                  &hard,
		CompletedTranscriptIDs: []string{"second"},
		Transcripts: map[string]any{
			"first": requests.DictationLineInput{Draft: "go", Completed: true, RevealAll: true},
		},
	}
	require.NoError(t, mergeDictationProgress(&first, requests.SaveVideoProgress{}, lines))
	require.Equal(t, []string{"first"}, first.CompletedTranscriptIDs)
	late := requests.SaveVideoProgress{
		Transcripts: map[string]any{
			"first":  requests.DictationLineInput{Draft: "go now", Completed: true},
			"second": requests.DictationLineInput{Draft: "yes", Completed: true},
		},
	}
	require.NoError(t, mergeDictationProgress(&late, first, lines))
	require.Equal(t, first.Transcripts["first"], late.Transcripts["first"])
	score, err := dictationScore(late, lines)
	require.NoError(t, err)
	require.NotNil(t, score)
	require.Equal(t, 66.67, *score)

	legacy := requests.SaveVideoProgress{CompletedTranscriptIDs: []string{"first"}}
	resumed := requests.SaveVideoProgress{Transcripts: map[string]any{
		"second": requests.DictationLineInput{Draft: "yes", Completed: true},
	}}
	require.NoError(t, mergeDictationProgress(&resumed, legacy, lines))
	score, err = dictationScore(resumed, lines)
	require.NoError(t, err)
	require.Nil(t, score)
}
