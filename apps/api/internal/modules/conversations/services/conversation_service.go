package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"gorm.io/datatypes"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/llm"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/conversations/dtos/callbacks"
	"engflex-api/internal/modules/conversations/dtos/requests"
	"engflex-api/internal/modules/conversations/dtos/responses"
	"engflex-api/internal/modules/conversations/prompts"
	"engflex-api/internal/modules/conversations/schemas"
	"engflex-api/internal/utils"
	"engflex-api/internal/voice"
)

// maxSummaryTranscriptChars bounds the LLM input: user turns only, truncated.
const maxSummaryTranscriptChars = 4000

var errEmptyAnswer = errors.New("voice engine returned an empty offer answer")

// ConversationService owns the conversation lifecycle rules.
type ConversationService struct {
	conversations repositories.ConversationRepository
	turns         repositories.ConversationTurnRepository
	voice         voice.VoiceClient
	pronounce     voice.PronounceEngine
	llm           llm.LLMClient
	maxDuration   int
	feedbacks     repositories.FeedbackRepository
	scenarios     repositories.ScenarioRepository
	personas      repositories.PersonaRepository
}

// NewConversationService builds the service over repository + engine interfaces.
func NewConversationService(conversations repositories.ConversationRepository, turns repositories.ConversationTurnRepository, voice voice.VoiceClient, pronounce voice.PronounceEngine, llm llm.LLMClient, maxDuration int, feedbacks repositories.FeedbackRepository, scenarios repositories.ScenarioRepository, personas repositories.PersonaRepository) *ConversationService {
	return &ConversationService{conversations: conversations, turns: turns, voice: voice, pronounce: pronounce, llm: llm, maxDuration: maxDuration, feedbacks: feedbacks, scenarios: scenarios, personas: personas}
}

// Create provisions a pending conversation. The close of the caller's other
// non-terminal conversations and the insert run atomically inside the
// repository (CreateClosingActive), so concurrent creates cannot both slip
// through — the row lock serializes them.
func (s *ConversationService) Create(ctx context.Context, userID string, req requests.StartConversation) (*models.Conversation, error) {
	if req.Mode == enums.ConversationModeRoleplay && (req.ScenarioID == nil || *req.ScenarioID == "") {
		return nil, common.BadRequest("scenarioId is required for roleplay")
	}

	m := &models.Conversation{}
	if err := utils.Map(m, req); err != nil {
		logger.Report(ctx, "map start conversation failed", err, "userID", userID)
		return nil, common.Internal()
	}
	m.UserID = userID
	m.Status = enums.ConversationStatusPending
	if req.Mode == enums.ConversationModeRoleplay && req.ScenarioID != nil && *req.ScenarioID != "" {
		if snap, ok := s.objectivesSnapshot(ctx, *req.ScenarioID); ok {
			m.Objectives = datatypes.JSON(snap)
		}
	}
	if err := s.conversations.CreateClosingActive(ctx, m); err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "create conversation failed", appErr, "userID", userID)
		return nil, appErr
	}
	slog.InfoContext(ctx, "conversation created", "conversationID", m.ID, "mode", m.Mode)
	return m, nil
}

// objectivesSnapshot copies the scenario's objectives at Create time, so a
// later scenario edit cannot rewrite what the learner actually practiced.
func (s *ConversationService) objectivesSnapshot(ctx context.Context, scenarioID string) ([]byte, bool) {
	scenario, err := s.scenarios.GetByID(ctx, scenarioID)
	if err != nil || scenario == nil || len(scenario.Objectives) == 0 {
		return nil, false
	}
	raw := make([]byte, len(scenario.Objectives))
	copy(raw, scenario.Objectives)
	return raw, true
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
func (s *ConversationService) Start(ctx context.Context, userID, id string) (*models.Conversation, *responses.IceConfig, error) {
	conv, err := s.getOwned(ctx, userID, id)
	if err != nil {
		return nil, nil, err
	}
	if conv.Status != enums.ConversationStatusPending {
		return nil, nil, common.Conflict("conversation is not pending")
	}
	if conv.SpeechSessionID != "" && len(conv.SpeechStartResponse) > 0 {
		if cached, ok := decodeStartResponse(conv.SpeechStartResponse); ok {
			return conv, cached.ICEConfig, nil
		}
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
		refreshed, err := s.conversations.GetByID(ctx, conv.ID)
		if err != nil {
			appErr := common.FromDBError(err, "conversation")
			logger.Report(ctx, "recheck conversation start failed", appErr, "conversationID", conv.ID)
			return nil, nil, appErr
		}
		if cached, ok := decodeStartResponse(refreshed.SpeechStartResponse); ok {
			return refreshed, cached.ICEConfig, nil
		}
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

// Finalize records the engine-reported duration (internal callback).
// It is intentionally status-agnostic: the engine reports sessions it ran,
// which are live, or already ended by the user — either way the duration
// must land (first ended_at wins, first duration wins, per SetEnded).
func (s *ConversationService) Finalize(ctx context.Context, id string, durationSec int) error {
	if durationSec < 0 {
		return common.BadRequest("durationSec must be >= 0")
	}
	if _, err := s.conversations.GetByID(ctx, id); err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "finalize conversation lookup failed", appErr, "conversationID", id)
		return appErr
	}
	// Zero turns at finalize is definitive: the engine posts its turn batch
	// (awaited) before this callback, so nothing can arrive later. Discard
	// the empty conversation instead of storing empty history.
	deleted, err := s.conversations.DeleteIfEmpty(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "finalize empty-check failed", appErr, "conversationID", id)
		return appErr
	}
	if deleted {
		slog.InfoContext(ctx, "empty conversation discarded", "conversationID", id)
		return nil
	}
	if _, err := s.conversations.SetEnded(ctx, id, &durationSec); err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "finalize conversation failed", appErr, "conversationID", id)
		return appErr
	}
	slog.InfoContext(ctx, "conversation finalized", "conversationID", id, "durationSec", durationSec)
	return nil
}

