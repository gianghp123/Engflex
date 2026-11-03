package tests

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	llmmocks "engflex-api/internal/llm/mocks"
	"engflex-api/internal/modules/conversations/dtos/responses"
	"engflex-api/internal/modules/conversations/services"
	"engflex-api/internal/voice"
	voicemocks "engflex-api/internal/voice/mocks"
)

func newTranscriptService(t *testing.T) (*services.ConversationService, *repomocks.MockConversationRepository, *voicemocks.MockVoiceClient, *voicemocks.MockPronounceEngine) {
	t.Helper()
	repo := repomocks.NewMockConversationRepository(t)
	engine := voicemocks.NewMockVoiceClient(t)
	pronounce := voicemocks.NewMockPronounceEngine(t)
	svc := services.NewConversationService(repo, repomocks.NewMockConversationTurnRepository(t), engine, pronounce, llmmocks.NewMockLLMClient(t), 300,
		repomocks.NewMockFeedbackRepository(t), repomocks.NewMockScenarioRepository(t), repomocks.NewMockPersonaRepository(t))
	return svc, repo, engine, pronounce
}

func ownedBy(repo *repomocks.MockConversationRepository) {
	repo.On("GetByID", mock.Anything, convID).
		Return(&models.Conversation{ID: convID, UserID: userID}, nil)
}

func TestTranscript(t *testing.T) {
	tests := []struct {
		name       string
		action     string
		text       string
		setup      func(*repomocks.MockConversationRepository, *voicemocks.MockVoiceClient)
		wantStatus int
		check      func(*testing.T, *responses.TranscriptResult)
	}{
		{
			name:   "forwards the action and text verbatim",
			action: "send",
			text:   "I went yesterday.",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcript", mock.Anything, convID, "send", "I went yesterday.").
					Return(&responses.TranscriptResult{State: "idle"}, nil)
			},
			check: func(t *testing.T, got *responses.TranscriptResult) {
				assert.Equal(t, "idle", got.State)
			},
		},
		{
			name:   "review returns the engine's copy of the turn",
			action: "review",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcript", mock.Anything, convID, "review", "").
					Return(&responses.TranscriptResult{State: "reviewing", Text: "I go yesterday."}, nil)
			},
			check: func(t *testing.T, got *responses.TranscriptResult) {
				assert.Equal(t, "reviewing", got.State)
				assert.Equal(t, "I go yesterday.", got.Text)
			},
		},
		{
			name:   "foreign conversation is 404",
			action: "review",
			setup: func(repo *repomocks.MockConversationRepository, _ *voicemocks.MockVoiceClient) {
				repo.On("GetByID", mock.Anything, convID).
					Return(&models.Conversation{ID: convID, UserID: "someone_else"}, nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			// The engine's refusals are answers, not outages: the web closes
			// the modal on 404 and can explain 409/422.
			name:   "engine 404 stays 404",
			action: "send",
			text:   "x",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcript", mock.Anything, convID, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/transcript", Status: http.StatusNotFound})
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "engine 409 stays 409",
			action: "send",
			text:   "x",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcript", mock.Anything, convID, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/transcript", Status: http.StatusConflict})
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:   "engine 422 stays 422",
			action: "send",
			text:   "x",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcript", mock.Anything, convID, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/transcript", Status: http.StatusUnprocessableEntity})
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:   "engine 500 is 503",
			action: "review",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcript", mock.Anything, convID, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/transcript", Status: http.StatusInternalServerError})
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:   "transport failure is 503",
			action: "review",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcript", mock.Anything, convID, mock.Anything, mock.Anything).
					Return(nil, errors.New("dial tcp: refused"))
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:   "nil engine reply is 503",
			action: "review",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcript", mock.Anything, convID, mock.Anything, mock.Anything).
					Return(nil, nil)
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, engine, _ := newTranscriptService(t)
			tt.setup(repo, engine)

			got, err := svc.Transcript(context.Background(), userID, convID, tt.action, tt.text)

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

