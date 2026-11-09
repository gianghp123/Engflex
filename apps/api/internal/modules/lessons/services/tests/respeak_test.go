package tests

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
	"engflex-api/internal/database/repositories/mocks"
	llmmocks "engflex-api/internal/llm/mocks"
	"engflex-api/internal/modules/lessons/dtos/responses"
	"engflex-api/internal/modules/lessons/services"
	"engflex-api/internal/voice/pronunciation"
	pronunciationmocks "engflex-api/internal/voice/pronunciation/mocks"
)

func TestRespeak(t *testing.T) {
	audio := []byte("fake-webm")

	tests := []struct {
		name       string
		text       string
		setup      func(*pronunciationmocks.MockScorer)
		wantStatus int
		check      func(*testing.T, *responses.RespeakResponse)
	}{
		{
			name: "clean attempt returns no tip",
			text: "think",
			setup: func(scorer *pronunciationmocks.MockScorer) {
				scorer.On("Score", mock.Anything, "think", audio, "audio/webm").
					Return(&pronunciation.Result{Score: 94}, nil).Once()
			},
			check: func(t *testing.T, got *responses.RespeakResponse) {
				t.Helper()
				assert.Equal(t, 94.0, got.Score)
				assert.Empty(t, got.Errors)
				assert.Zero(t, got.TotalErrors)
				assert.Empty(t, got.Tip)
			},
		},
		{
			name: "gated errors keep the pre-gate total",
			text: "think",
			setup: func(scorer *pronunciationmocks.MockScorer) {
				scorer.On("Score", mock.Anything, "think", audio, "audio/webm").
					Return(&pronunciation.Result{
						Score: 62,
						Errors: []pronunciation.MispronouncedWord{
							{Word: "think", Expected: "θɪŋk", Heard: "tɪŋk", Confidence: 0.92},
							{Word: "she", Expected: "ʃiː", Heard: "siː", Confidence: 0.81},
							{Word: "light", Expected: "laɪt", Heard: "naɪt", Confidence: 0.74},
							{Word: "call", Expected: "kɔːl", Heard: "kɑːl", Confidence: 0.40},
						},
					}, nil).Once()
			},
			check: func(t *testing.T, got *responses.RespeakResponse) {
				t.Helper()
				assert.Equal(t, 62.0, got.Score)
				require.Len(t, got.Errors, 3)
				assert.Equal(t, 4, got.TotalErrors)
				assert.Contains(t, got.Tip, "θ")
			},
		},
		{
			name: "empty text is rejected without scoring",
			text: "   ",
			setup: func(scorer *pronunciationmocks.MockScorer) {
				scorer.AssertNotCalled(t, "Score", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "overlong text is rejected without scoring",
			text: strings.Repeat("a", 121),
			setup: func(scorer *pronunciationmocks.MockScorer) {
				scorer.AssertNotCalled(t, "Score", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "scorer error maps through",
			text: "think",
			setup: func(scorer *pronunciationmocks.MockScorer) {
				scorer.On("Score", mock.Anything, "think", audio, "audio/webm").
					Return(nil, &common.AppError{Status: http.StatusUnprocessableEntity, Message: "unscorable audio"}).Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := pronunciationmocks.NewMockScorer(t)
			tt.setup(scorer)
			svc := services.NewLessonActivityService(
				mocks.NewMockLessonActivityRepository(t),
				scorer,
				llmmocks.NewMockLLMClient(t),
				mocks.NewMockAttemptRepository(t),
			)

			got, err := svc.Respeak(context.Background(), tt.text, audio, "audio/webm")

			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			tt.check(t, got)
		})
	}
}
