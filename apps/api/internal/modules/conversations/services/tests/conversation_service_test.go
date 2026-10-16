package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/conversations/dtos/requests"
	"engflex-api/internal/modules/conversations/dtos/responses"
	"engflex-api/internal/modules/conversations/services"
	"engflex-api/internal/voice"
	voicemocks "engflex-api/internal/voice/mocks"
)

const (
	userID = "user_1"
	convID = "11111111-1111-1111-1111-111111111111"
)

func requireAppError(t *testing.T, err error, status int) {
	t.Helper()
	var appErr *common.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, status, appErr.Status)
}

func newService(t *testing.T) (*services.ConversationService, *repomocks.MockConversationRepository, *repomocks.MockConversationTurnRepository, *voicemocks.MockVoiceClient, *repomocks.MockScenarioRepository) {
	t.Helper()
	repo := repomocks.NewMockConversationRepository(t)
	turns := repomocks.NewMockConversationTurnRepository(t)
	engine := voicemocks.NewMockVoiceClient(t)
	pronounce := voicemocks.NewMockPronounceEngine(t)
	scenarios := repomocks.NewMockScenarioRepository(t)
	return services.NewConversationService(repo, turns, engine, pronounce, 300, repomocks.NewMockFeedbackRepository(t), scenarios, repomocks.NewMockPersonaRepository(t)), repo, turns, engine, scenarios
}

func pendingConv() *models.Conversation {
	return &models.Conversation{ID: convID, UserID: userID, Status: enums.ConversationStatusPending}
}

func TestCreate(t *testing.T) {
	scenarioID := "22222222-2222-2222-2222-222222222222"

	tests := []struct {
		name       string
		req        requests.StartConversation
		setup      func(*repomocks.MockConversationRepository)
		wantStatus int
	}{
		{
			name:       "roleplay without scenario",
			req:        requests.StartConversation{Mode: enums.ConversationModeRoleplay},
			setup:      func(*repomocks.MockConversationRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "free talk atomically closes active rows and creates",
			req:  requests.StartConversation{Mode: enums.ConversationModeFreeTalk},
			setup: func(repo *repomocks.MockConversationRepository) {
				repo.On("CreateClosingActive", mock.Anything, mock.MatchedBy(func(m *models.Conversation) bool {
					return m.UserID == userID && m.Status == enums.ConversationStatusPending
				})).Return(nil).Once()
			},
			wantStatus: 0,
		},
		{
			name: "roleplay with scenario succeeds",
			req:  requests.StartConversation{Mode: enums.ConversationModeRoleplay, ScenarioID: &scenarioID},
			setup: func(repo *repomocks.MockConversationRepository) {
				repo.On("CreateClosingActive", mock.Anything, mock.MatchedBy(func(m *models.Conversation) bool {
					return m.ScenarioID != nil && *m.ScenarioID == scenarioID
				})).Return(nil).Once()
			},
			wantStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, _, _ := newService(t)
			tt.setup(repo)

			m, err := svc.Create(context.Background(), userID, tt.req)
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, m)
		})
	}
}

func TestGetOwnership(t *testing.T) {
	tests := []struct {
		name       string
		conv       *models.Conversation
		dbErr      error
		wantStatus int
	}{
		{name: "not found", dbErr: errNotFound(), wantStatus: http.StatusNotFound},
		{name: "other user", conv: &models.Conversation{ID: convID, UserID: "someone_else", Status: enums.ConversationStatusPending}, wantStatus: http.StatusNotFound},
		{name: "owner", conv: pendingConv(), wantStatus: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, _, _ := newService(t)
			repo.On("GetByID", mock.Anything, convID).Return(tt.conv, tt.dbErr).Once()

			m, err := svc.Get(context.Background(), userID, convID)
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
			require.Equal(t, convID, m.ID)
		})
	}
}

