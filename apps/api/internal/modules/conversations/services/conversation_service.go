package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/conversations/dtos/callbacks"
	"engflex-api/internal/modules/conversations/dtos/requests"
	"engflex-api/internal/utils"
	"engflex-api/internal/voice"
)

var errEmptyAnswer = errors.New("voice engine returned an empty offer answer")

// ConversationService owns the conversation lifecycle rules.
type ScenarioAccessPolicy interface {
	CheckAccess(ctx context.Context, userID, id string) error
}

type ConversationService struct {
	access        ScenarioAccessPolicy
	conversations repositories.ConversationRepository
	turns         repositories.ConversationTurnRepository
	voice         voice.VoiceClient
	pronounce     voice.PronounceEngine
	maxDuration   int
	feedbacks     repositories.FeedbackRepository
	scenarios     repositories.ScenarioRepository
	personas      repositories.PersonaRepository
}

// NewConversationService builds the service over repository + engine interfaces.
func NewConversationService(conversations repositories.ConversationRepository, turns repositories.ConversationTurnRepository, voice voice.VoiceClient, pronounce voice.PronounceEngine, maxDuration int, feedbacks repositories.FeedbackRepository, scenarios repositories.ScenarioRepository, personas repositories.PersonaRepository, access ScenarioAccessPolicy) *ConversationService {
	return &ConversationService{conversations: conversations, turns: turns, voice: voice, pronounce: pronounce, maxDuration: maxDuration, feedbacks: feedbacks, scenarios: scenarios, personas: personas, access: access}
}

// Create provisions a pending conversation. The close of the caller's other
// non-terminal conversations and the insert run atomically inside the
// repository (CreateClosingActive), so concurrent creates cannot both slip
// through — the row lock serializes them.
func (s *ConversationService) Create(ctx context.Context, userID string, req requests.StartConversation) (*models.Conversation, error) {
	if req.Mode != enums.ConversationModeFreeTalk && req.Mode != enums.ConversationModeRoleplay {
		return nil, common.BadRequest("invalid conversation mode")
	}
	if req.Mode == enums.ConversationModeRoleplay && (req.ScenarioID == nil || *req.ScenarioID == "") {
		return nil, common.BadRequest("scenarioId is required for roleplay")
	}

	if req.ScenarioID != nil {
		if err := s.access.CheckAccess(ctx, userID, *req.ScenarioID); err != nil {
			return nil, err
		}
	}
	m := &models.Conversation{}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map start conversation failed", err, "userID", userID)
		return nil, common.Internal()
	}
	m.UserID = userID
	m.Status = enums.ConversationStatusPending
	if err := s.conversations.CreateClosingActive(ctx, m); err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "create conversation failed", appErr, "userID", userID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "conversation created", "conversationID", m.ID, "mode", m.Mode)
	return m, nil
}

// Get returns the caller's conversation.
func (s *ConversationService) Get(ctx context.Context, userID, id string) (*models.Conversation, error) {
	return s.getOwned(ctx, userID, id)
}

// List returns the caller's finished conversations (ended or failed),
// newest first, with scenario + persona attached for the history row.
func (s *ConversationService) List(ctx context.Context, userID string, scenarioID *string, limit, offset int) ([]*models.Conversation, int64, error) {
	rows, err := s.conversations.ListByUser(ctx, userID, scenarioID, limit, offset)
	if err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "list conversations failed", appErr, "userID", userID)
		return nil, 0, appErr
	}
	total, err := s.conversations.CountByUser(ctx, userID, scenarioID)
	if err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "count conversations failed", appErr, "userID", userID)
		return nil, 0, appErr
	}
	slog.InfoContext(ctx, "conversations listed", "userID", userID, "count", len(rows), "total", total)
	return rows, total, nil
}

