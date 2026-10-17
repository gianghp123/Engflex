package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/conversations/services"
	"engflex-api/internal/voice"
	voicemocks "engflex-api/internal/voice/mocks"
)

const turnID = "33333333-3333-3333-3333-333333333333"

func userTurn() *models.ConversationTurn {
	return &models.ConversationTurn{
		ID: turnID, ConversationID: convID, Position: 1,
		Role: enums.TurnRoleUser, Text: "I go yesterday.",
	}
}

func newAnalysisService(t *testing.T) (*services.ConversationService, *repomocks.MockConversationRepository, *repomocks.MockConversationTurnRepository, *voicemocks.MockVoiceClient, *repomocks.MockFeedbackRepository) {
	t.Helper()
	repo := repomocks.NewMockConversationRepository(t)
	turnsRepo := repomocks.NewMockConversationTurnRepository(t)
	feedback := repomocks.NewMockFeedbackRepository(t)
	engine := voicemocks.NewMockVoiceClient(t)
	pronounce := voicemocks.NewMockPronounceEngine(t)
	return services.NewConversationService(repo, turnsRepo, engine, pronounce, 300, feedback, repomocks.NewMockScenarioRepository(t), repomocks.NewMockPersonaRepository(t)), repo, turnsRepo, engine, feedback
}

const goodFeedback = `{"corrected":"I went yesterday.","spans":[],"relevance":{"status":"relevant","reason":null},"alternatives":{"language":null,"contextual":null},"tip":"past tense"}`

