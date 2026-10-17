package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/config"
	"engflex-api/internal/voice"
)

func testVoiceClient(t *testing.T, handler http.HandlerFunc) voice.VoiceClient {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return voice.NewHTTPVoiceClient(config.VoiceConfig{
		ServiceURL:     srv.URL,
		StartTimeout:   2 * time.Second,
		OfferTimeout:   2 * time.Second,
		AnalyzeTimeout: 2 * time.Second,
	})
}

// TestVoiceClientAnalyzeCarriesLanguage pins the one thing the analyze payload
// gained: the resolved language name rides under a `language` key the engine's
// model reads (and safely ignores until it has been taught the field).
func TestVoiceClientAnalyzeCarriesLanguage(t *testing.T) {
	client := testVoiceClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/analyze", r.URL.Path)
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "Vietnamese", body["language"])
		assert.Equal(t, "conv-1", body["conversationId"])
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"feedback":{"corrected":"I went.","tip":"past tense"}}`))
	})

	_, err := client.AnalyzeTurn(context.Background(), voice.AnalyzeTurnRequest{
		ConversationID: "conv-1",
		TurnID:         "turn-1",
		Text:           "I go yesterday.",
		Language:       "Vietnamese",
	})
	require.NoError(t, err)
}

func TestVoiceClientStart(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		wantErr  bool
		wantSess string
	}{
		{name: "ok", status: http.StatusOK, body: `{"sessionId":"eng-1","iceConfig":{"iceServers":[{"urls":["stun:stun.l.google.com:19302"]}]}}`, wantSess: "eng-1"},
		{name: "engine 500", status: http.StatusInternalServerError, body: `{"message":"boom"}`, wantErr: true},
		{name: "malformed json", status: http.StatusOK, body: `{`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := testVoiceClient(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/start", r.URL.Path)
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})

			res, err := client.Start(context.Background(), voice.StartRequest{
				Transport:               "webrtc",
				EnableDefaultICEServers: true,
				Body: voice.VoiceSessionBody{
					UserID: "u1", ConversationID: "c1", MaxDuration: 300,
				},
			})
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSess, res.SessionID)
			require.NotNil(t, res.ICEConfig)
			assert.Len(t, res.ICEConfig.IceServers, 1)
		})
	}
}

func TestVoiceClientOfferPassThrough(t *testing.T) {
	client := testVoiceClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/sessions/eng-1/api/offer", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		raw, _ := json.Marshal(map[string]any{"echo": true})
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(raw)
	})

	body, status, err := client.Offer(context.Background(), "eng-1", http.MethodPost, []byte(`{"sdp":"x"}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, status)
	assert.JSONEq(t, `{"echo": true}`, string(body))
}

func TestVoiceClientTranscribeForwardsTheRealContainer(t *testing.T) {
	// The engine gates on the part's Content-Type. multipart.CreateFormFile
	// labels everything application/octet-stream, which the engine refuses —
	// so the client must set the real container on the part.
	client := testVoiceClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/transcribe", r.URL.Path)
		require.NoError(t, r.ParseMultipartForm(1<<20))
		_, header, err := r.FormFile("audio")
		require.NoError(t, err)
		assert.Equal(t, "audio/wav", header.Header.Get("Content-Type"))
		assert.Equal(t, "c1", r.FormValue("conversationId"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"text":"Well, I went."}`))
	})

	res, err := client.Transcribe(context.Background(), "c1", []byte("fake-wav"), "audio/wav")
	require.NoError(t, err)
	assert.Equal(t, "Well, I went.", res.Text)
}
