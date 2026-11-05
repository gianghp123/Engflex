package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/modules/lessons/activityconfig"
)

func TestParseReadingConfig(t *testing.T) {
	raw := datatypes.JSON(`{"text":"Office memo","questions":[{"stem":"s","instruction":"pick","options":[{"key":"A","text":"a"},{"key":"B","text":"b"}],"answerKey":"B","explanation":"because"}]}`)
	tests := []struct {
		name    string
		raw     datatypes.JSON
		wantErr bool
		check   func(t *testing.T, cfg activityconfig.ReadingConfig)
	}{
		{name: "valid keeps hidden fields", raw: raw, check: func(t *testing.T, cfg activityconfig.ReadingConfig) {
			require.Len(t, cfg.Questions, 1)
			assert.Equal(t, "B", cfg.Questions[0].AnswerKey)
			assert.Equal(t, "because", cfg.Questions[0].Explanation)
			assert.Len(t, cfg.Questions[0].Options, 2)
		}},
		{name: "empty blob errors", raw: datatypes.JSON(``), wantErr: true},
		{name: "broken json errors", raw: datatypes.JSON(`{`), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := activityconfig.ParseReadingConfig(tt.raw)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, cfg)
		})
	}
}

func TestParseListeningConfigTranscript(t *testing.T) {
	raw := datatypes.JSON(`{"audioKey":"listening/abc.mp3","script":[{"speaker":"W","voice":"en-US-JennyNeural","text":"Good morning."},{"voice":"en-US-GuyNeural","text":"Hi there."}],"questions":[]}`)
	cfg, err := activityconfig.ParseListeningConfig(raw)
	require.NoError(t, err)
	assert.Equal(t, "listening/abc.mp3", cfg.AudioKey)
	assert.Equal(t, "W: Good morning.\nHi there.", cfg.Transcript())
}

func TestParseWritingAndSpeakingConfig(t *testing.T) {
	w, err := activityconfig.ParseWritingConfig(datatypes.JSON(`{"task":"email","instructions":"i","stimulus":"s","minWords":40,"maxWords":60,"modelAnswer":"m","checklist":["c1"]}`))
	require.NoError(t, err)
	assert.Equal(t, "email", w.Task)
	assert.Equal(t, []string{"c1"}, w.Checklist)

	s, err := activityconfig.ParseSpeakingConfig(datatypes.JSON(`{"items":[{"text":"Hello.","modelAudioKey":"speaking/x/0.mp3","modelVoice":"en-GB-SoniaNeural"}]}`))
	require.NoError(t, err)
	require.Len(t, s.Items, 1)
	assert.Equal(t, "en-GB-SoniaNeural", s.Items[0].ModelVoice)
}

func TestQuestionsFor(t *testing.T) {
	reading := datatypes.JSON(`{"text":"t","questions":[{"stem":"s","instruction":"i","options":[{"key":"A","text":"a"}],"answerKey":"A","explanation":"e"}]}`)
	qs, err := activityconfig.QuestionsFor(enums.ActivityTypeReading, reading)
	require.NoError(t, err)
	require.Len(t, qs, 1)

	listening := datatypes.JSON(`{"audioKey":"k","script":[],"questions":[{"stem":"s","instruction":"i","options":[{"key":"A","text":"a"}],"answerKey":"A","explanation":"e"}]}`)
	qs, err = activityconfig.QuestionsFor(enums.ActivityTypeListening, listening)
	require.NoError(t, err)
	require.Len(t, qs, 1)

	_, err = activityconfig.QuestionsFor(enums.ActivityTypeWriting, datatypes.JSON(`{}`))
	require.ErrorIs(t, err, activityconfig.ErrNotCheckable)

	_, err = activityconfig.QuestionsFor(enums.ActivityTypeReading, datatypes.JSON(`{`))
	require.Error(t, err)
}