// TestTranscriptDoesNotOwnTheState pins that Go stays a pass-through: the engine
// owns the window, so Go must not second-guess the action or the text.
func TestTranscriptDoesNotOwnTheState(t *testing.T) {
	svc, repo, engine, _ := newTranscriptService(t)
	ownedBy(repo)
	engine.On("Transcript", mock.Anything, convID, "dismiss", "").
		Return(&responses.TranscriptResult{State: "idle"}, nil)

	got, err := svc.Transcript(context.Background(), userID, convID, "dismiss", "")

	require.NoError(t, err)
	assert.Equal(t, "idle", got.State)
}

func TestTranscribe(t *testing.T) {
	audio := []byte("fake-webm-bytes")
	tests := []struct {
		name       string
		audio      []byte
		setup      func(*repomocks.MockConversationRepository, *voicemocks.MockVoiceClient)
		wantStatus int
		check      func(*testing.T, *responses.TranscribeResult)
	}{
		{
			name:  "forwards audio and returns the sentence",
			audio: audio,
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcribe", mock.Anything, convID, audio, "audio/webm").
					Return(&responses.TranscribeResult{Text: "I went yesterday."}, nil)
			},
			check: func(t *testing.T, got *responses.TranscribeResult) {
				assert.Equal(t, "I went yesterday.", got.Text)
			},
		},
		{
			name:  "oversize audio never reaches the engine",
			audio: bytes.Repeat([]byte("x"), services.MaxTranscribeBytes+1),
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
			},
			wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:  "engine 404 stays 404",
			audio: audio,
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/transcribe", Status: http.StatusNotFound})
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "engine 502 stays unavailable",
			audio: audio,
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient) {
				ownedBy(repo)
				engine.On("Transcribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/transcribe", Status: http.StatusBadGateway})
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, engine, _ := newTranscriptService(t)
			tt.setup(repo, engine)
			got, err := svc.Transcribe(context.Background(), userID, convID, tt.audio, "audio/webm")
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

func TestPronounce(t *testing.T) {
	audio := []byte("fake-webm-bytes")
	scored := &responses.PronounceResult{Score: 91.11, Transcription: "I WENT YESTERDAY"}
	tests := []struct {
		name       string
		audio      []byte
		setup      func(*repomocks.MockConversationRepository, *voicemocks.MockPronounceEngine)
		wantStatus int
		check      func(*testing.T, *responses.PronounceResult)
	}{
		{
			name:  "forwards attempt and returns the assessment",
			audio: audio,
			setup: func(repo *repomocks.MockConversationRepository, pronounce *voicemocks.MockPronounceEngine) {
				ownedBy(repo)
				pronounce.On("Pronounce", mock.Anything, "I went yesterday.", "en", audio, "audio/webm").
					Return(scored, nil)
			},
			check: func(t *testing.T, got *responses.PronounceResult) {
				assert.Equal(t, 91.11, got.Score)
			},
		},
		{
			name:  "oversize audio never reaches the engine",
			audio: bytes.Repeat([]byte("x"), services.MaxPronounceBytes+1),
			setup: func(repo *repomocks.MockConversationRepository, pronounce *voicemocks.MockPronounceEngine) {
				ownedBy(repo)
			},
			wantStatus: http.StatusRequestEntityTooLarge,
		},
		{
			name:  "engine 422 stays 422",
			audio: audio,
			setup: func(repo *repomocks.MockConversationRepository, pronounce *voicemocks.MockPronounceEngine) {
				ownedBy(repo)
				pronounce.On("Pronounce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/pronunciation", Status: http.StatusUnprocessableEntity})
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:  "engine 502 stays unavailable",
			audio: audio,
			setup: func(repo *repomocks.MockConversationRepository, pronounce *voicemocks.MockPronounceEngine) {
				ownedBy(repo)
				pronounce.On("Pronounce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, &voice.TranscriptStatusError{Path: "/pronunciation", Status: http.StatusBadGateway})
			},
			wantStatus: http.StatusServiceUnavailable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, pronounce := newTranscriptService(t)
			tt.setup(repo, pronounce)
			got, err := svc.Pronounce(context.Background(), userID, convID, "I went yesterday.", "en", tt.audio, "audio/webm")
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