// Start provisions the engine session (idempotent) and returns the ICE config.
func (s *ConversationService) Start(ctx context.Context, userID, id string) (*models.Conversation, *voice.IceConfig, error) {
	conv, err := s.getOwned(ctx, userID, id)
	if err != nil {
		return nil, nil, err
	}
	if err := s.checkAccess(ctx, userID, conv); err != nil {
		return nil, nil, err
	}

	if conv.Status != enums.ConversationStatusPending {
		return nil, nil, common.Conflict("conversation is not pending")
	}

	persona, scenario := s.promptInputs(ctx, conv)
	res, err := s.voice.Start(ctx, voice.StartRequest{
		Transport:               "webrtc",
		EnableDefaultICEServers: true,
		Body: voice.VoiceSessionBody{
			UserID:         userID,
			ConversationID: conv.ID,
			MaxDuration:    s.maxDuration,
			Persona:        persona,
			Scenario:       scenario,
		},
	})
	if err != nil {
		logger.Report(ctx, "voice engine start failed", err, "conversationID", conv.ID)
		return nil, nil, common.ServiceUnavailable("voice service unavailable")
	}

	raw, err := json.Marshal(res)
	if err != nil {
		logger.Report(ctx, "marshal voice start response failed", err, "conversationID", conv.ID)
		return nil, nil, common.Internal()
	}
	affected, err := s.conversations.SetSpeechStart(ctx, conv.ID, res.SessionID, raw)
	if err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "cache voice start failed", appErr, "conversationID", conv.ID)
		return nil, nil, appErr
	}
	if affected == 0 {
		return nil, nil, common.Conflict("conversation start raced")
	}

	conv.SpeechSessionID = res.SessionID
	conv.SpeechStartResponse = raw
	slog.InfoContext(ctx, "conversation start provisioned", "conversationID", conv.ID, "engineSessionID", res.SessionID)
	return conv, res.ICEConfig, nil
}

// Offer proxies SDP/ICE to the engine and applies the state transition on POST.
func (s *ConversationService) Offer(ctx context.Context, userID, id, method string, body []byte) ([]byte, int, error) {
	conv, err := s.getOwned(ctx, userID, id)
	if err != nil {
		return nil, 0, err
	}
	if err := s.checkAccess(ctx, userID, conv); err != nil {
		return nil, 0, err
	}

	if conv.Status == enums.ConversationStatusEnded || conv.Status == enums.ConversationStatusFailed {
		return nil, 0, common.Conflict("conversation is closed")
	}
	if conv.SpeechSessionID == "" {
		return nil, 0, common.Conflict("conversation not started")
	}

	respBody, status, err := s.voice.Offer(ctx, conv.SpeechSessionID, method, body)
	if err != nil {
		logger.Report(ctx, "voice engine offer failed", err, "conversationID", conv.ID)
		return nil, 0, common.ServiceUnavailable("voice service unavailable")
	}

	if method != http.MethodPost {
		return respBody, status, nil
	}

	ok := status >= 200 && status < 300 && len(bytes.TrimSpace(respBody)) > 0
	to := enums.ConversationStatusLive
	if !ok {
		to = enums.ConversationStatusFailed
	}
	if conv.Status == enums.ConversationStatusPending {
		if _, err := s.conversations.SetStatus(ctx, conv.ID, enums.ConversationStatusPending, to); err != nil {
			// The engine already answered, but we could not record it.
			// Fail loudly so the client retries instead of proceeding
			// on a state the database does not reflect. Retry is safe:
			// Start replays the cache and Offer re-applies the transition.
			appErr := common.FromDBError(err, "conversation")
			logger.Report(ctx, "update conversation status failed", appErr, "conversationID", conv.ID)
			return nil, 0, appErr
		}
	}
	if !ok {
		if status >= 200 && status < 300 {
			logger.Report(ctx, "voice engine returned an empty answer", errEmptyAnswer, "conversationID", conv.ID)
			return nil, 0, common.ServiceUnavailable("voice service unavailable")
		}
		return respBody, status, nil
	}
	return respBody, status, nil
}

// End closes a conversation; idempotent for terminal rows.
func (s *ConversationService) End(ctx context.Context, userID, id string) (*models.Conversation, error) {
	conv, err := s.getOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.conversations.SetEnded(ctx, conv.ID, nil); err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "end conversation failed", appErr, "conversationID", conv.ID)
		return nil, appErr
	}
	updated, err := s.conversations.GetByID(ctx, conv.ID)
	if err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "get ended conversation failed", appErr, "conversationID", conv.ID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "conversation ended", "conversationID", conv.ID)
	return updated, nil
}

// Finalize commits the final transcript and lifecycle in one database transaction.
func (s *ConversationService) Finalize(ctx context.Context, id string, req callbacks.FinalizeConversation) error {
	if req.DurationSec < 0 {
		return common.BadRequest("durationSec must be >= 0")
	}
	_, err := s.applySnapshot(ctx, id, req.IngestTurns, &req.DurationSec)
	return err
}