// IngestTurns stores the engine's finalize-time batch. Positions are made
// idempotent by the repository's ON CONFLICT DO NOTHING upsert, so a retried
// batch cannot duplicate turns.
func (s *ConversationService) IngestTurns(ctx context.Context, id string, req callbacks.IngestTurns) (int, error) {
	if len(req.Turns) == 0 {
		return 0, common.BadRequest("turns must not be empty")
	}
	if len(req.Turns) > common.MaxTurnsPerBatch {
		return 0, common.BadRequest("too many turns in one batch")
	}
	if _, err := s.conversations.GetByID(ctx, id); err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "ingest turns: conversation lookup failed", appErr, "conversationID", id)
		return 0, appErr
	}

	var turns []*models.ConversationTurn
	if err := utils.MapSlice(&turns, req.Turns); err != nil {
		logger.Report(ctx, "map ingest turns failed", err, "conversationID", id)
		return 0, common.Internal()
	}
	for _, turn := range turns {
		turn.ConversationID = id
	}

	stored, err := s.turns.UpsertTurns(ctx, id, turns)
	if err != nil {
		appErr := common.FromDBError(err, "conversation turn")
		logger.Report(ctx, "ingest turns failed", appErr, "conversationID", id)
		return 0, appErr
	}
	slog.InfoContext(ctx, "turns ingested", "conversationID", id, "received", len(turns), "stored", stored)
	return stored, nil
}

// Summarize generates the post-session summary once, on first view, and
// stores it. Idempotent: a stored summary is returned without an LLM call.
// Live sessions are rejected (409); sessions with no learner turns get 422.
func (s *ConversationService) Summarize(ctx context.Context, userID, id string) (*models.ConversationSummary, error) {
	conv, err := s.getOwned(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if len(conv.Summary) > 0 && string(conv.Summary) != "null" && string(conv.Summary) != "{}" {
		var existing models.ConversationSummary
		if err := json.Unmarshal(conv.Summary, &existing); err == nil {
			return &existing, nil
		}
	}
	if conv.Status != enums.ConversationStatusEnded && conv.Status != enums.ConversationStatusFailed {
		return nil, common.Conflict("conversation is not finished")
	}
	turns, err := s.turns.ListTurnsWithFeedback(ctx, id)
	if err != nil {
		appErr := common.FromDBError(err, "conversation turn")
		logger.Report(ctx, "summarize turns load failed", appErr, "conversationID", id)
		return nil, appErr
	}
	userTurns := filterUserTurns(turns)
	if len(userTurns) == 0 {
		return nil, common.UnprocessableEntity("no learner turns to summarize")
	}
	var objs []models.ScenarioObjective
	_ = json.Unmarshal(conv.Objectives, &objs)
	transcript := joinUserTurns(userTurns, maxSummaryTranscriptChars)
	prompt := prompts.BuildSessionSummaryPrompt(objs, transcript)
	var out schemas.SessionSummaryLLM
	if err := s.llm.CompleteStructured(ctx, prompt, &out, schemas.SessionSummaryName, schemas.SessionSummarySchema); err != nil {
		logger.Report(ctx, "session summary failed", err, "conversationID", id)
		return nil, common.ServiceUnavailable("summary unavailable")
	}
	summary, err := schemas.NormalizeSessionSummary(out, objs, transcript)
	if err != nil {
		logger.Report(ctx, "session summary rejected", err, "conversationID", id)
		return nil, common.ServiceUnavailable("summary unavailable")
	}
	raw, err := json.Marshal(summary)
	if err != nil {
		logger.Report(ctx, "encode summary failed", common.Internal(), "conversationID", id)
		return nil, common.Internal()
	}
	conv.Summary = raw
	if err := s.conversations.Update(ctx, conv); err != nil {
		appErr := common.FromDBError(err, "conversation")
		logger.Report(ctx, "store summary failed", appErr, "conversationID", id)
		return nil, appErr
	}
	slog.InfoContext(ctx, "session summarized", "conversationID", id, "turns", len(userTurns))
	return summary, nil
}

// filterUserTurns keeps learner turns with text, in position order.
func filterUserTurns(turns []*models.ConversationTurn) []*models.ConversationTurn {
	var out []*models.ConversationTurn
	for _, t := range turns {
		if t.Role == enums.TurnRoleUser && strings.TrimSpace(t.Text) != "" {
			out = append(out, t)
		}
	}
	return out
}

// joinUserTurns concatenates learner turns, truncated at maxChars.
func joinUserTurns(turns []*models.ConversationTurn, maxChars int) string {
	var b strings.Builder
	for _, t := range turns {
		text := strings.TrimSpace(t.Text)
		if text == "" {
			continue
		}
		if b.Len()+len(text)+1 > maxChars {
			break
		}
		b.WriteString(text)
		b.WriteString("\n")
	}
	return b.String()
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
	var objs []models.ScenarioObjective
	if len(conv.Objectives) > 0 {
		_ = json.Unmarshal(conv.Objectives, &objs)
	} else {
		_ = json.Unmarshal(scenario.Objectives, &objs)
	}
	for _, o := range objs {
		scenarioBody.Objectives = append(scenarioBody.Objectives, voice.VoiceObjectiveBody{ID: o.ID, Label: o.Label, Hint: o.Hint})
	}
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