func TestStart(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*repomocks.MockConversationRepository, *voicemocks.MockVoiceClient, *repomocks.MockScenarioRepository)
		wantStatus int
	}{
		{
			name: "cached response skips the engine",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient, _ *repomocks.MockScenarioRepository) {
				raw, _ := json.Marshal(voice.StartResponse{
					SessionID: "eng-1",
					ICEConfig: &responses.IceConfig{IceServers: []responses.IceServer{{URLs: []string{"stun:x"}}}},
				})
				repo.On("GetByID", mock.Anything, convID).Return(&models.Conversation{
					ID: convID, UserID: userID, Status: enums.ConversationStatusPending,
					SpeechSessionID: "eng-1", SpeechStartResponse: raw,
				}, nil).Once()
			},
		},
		{
			name: "fresh start caches the engine response",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient, _ *repomocks.MockScenarioRepository) {
				repo.On("GetByID", mock.Anything, convID).Return(pendingConv(), nil).Once()
				engine.On("Start", mock.Anything, mock.AnythingOfType("voice.StartRequest")).
					Return(&voice.StartResponse{SessionID: "eng-2"}, nil).Once()
				repo.On("SetSpeechStart", mock.Anything, convID, "eng-2", mock.Anything).Return(int64(1), nil).Once()
			},
		},
		{
			name: "missing scenario still starts on the generic prompt",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient, scenarios *repomocks.MockScenarioRepository) {
				scenarioID := "22222222-2222-2222-2222-222222222222"
				repo.On("GetByID", mock.Anything, convID).Return(&models.Conversation{
					ID: convID, UserID: userID, Status: enums.ConversationStatusPending,
					ScenarioID: &scenarioID,
				}, nil).Once()
				scenarios.On("GetByID", mock.Anything, scenarioID).Return(nil, gorm.ErrRecordNotFound).Once()
				engine.On("Start", mock.Anything, mock.MatchedBy(func(req voice.StartRequest) bool {
					return req.Body.Persona == nil && req.Body.Scenario == nil
				})).Return(&voice.StartResponse{SessionID: "eng-3"}, nil).Once()
				repo.On("SetSpeechStart", mock.Anything, convID, "eng-3", mock.Anything).Return(int64(1), nil).Once()
			},
		},
		{
			name: "engine failure is not cached and surfaces 503",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient, _ *repomocks.MockScenarioRepository) {
				repo.On("GetByID", mock.Anything, convID).Return(pendingConv(), nil).Once()
				engine.On("Start", mock.Anything, mock.AnythingOfType("voice.StartRequest")).
					Return(nil, errBoom()).Once()
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "live conversation conflicts",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient, _ *repomocks.MockScenarioRepository) {
				repo.On("GetByID", mock.Anything, convID).Return(&models.Conversation{
					ID: convID, UserID: userID, Status: enums.ConversationStatusLive,
				}, nil).Once()
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "corrupt cache falls through to the engine",
			setup: func(repo *repomocks.MockConversationRepository, engine *voicemocks.MockVoiceClient, _ *repomocks.MockScenarioRepository) {
				repo.On("GetByID", mock.Anything, convID).Return(&models.Conversation{
					ID: convID, UserID: userID, Status: enums.ConversationStatusPending,
					SpeechSessionID: "eng-9", SpeechStartResponse: []byte(`{}`),
				}, nil).Once()
				engine.On("Start", mock.Anything, mock.AnythingOfType("voice.StartRequest")).
					Return(&voice.StartResponse{SessionID: "eng-9"}, nil).Once()
				repo.On("SetSpeechStart", mock.Anything, convID, "eng-9", mock.Anything).Return(int64(1), nil).Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, engine, scenarios := newService(t)
			tt.setup(repo, engine, scenarios)

			_, _, err := svc.Start(context.Background(), userID, convID)
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestOfferTransitions(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		engineBody []byte
		engineCode int
		engineErr  error
		setupRepo  func(*repomocks.MockConversationRepository)
		wantStatus int
		wantHTTP   int
	}{
		{
			name:       "post 2xx marks live",
			method:     http.MethodPost,
			engineBody: []byte(`{"sdp":"answer"}`),
			engineCode: http.StatusOK,
			setupRepo: func(repo *repomocks.MockConversationRepository) {
				repo.On("SetStatus", mock.Anything, convID, enums.ConversationStatusPending, enums.ConversationStatusLive).Return(int64(1), nil).Once()
			},
			wantHTTP: http.StatusOK,
		},
		{
			name:       "post 2xx empty answer marks failed",
			method:     http.MethodPost,
			engineBody: []byte(""),
			engineCode: http.StatusOK,
			setupRepo: func(repo *repomocks.MockConversationRepository) {
				repo.On("SetStatus", mock.Anything, convID, enums.ConversationStatusPending, enums.ConversationStatusFailed).Return(int64(1), nil).Once()
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "patch does not change state",
			method:     http.MethodPatch,
			engineBody: []byte(`{"status":"success"}`),
			engineCode: http.StatusOK,
			setupRepo:  func(*repomocks.MockConversationRepository) {},
			wantHTTP:   http.StatusOK,
		},
		{
			name:       "post 2xx with failing transition surfaces 500",
			method:     http.MethodPost,
			engineBody: []byte(`{"sdp":"answer"}`),
			engineCode: http.StatusOK,
			setupRepo: func(repo *repomocks.MockConversationRepository) {
				repo.On("SetStatus", mock.Anything, convID, enums.ConversationStatusPending, enums.ConversationStatusLive).Return(int64(0), errBoom()).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, engine, _ := newService(t)
			repo.On("GetByID", mock.Anything, convID).Return(&models.Conversation{
				ID: convID, UserID: userID, Status: enums.ConversationStatusPending, SpeechSessionID: "eng-1",
			}, nil).Once()
			tt.setupRepo(repo)
			engine.On("Offer", mock.Anything, "eng-1", tt.method, mock.Anything).
				Return(tt.engineBody, tt.engineCode, tt.engineErr).Once()

			body, status, err := svc.Offer(context.Background(), userID, convID, tt.method, []byte(`{}`))
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantHTTP, status)
			require.NotNil(t, body)
		})
	}
}

func TestEnd(t *testing.T) {
	tests := []struct {
		name       string
		conv       *models.Conversation
		dbErr      error
		setEnded   bool
		wantStatus int
	}{
		{
			name:     "live conversation ends",
			conv:     &models.Conversation{ID: convID, UserID: userID, Status: enums.ConversationStatusLive},
			setEnded: true,
		},
		{
			name:     "already ended is idempotent",
			conv:     &models.Conversation{ID: convID, UserID: userID, Status: enums.ConversationStatusEnded},
			setEnded: true,
		},
		{
			name:       "other user is not found",
			conv:       &models.Conversation{ID: convID, UserID: "someone_else", Status: enums.ConversationStatusLive},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "missing conversation is not found",
			dbErr:      errNotFound(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, _, _ := newService(t)
			repo.On("GetByID", mock.Anything, convID).Return(tt.conv, tt.dbErr).Once()
			if tt.setEnded {
				repo.On("SetEnded", mock.Anything, convID, (*int)(nil)).Return(int64(1), nil).Once()
				ended := &models.Conversation{ID: convID, UserID: userID, Status: enums.ConversationStatusEnded}
				repo.On("GetByID", mock.Anything, convID).Return(ended, nil).Once()
			}

			m, err := svc.End(context.Background(), userID, convID)
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
			require.Equal(t, enums.ConversationStatusEnded, m.Status)
		})
	}
}

func TestFinalize(t *testing.T) {
	tests := []struct {
		name         string
		status       enums.ConversationStatus
		duration     int
		deleted      bool
		deleteErr    error
		wantSetEnded bool
		wantStatus   int
	}{
		{name: "live conversation records duration", status: enums.ConversationStatusLive, duration: 120, wantSetEnded: true},
		{name: "finalize after end keeps duration", status: enums.ConversationStatusEnded, duration: 42, wantSetEnded: true},
		{name: "empty conversation is discarded without SetEnded", status: enums.ConversationStatusLive, duration: 60, deleted: true},
		{name: "empty-check failure surfaces 500", status: enums.ConversationStatusLive, duration: 60, deleteErr: errBoom(), wantStatus: http.StatusInternalServerError},
		{name: "negative duration is rejected", status: enums.ConversationStatusLive, duration: -1, wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, _, _ := newService(t)
			if tt.wantStatus == 0 || tt.deleteErr != nil {
				repo.On("GetByID", mock.Anything, convID).Return(&models.Conversation{
					ID: convID, UserID: userID, Status: tt.status,
				}, nil).Once()
				repo.On("DeleteIfEmpty", mock.Anything, convID).Return(tt.deleted, tt.deleteErr).Once()
			}
			if tt.wantSetEnded {
				repo.On("SetEnded", mock.Anything, convID, mock.MatchedBy(func(d *int) bool {
					return d != nil && *d == tt.duration
				})).Return(int64(1), nil).Once()
			}

			err := svc.Finalize(context.Background(), convID, tt.duration)
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
		})
	}
}

func errNotFound() error { return gorm.ErrRecordNotFound }
func errBoom() error     { return errors.New("boom") }