func (s *ConversationService) IngestTurns(ctx context.Context, id string, req callbacks.IngestTurns) (int, error) {
	_, err := s.applySnapshot(ctx, id, req, nil)
	if err != nil {
		return 0, err
	}
	return len(req.Turns), nil
}

func (s *ConversationService) applySnapshot(ctx context.Context, id string, req callbacks.IngestTurns, duration *int) (*models.Conversation, error) {
	if req.EngineSessionID == "" || req.Revision < 1 {
		return nil, common.BadRequest("session identity and positive revision required")
	}
	if req.Turns == nil || len(req.Turns) > common.MaxTurnsPerBatch {
		return nil, common.BadRequest("invalid snapshot turn count")
	}
	for i, turn := range req.Turns {
		if turn.Position != i+1 || (turn.Role != enums.TurnRoleUser && turn.Role != enums.TurnRoleAI) || strings.TrimSpace(turn.Text) == "" || utf8.RuneCountInString(turn.Text) > 10000 {
			return nil, common.BadRequest("invalid snapshot turn")
		}
	}
	raw, err := json.Marshal(req.Turns)
	if err != nil {
		logger.Report(ctx, "snapshot hash failed", err)
		return nil, common.Internal()
	}
	hash := sha256.Sum256(raw)
	var turns []*models.ConversationTurn
	if err := utils.MapSlice(&turns, req.Turns); err != nil {
		logger.Report(ctx, "map snapshot failed", err)
		return nil, common.Internal()
	}
	for _, turn := range turns {
		turn.ConversationID = id
	}
	out, err := s.conversations.ApplySnapshot(ctx, id, req.EngineSessionID, req.Revision, hex.EncodeToString(hash[:]), turns, duration)
	if err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "snapshot rejected", appErr, "conversationID", id, "revision", req.Revision)
		return nil, appErr
	}
	slog.InfoContext(ctx, "snapshot acknowledged", "conversationID", id, "revision", req.Revision, "turns", len(turns), "final", duration != nil)
	return out, nil
}

// promptInputs resolves the conversation's scenario, persona, and learner for
// the session prompt. A missing scenario or persona is not fatal: log a
// warning and fall back to the engine's generic free-talk prompt, because a
// learner who clicked through should still be able to talk.
func (s *ConversationService) promptInputs(ctx context.Context, conv *models.Conversation) (*voice.VoicePersonaBody, *voice.VoiceScenarioBody) {
	if conv.ScenarioID == nil || *conv.ScenarioID == "" {
		return nil, nil
	}
	scenario, err := s.scenarios.GetByID(ctx, *conv.ScenarioID)
	if err != nil || scenario == nil {
		logger.Report(ctx, "resolve scenario failed", common.NotFound("scenario not found"), "scenarioID", *conv.ScenarioID)
		return nil, nil
	}
	var scenarioBody voice.VoiceScenarioBody
	_ = utils.Map(&scenarioBody, scenario)
	out := &scenarioBody
	if scenario.PersonaID == nil || *scenario.PersonaID == "" {
		return nil, out
	}
	persona, err := s.personas.GetByID(ctx, *scenario.PersonaID)
	if err != nil || persona == nil {
		logger.Report(ctx, "resolve persona failed", common.NotFound("persona not found"), "personaID", *scenario.PersonaID)
		return nil, out
	}
	var personaBody voice.VoicePersonaBody
	_ = utils.Map(&personaBody, persona)
	return &personaBody, out
}

func (s *ConversationService) getOwned(ctx context.Context, userID, id string) (*models.Conversation, error) {
	m, err := s.conversations.GetByID(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "conversation")
		if appErr.Status >= http.StatusInternalServerError {
			logger.Report(ctx, "get conversation failed", appErr, "conversationID", id)
		}
		return nil, appErr
	}
	if appErr := common.RequireOwner(m.UserID, userID, "conversation"); appErr != nil {
		return nil, appErr
	}
	return m, nil
}

func decodeStartResponse(raw []byte) (*voice.StartResponse, bool) {
	var out voice.StartResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, false
	}
	if out.SessionID == "" {
		return nil, false
	}
	return &out, true
}

func (s *ConversationService) checkAccess(ctx context.Context, userID string, conv *models.Conversation) error {
	if conv.ScenarioID != nil {
		return s.access.CheckAccess(ctx, userID, *conv.ScenarioID)
	}
	return nil
}