func TestAnalyzeTurn(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*repomocks.MockConversationRepository, *repomocks.MockConversationTurnRepository, *voicemocks.MockVoiceClient, *repomocks.MockFeedbackRepository)
		wantStatus int
		check      func(*testing.T, *models.Feedback)
	}{
		{
			name: "upserts feedback for the turn",
			setup: func(repo *repomocks.MockConversationRepository, turnsRepo *repomocks.MockConversationTurnRepository, engine *voicemocks.MockVoiceClient, fb *repomocks.MockFeedbackRepository) {
				repo.On("GetByID", mock.Anything, convID).
					Return(&models.Conversation{ID: convID, UserID: userID}, nil)
				turnsRepo.On("GetTurnByPosition", mock.Anything, convID, 1).Return(userTurn(), nil)
				turnsRepo.On("ListTurns", mock.Anything, convID).Return([]*models.ConversationTurn{userTurn()}, nil)
				engine.On("AnalyzeTurn", mock.Anything, mock.MatchedBy(
					func(req voice.AnalyzeTurnRequest) bool {
						return req.Text == "I go yesterday." && req.TurnID == turnID
					})).Return(&voice.AnalyzeTurnResponse{Feedback: json.RawMessage(goodFeedback)}, nil)
				fb.On("Upsert", mock.Anything, mock.Anything).Return(nil)
			},
			check: func(t *testing.T, f *models.Feedback) {
				assert.Equal(t, enums.FeedbackSubjectConversationTurn, f.SubjectType)
				assert.Equal(t, turnID, f.SubjectID)
				assert.Equal(t, userID, f.UserID)
				raw, err := json.Marshal(f.Payload)
				require.NoError(t, err)
				assert.JSONEq(t, goodFeedback, string(raw))
			},
		},
		{
			name: "rejects a bot turn",
			setup: func(repo *repomocks.MockConversationRepository, turnsRepo *repomocks.MockConversationTurnRepository, _ *voicemocks.MockVoiceClient, _ *repomocks.MockFeedbackRepository) {
				repo.On("GetByID", mock.Anything, convID).
					Return(&models.Conversation{ID: convID, UserID: userID}, nil)
				turnsRepo.On("GetTurnByPosition", mock.Anything, convID, 1).Return(&models.ConversationTurn{
					ID: turnID, ConversationID: convID, Role: enums.TurnRoleAI, Text: "hi",
				}, nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "unknown position is 404",
			setup: func(repo *repomocks.MockConversationRepository, turnsRepo *repomocks.MockConversationTurnRepository, _ *voicemocks.MockVoiceClient, _ *repomocks.MockFeedbackRepository) {
				repo.On("GetByID", mock.Anything, convID).
					Return(&models.Conversation{ID: convID, UserID: userID}, nil)
				turnsRepo.On("GetTurnByPosition", mock.Anything, convID, 1).
					Return(nil, gorm.ErrRecordNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "engine failure persists nothing",
			setup: func(repo *repomocks.MockConversationRepository, turnsRepo *repomocks.MockConversationTurnRepository, engine *voicemocks.MockVoiceClient, _ *repomocks.MockFeedbackRepository) {
				repo.On("GetByID", mock.Anything, convID).
					Return(&models.Conversation{ID: convID, UserID: userID}, nil)
				turnsRepo.On("GetTurnByPosition", mock.Anything, convID, 1).Return(userTurn(), nil)
				turnsRepo.On("ListTurns", mock.Anything, convID).Return([]*models.ConversationTurn{userTurn()}, nil)
				engine.On("AnalyzeTurn", mock.Anything, mock.Anything).
					Return(nil, errors.New("engine 500"))
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "empty engine feedback persists nothing",
			setup: func(repo *repomocks.MockConversationRepository, turnsRepo *repomocks.MockConversationTurnRepository, engine *voicemocks.MockVoiceClient, _ *repomocks.MockFeedbackRepository) {
				repo.On("GetByID", mock.Anything, convID).
					Return(&models.Conversation{ID: convID, UserID: userID}, nil)
				turnsRepo.On("GetTurnByPosition", mock.Anything, convID, 1).Return(userTurn(), nil)
				turnsRepo.On("ListTurns", mock.Anything, convID).Return([]*models.ConversationTurn{userTurn()}, nil)
				engine.On("AnalyzeTurn", mock.Anything, mock.Anything).
					Return(&voice.AnalyzeTurnResponse{Feedback: json.RawMessage("")}, nil)
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "non-object feedback persists nothing",
			setup: func(repo *repomocks.MockConversationRepository, turnsRepo *repomocks.MockConversationTurnRepository, engine *voicemocks.MockVoiceClient, _ *repomocks.MockFeedbackRepository) {
				repo.On("GetByID", mock.Anything, convID).
					Return(&models.Conversation{ID: convID, UserID: userID}, nil)
				turnsRepo.On("GetTurnByPosition", mock.Anything, convID, 1).Return(userTurn(), nil)
				turnsRepo.On("ListTurns", mock.Anything, convID).Return([]*models.ConversationTurn{userTurn()}, nil)
				engine.On("AnalyzeTurn", mock.Anything, mock.Anything).
					Return(&voice.AnalyzeTurnResponse{Feedback: json.RawMessage(`[1,2,3]`)}, nil)
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "wrong-shape feedback persists nothing",
			setup: func(repo *repomocks.MockConversationRepository, turnsRepo *repomocks.MockConversationTurnRepository, engine *voicemocks.MockVoiceClient, _ *repomocks.MockFeedbackRepository) {
				repo.On("GetByID", mock.Anything, convID).
					Return(&models.Conversation{ID: convID, UserID: userID}, nil)
				turnsRepo.On("GetTurnByPosition", mock.Anything, convID, 1).Return(userTurn(), nil)
				turnsRepo.On("ListTurns", mock.Anything, convID).Return([]*models.ConversationTurn{userTurn()}, nil)
				// The engine's pydantic validation should have caught this, but
				// Go is the last gate before the database.
				engine.On("AnalyzeTurn", mock.Anything, mock.Anything).
					Return(&voice.AnalyzeTurnResponse{
						Feedback: json.RawMessage(`{"corrected":"a","tip":"t"}`),
					}, nil)
			},
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name: "foreign conversation is 404",
			setup: func(repo *repomocks.MockConversationRepository, turnsRepo *repomocks.MockConversationTurnRepository, _ *voicemocks.MockVoiceClient, _ *repomocks.MockFeedbackRepository) {
				repo.On("GetByID", mock.Anything, convID).
					Return(&models.Conversation{ID: convID, UserID: "someone_else"}, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, turnsRepo, engine, fb := newAnalysisService(t)
			tt.setup(repo, turnsRepo, engine, fb)

			stored, err := svc.AnalyzeTurn(context.Background(), userID, convID, 1, "")

			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, stored)
				return
			}
			require.NoError(t, err)
			tt.check(t, stored)
		})
	}
}

func TestAnalyze_SendsUpToFiveContextTurns(t *testing.T) {
	svc, repo, turnsRepo, engine, fb := newAnalysisService(t)
	turns := []*models.ConversationTurn{}
	for i := 1; i <= 8; i++ {
		turns = append(turns, &models.ConversationTurn{
			ID:       "44444444-4444-4444-4444-44444444444" + string(rune('0'+i)),
			Text:     "prior",
			Role:     enums.TurnRoleAI,
			Position: i,
		})
	}
	target := userTurn()
	target.Position = 9
	turns = append(turns, target)
	repo.On("GetByID", mock.Anything, convID).
		Return(&models.Conversation{ID: convID, UserID: userID}, nil)
	turnsRepo.On("GetTurnByPosition", mock.Anything, convID, 1).Return(userTurn(), nil)
	turnsRepo.On("ListTurns", mock.Anything, convID).Return(turns, nil)
	engine.On("AnalyzeTurn", mock.Anything, mock.MatchedBy(
		func(req voice.AnalyzeTurnRequest) bool { return len(req.Context) == 5 },
	)).Return(&voice.AnalyzeTurnResponse{Feedback: json.RawMessage(goodFeedback)}, nil)
	fb.On("Upsert", mock.Anything, mock.Anything).Return(nil)

	_, err := svc.AnalyzeTurn(context.Background(), userID, convID, 1, "")
	require.NoError(t, err)
}

// TestBuildAnalyzeRequest_CarriesTheResolvedLanguage pins the explanation
// language reaching the engine payload. Weight and resolution live in
// AnalyzeTurn / llm.ExplanationLanguage; this is the hand-off itself.
func TestBuildAnalyzeRequest_CarriesTheResolvedLanguage(t *testing.T) {
	svc, _, turnsRepo, _, _ := newAnalysisService(t)
	turnsRepo.On("GetTurnByPosition", mock.Anything, convID, 1).Return(userTurn(), nil)
	turnsRepo.On("ListTurns", mock.Anything, convID).Return([]*models.ConversationTurn{userTurn()}, nil)

	req, err := svc.BuildAnalyzeRequest(context.Background(), convID, 1, "Vietnamese")
	require.NoError(t, err)
	assert.Equal(t, "Vietnamese", req.Language)
}

func TestAnalyze_RequestCarriesNoLevelOrObjective(t *testing.T) {
	svc, _, turnsRepo, _, _ := newAnalysisService(t)
	// BuildAnalyzeRequest performs no ownership check (AnalyzeTurn does that
	// first), so GetByID is deliberately unmocked here.
	turnsRepo.On("GetTurnByPosition", mock.Anything, convID, 1).Return(userTurn(), nil)
	turnsRepo.On("ListTurns", mock.Anything, convID).Return([]*models.ConversationTurn{userTurn()}, nil)
	// BuildAnalyzeRequest stops before the engine call: engine and Upsert are
	// deliberately unmocked, so any call to them would fail the test.

	req, err := svc.BuildAnalyzeRequest(context.Background(), convID, 1, "")
	require.NoError(t, err)
	raw, err := json.Marshal(req)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "cefr")
	assert.NotContains(t, string(raw), "objective")
}

func TestGetWithFeedback_AttachesFeedbackToTurns(t *testing.T) {
	tests := []struct {
		name      string
		turns     []*models.ConversationTurn
		wantTip   string
		wantFirst bool
	}{
		{
			name: "coaching record rides along with the turn",
			turns: []*models.ConversationTurn{
				{
					ID: turnID, ConversationID: convID, Position: 1,
					Feedback: &models.Feedback{
						SubjectID: turnID,
						Payload:   models.TurnFeedback{Corrected: "I went yesterday.", Tip: "past tense"},
					},
				},
				{ID: "44444444-4444-4444-4444-444444444444", ConversationID: convID, Position: 2, Role: enums.TurnRoleAI},
			},
			wantTip:   "past tense",
			wantFirst: true,
		},
		{
			name: "a corrupt payload leaves the turn without coaching",
			turns: []*models.ConversationTurn{
				{
					ID: turnID, ConversationID: convID, Position: 1,
					Feedback: &models.Feedback{SubjectID: turnID},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, turnsRepo, _, _ := newAnalysisService(t)
			repo.On("GetByID", mock.Anything, convID).
				Return(&models.Conversation{ID: convID, UserID: userID}, nil)
			turnsRepo.On("ListTurnsWithFeedback", mock.Anything, convID).Return(tt.turns, nil)

			_, turns, err := svc.GetWithFeedback(context.Background(), userID, convID)

			require.NoError(t, err)
			require.Len(t, turns, len(tt.turns))
			if tt.wantFirst {
				require.NotNil(t, turns[0].Feedback)
				assert.Equal(t, tt.wantTip, turns[0].Feedback.Payload.Tip)
				assert.Nil(t, turns[1].Feedback, "a turn with no feedback keeps a nil relation")
				return
			}
			assert.Nil(t, turns[0].Feedback)
		})
	}
}
