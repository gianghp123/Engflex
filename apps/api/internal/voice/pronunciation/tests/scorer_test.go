package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
	"engflex-api/internal/voice"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
)

func TestScorer(t *testing.T) {
	audio := []byte("fake-webm")
	want := &voice.PronounceResult{
		Score:            91.11,
		Transcription:    "please call me back",
		PhonemeErrorRate: 0.12,
		WordErrorRate:    0.25,
		AcousticDistance: 3.5,
		Errors: []voice.MispronouncedWord{
			{Word: "call", Expected: "kɔːl", Heard: "kɑːl", Confidence: 0.4},
		},
	}

	tests := []struct {
		name       string
		audio      []byte
		setup      func(*voicemocks.MockPronounceEngine)
		wantStatus int
		check      func(*testing.T, *pronunciation.Result)
	}{
		{
			name:  "happy path maps every field",
			audio: audio,
			setup: func(engine *voicemocks.MockPronounceEngine) {
				engine.On("Pronounce", mock.Anything, "Please call me back.", "en", audio, "audio/webm").
					Return(want, nil).Once()
			},
			check: func(t *testing.T, got *pronunciation.Result) {
				t.Helper()
				require.NotNil(t, got)
				assert.Equal(t, 91.11, got.Score)
				assert.Equal(t, "please call me back", got.Transcription)
				assert.Equal(t, 0.12, got.PhonemeErrorRate)
				assert.Equal(t, 0.25, got.WordErrorRate)
				assert.Equal(t, 3.5, got.AcousticDistance)
				require.Len(t, got.Errors, 1)
				assert.Equal(t, "call", got.Errors[0].Word)
				assert.Equal(t, "kɔːl", got.Errors[0].Expected)
				assert.Equal(t, "kɑːl", got.Errors[0].Heard)
				assert.Equal(t, 0.4, got.Errors[0].Confidence)
			},
		},
		{
			name:  "engine 422 maps to 422",
			audio: audio,
			setup: func(engine *voicemocks.MockPronounceEngine) {
				engine.On("Pronounce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/pronunciation", Status: http.StatusUnprocessableEntity}).Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:  "engine 400 maps to 422",
			audio: audio,
			setup: func(engine *voicemocks.MockPronounceEngine) {
				engine.On("Pronounce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/pronunciation", Status: http.StatusBadRequest}).Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:  "engine 413 maps to 413",
			audio: audio,
			setup: func(engine *voicemocks.MockPronounceEngine) {
				engine.On("Pronounce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/pronunciation", Status: http.StatusRequestEntityTooLarge}).Once()
			},
			wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:  "generic engine error maps to 503",
			audio: audio,
			setup: func(engine *voicemocks.MockPronounceEngine) {
				engine.On("Pronounce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("boom")).Once()
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "oversize audio refused without engine call",
			audio:      make([]byte, pronunciation.MaxAudioBytes+1),
			setup:      func(engine *voicemocks.MockPronounceEngine) {},
			wantStatus: http.StatusRequestEntityTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := voicemocks.NewMockPronounceEngine(t)
			tt.setup(engine)
			got, err := pronunciation.NewScorer(engine).Score(context.Background(), "Please call me back.", tt.audio, "audio/webm")
			if tt.wantStatus != 0 {
				require.Error(t, err)
				var appErr *common.AppError
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.wantStatus, appErr.Status)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			tt.check(t, got)
		})
	}
}
