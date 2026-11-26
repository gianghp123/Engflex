// Voice charging contract (spec §8.1): cached ICE replays free, engine
// failure releases, the winner of the SetSpeechStart race commits and the
// loser releases, shadow mode never reserves, and AnalyzeTurn requires a
// committed (or grandfathered) session when enforcement is on.
package tests

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	accessservices "engflex-api/internal/modules/access/services"
	accessmocks "engflex-api/internal/modules/access/services/mocks"
	"engflex-api/internal/modules/conversations/dtos/responses"
	"engflex-api/internal/modules/conversations/services"
	"engflex-api/internal/voice"
	voicemocks "engflex-api/internal/voice/mocks"
)

// fakeVoice is a scriptable VoiceClient. It embeds the interface so only
// the methods a test exercises need overriding; any other engine call
// panics on the nil embedded value — an unexpected call fails loudly.
type fakeVoice struct {
	voice.VoiceClient
	start      *voice.StartResponse
	startErr   error
	analyze    *voice.AnalyzeTurnResponse
	analyzeErr error
}

func (f fakeVoice) Start(ctx context.Context, req voice.StartRequest) (*voice.StartResponse, error) {
	return f.start, f.startErr
}

func (f fakeVoice) AnalyzeTurn(ctx context.Context, req voice.AnalyzeTurnRequest) (*voice.AnalyzeTurnResponse, error) {
	return f.analyze, f.analyzeErr
}

func pendingConversation() *models.Conversation {
	return &models.Conversation{
		ID:     "00000000-0000-0000-0000-000000000101",
		UserID: "user_1",
		Status: enums.ConversationStatusPending,
	}
}

func cachedStartResponse() datatypes.JSON {
	return datatypes.JSON([]byte(`{"sessionId":"engine-1","iceConfig":{"iceServers":[]}}`))
}

func engineStartResponse(sessionID string) *voice.StartResponse {
	return &voice.StartResponse{
		SessionID: sessionID,
		ICEConfig: &responses.IceConfig{IceServers: []responses.IceServer{{URLs: []string{"stun:x"}}}},
	}
}

func allowVoiceDecision() accessservices.Decision {
	return accessservices.Decision{Allow: true, Plan: enums.PlanKeyFree,
		Charges: []accessservices.Charge{
			{Counter: enums.CounterKindAICredit, Amount: 1},
			{Counter: enums.CounterKindVoiceSession, Amount: 1},
		}}
}

func newAccessMocks(t *testing.T, enforce bool) (*accessmocks.MockPolicyDecider, *accessmocks.MockUsageMeter, *services.AccessDependencies) {
	t.Helper()
	policy := accessmocks.NewMockPolicyDecider(t)
	usage := accessmocks.NewMockUsageMeter(t)
	return policy, usage, &services.AccessDependencies{Policy: policy, Usage: usage, Enforce: enforce}
}

func expectVoicePrincipal(t *testing.T, policy *accessmocks.MockPolicyDecider) {
	t.Helper()
	policy.EXPECT().Principal(mock.Anything, "user_1", false).
		Return(accessservices.Principal{UserID: "user_1", Plan: enums.PlanKeyFree}, accessservices.TrialInfo{}, nil).Once()
}

// shadowAccessBundle wires the access mocks in shadow mode for the
// pre-access tests: Decide/Principal answer allow, nothing is enforced or
// metered. Maybe() keeps the stubs optional so tests that never start a
// session still pass AssertExpectations.
func shadowAccessBundle(t *testing.T) *services.AccessDependencies {
	t.Helper()
	policy := accessmocks.NewMockPolicyDecider(t)
	usage := accessmocks.NewMockUsageMeter(t)
	policy.EXPECT().Principal(mock.Anything, mock.Anything, mock.Anything).
		Return(accessservices.Principal{Plan: enums.PlanKeyFree}, accessservices.TrialInfo{}, nil).Maybe()
	policy.EXPECT().Decide(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(accessservices.Decision{Allow: true, Plan: enums.PlanKeyFree}, nil).Maybe()
	return &services.AccessDependencies{Policy: policy, Usage: usage}
}

func TestStartChargesVoiceSession(t *testing.T) {
	policy, usage, deps := newAccessMocks(t, true)
	conv := pendingConversation()

	expectVoicePrincipal(t, policy)
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionVoiceStart).
		Return(allowVoiceDecision(), nil).Once()
	usage.EXPECT().Reserve(mock.Anything, "user_1", enums.PlanKeyFree,
		[]accessservices.Charge{
			{Counter: enums.CounterKindAICredit, Amount: 1},
			{Counter: enums.CounterKindVoiceSession, Amount: 1},
		},
		"conversation:00000000-0000-0000-0000-000000000101", enums.UsageRefTypeConversation,
		"00000000-0000-0000-0000-000000000101", common.VoiceLeaseTTL, mock.Anything, mock.Anything).
		Return(&accessservices.Reservation{ID: "res-1", Status: accessservices.ReservationStatusReserved}, nil).Once()
	usage.EXPECT().Commit(mock.Anything, "res-1").Return(nil).Once()

	conversations := repomocks.NewMockConversationRepository(t)
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(conv, nil).Once()
	conversations.EXPECT().SetSpeechStart(mock.Anything, "00000000-0000-0000-0000-000000000101", "engine-1", mock.Anything).Return(int64(1), nil).Once()

	svc := services.NewConversationServiceForTest(conversations, repomocks.NewMockUsageEventRepository(t), fakeVoice{
		start: engineStartResponse("engine-1"),
	}, deps)
	got, ice, err := svc.Start(context.Background(), "user_1", "00000000-0000-0000-0000-000000000101")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.NotNil(t, ice)
}

func TestStartCachedNeverCharges(t *testing.T) {
	_, _, deps := newAccessMocks(t, true)
	conv := pendingConversation()
	conv.SpeechSessionID = "engine-1"
	conv.SpeechStartResponse = cachedStartResponse()

	conversations := repomocks.NewMockConversationRepository(t)
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(conv, nil).Once()

	svc := services.NewConversationServiceForTest(conversations, repomocks.NewMockUsageEventRepository(t), fakeVoice{}, deps)
	_, ice, err := svc.Start(context.Background(), "user_1", "00000000-0000-0000-0000-000000000101")
	require.NoError(t, err)
	require.NotNil(t, ice, "cached ICE replays without touching the access services")
}

func TestStartDeniedByPolicy(t *testing.T) {
	policy, _, deps := newAccessMocks(t, true)
	conv := pendingConversation()
	expectVoicePrincipal(t, policy)
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionVoiceStart).
		Return(accessservices.Decision{Allow: false, Code: common.CodeInsufficientCredits, Plan: enums.PlanKeyFree}, nil).Once()

	conversations := repomocks.NewMockConversationRepository(t)
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(conv, nil).Once()

	svc := services.NewConversationServiceForTest(conversations, repomocks.NewMockUsageEventRepository(t), fakeVoice{}, deps)
	_, _, err := svc.Start(context.Background(), "user_1", "00000000-0000-0000-0000-000000000101")
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 429, appErr.Status)
	require.Equal(t, common.CodeInsufficientCredits, appErr.Code)
}

func TestStartEngineFailureReleases(t *testing.T) {
	policy, usage, deps := newAccessMocks(t, true)
	conv := pendingConversation()
	expectVoicePrincipal(t, policy)
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionVoiceStart).
		Return(accessservices.Decision{Allow: true, Plan: enums.PlanKeyFree,
			Charges: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}}}, nil).Once()
	usage.EXPECT().Reserve(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&accessservices.Reservation{ID: "res-9", Status: accessservices.ReservationStatusReserved}, nil).Once()
	usage.EXPECT().Release(mock.Anything, "res-9").Return(nil).Once()

	conversations := repomocks.NewMockConversationRepository(t)
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(conv, nil).Once()

	svc := services.NewConversationServiceForTest(conversations, repomocks.NewMockUsageEventRepository(t), fakeVoice{startErr: errors.New("boom")}, deps)
	_, _, err := svc.Start(context.Background(), "user_1", "00000000-0000-0000-0000-000000000101")
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 503, appErr.Status, "engine never started: released, not charged")
}

// TestStartCacheWriteFailureReleases pins the pre-commit guarantee: when
// the start row cannot be cached, the hold is released and the request
// fails — the user is never charged for a session that was not recorded.
func TestStartCacheWriteFailureReleases(t *testing.T) {
	policy, usage, deps := newAccessMocks(t, true)
	conv := pendingConversation()
	expectVoicePrincipal(t, policy)
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionVoiceStart).
		Return(allowVoiceDecision(), nil).Once()
	usage.EXPECT().Reserve(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&accessservices.Reservation{ID: "res-cache", Status: accessservices.ReservationStatusReserved}, nil).Once()
	usage.EXPECT().Release(mock.Anything, "res-cache").Return(nil).Once()
	// No Commit expectation: mockery fails the test if Commit is called.

	conversations := repomocks.NewMockConversationRepository(t)
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(conv, nil).Once()
	conversations.EXPECT().SetSpeechStart(mock.Anything, "00000000-0000-0000-0000-000000000101", "engine-1", mock.Anything).
		Return(int64(0), errors.New("boom")).Once()

	svc := services.NewConversationServiceForTest(conversations, repomocks.NewMockUsageEventRepository(t), fakeVoice{
		start: engineStartResponse("engine-1"),
	}, deps)
	_, _, err := svc.Start(context.Background(), "user_1", "00000000-0000-0000-0000-000000000101")
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 500, appErr.Status)
}

// TestStartCommitFailureStillSucceeds pins fail-open settlement: the engine
// session is live and cached, so a Commit error must not fail the request
// (the sweeper owns the unsettled hold) and must not release it either.
func TestStartCommitFailureStillSucceeds(t *testing.T) {
	policy, usage, deps := newAccessMocks(t, true)
	conv := pendingConversation()
	expectVoicePrincipal(t, policy)
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionVoiceStart).
		Return(allowVoiceDecision(), nil).Once()
	usage.EXPECT().Reserve(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&accessservices.Reservation{ID: "res-commit", Status: accessservices.ReservationStatusReserved}, nil).Once()
	usage.EXPECT().Commit(mock.Anything, "res-commit").Return(errors.New("boom")).Once()
	// No Release expectation: mockery fails the test if Release is called.

	conversations := repomocks.NewMockConversationRepository(t)
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(conv, nil).Once()
	conversations.EXPECT().SetSpeechStart(mock.Anything, "00000000-0000-0000-0000-000000000101", "engine-9", mock.Anything).Return(int64(1), nil).Once()

	svc := services.NewConversationServiceForTest(conversations, repomocks.NewMockUsageEventRepository(t), fakeVoice{
		start: engineStartResponse("engine-9"),
	}, deps)
	got, ice, err := svc.Start(context.Background(), "user_1", "00000000-0000-0000-0000-000000000101")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "engine-9", got.SpeechSessionID)
	require.NotNil(t, ice)
}

func TestStartRaceLoserReleasesReservation(t *testing.T) {
	policy, usage, deps := newAccessMocks(t, true)
	conv := pendingConversation()
	expectVoicePrincipal(t, policy)
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionVoiceStart).
		Return(accessservices.Decision{Allow: true, Plan: enums.PlanKeyFree,
			Charges: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}}}, nil).Once()
	usage.EXPECT().Reserve(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&accessservices.Reservation{ID: "res-race", Status: accessservices.ReservationStatusReserved}, nil).Once()
	usage.EXPECT().Release(mock.Anything, "res-race").Return(nil).Once()

	raced := pendingConversation()
	raced.SpeechSessionID = "engine-winner"
	raced.SpeechStartResponse = cachedStartResponse()
	conversations := repomocks.NewMockConversationRepository(t)
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(conv, nil).Once()
	conversations.EXPECT().SetSpeechStart(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(0), nil).Once()
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(raced, nil).Once()

	svc := services.NewConversationServiceForTest(conversations, repomocks.NewMockUsageEventRepository(t), fakeVoice{
		start: engineStartResponse("engine-loser"),
	}, deps)
	_, ice, err := svc.Start(context.Background(), "user_1", "00000000-0000-0000-0000-000000000101")
	require.NoError(t, err, "the loser replays the winner's cached ICE")
	require.NotNil(t, ice)
}

func TestStartInProgressReplayReturns409(t *testing.T) {
	policy, usage, deps := newAccessMocks(t, true)
	conv := pendingConversation()
	expectVoicePrincipal(t, policy)
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionVoiceStart).
		Return(accessservices.Decision{Allow: true, Plan: enums.PlanKeyFree,
			Charges: []accessservices.Charge{{Counter: enums.CounterKindAICredit, Amount: 1}}}, nil).Once()
	usage.EXPECT().Reserve(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, common.Conflict("a request for this item is already in progress").
			WithCode("in_progress")).Once()

	conversations := repomocks.NewMockConversationRepository(t)
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(conv, nil).Once()

	svc := services.NewConversationServiceForTest(conversations, repomocks.NewMockUsageEventRepository(t), fakeVoice{}, deps)
	_, _, err := svc.Start(context.Background(), "user_1", "00000000-0000-0000-0000-000000000101")
	appErr := &common.AppError{}
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, 409, appErr.Status)
	require.Equal(t, "in_progress", appErr.Code)
}

// TestStartCommittedReplayReplaysCache pins the Reserve-contract branch the
// brief sketch omits: an all-committed replay returns (Committed, nil), so
// Start must re-read and replay the cached ICE without calling the engine.
func TestStartCommittedReplayReplaysCache(t *testing.T) {
	policy, usage, deps := newAccessMocks(t, true)
	conv := pendingConversation()
	expectVoicePrincipal(t, policy)
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionVoiceStart).
		Return(allowVoiceDecision(), nil).Once()
	usage.EXPECT().Reserve(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&accessservices.Reservation{ID: "res-old", Status: accessservices.ReservationStatusCommitted}, nil).Once()

	winner := pendingConversation()
	winner.SpeechSessionID = "engine-1"
	winner.SpeechStartResponse = cachedStartResponse()
	conversations := repomocks.NewMockConversationRepository(t)
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(conv, nil).Once()
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(winner, nil).Once()

	svc := services.NewConversationServiceForTest(conversations, repomocks.NewMockUsageEventRepository(t), fakeVoice{
		startErr: errors.New("engine must not be called on a committed replay"),
	}, deps)
	got, ice, err := svc.Start(context.Background(), "user_1", "00000000-0000-0000-0000-000000000101")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "engine-1", got.SpeechSessionID)
	require.NotNil(t, ice)
}

func TestStartShadowModeSkipsMetering(t *testing.T) {
	policy, _, deps := newAccessMocks(t, false)
	conv := pendingConversation()
	// Shadow mode still resolves the principal and evaluates the decision
	// for logging; it just never reserves or commits.
	expectVoicePrincipal(t, policy)
	policy.EXPECT().Decide(mock.Anything, "user_1", false, accessservices.ActionVoiceStart).
		Return(accessservices.Decision{Allow: true, Plan: enums.PlanKeyFree}, nil).Once()

	conversations := repomocks.NewMockConversationRepository(t)
	conversations.EXPECT().GetByID(mock.Anything, "00000000-0000-0000-0000-000000000101").Return(conv, nil).Once()
	conversations.EXPECT().SetSpeechStart(mock.Anything, "00000000-0000-0000-0000-000000000101", "engine-1", mock.Anything).Return(int64(1), nil).Once()

	svc := services.NewConversationServiceForTest(conversations, repomocks.NewMockUsageEventRepository(t), fakeVoice{
		start: engineStartResponse("engine-1"),
	}, deps)
	_, _, err := svc.Start(context.Background(), "user_1", "00000000-0000-0000-0000-000000000101")
	require.NoError(t, err, "shadow mode: decide + log, no reserve/commit")
}

func voiceChargingTurn() *models.ConversationTurn {
	return &models.ConversationTurn{
		ID: "33333333-3333-3333-3333-333333333333", ConversationID: "00000000-0000-0000-0000-000000000101",
		Position: 1, Role: enums.TurnRoleUser, Text: "I go yesterday.",
	}
}

// TestAnalyzeTurnEnforcement pins the AnalyzeTurn guard: a committed event
// (or a grandfathered speech session) analyzes; an unpaid conversation is
// 403; a store failure is 503; shadow mode skips the check.
func TestAnalyzeTurnEnforcement(t *testing.T) {
	const convID = "00000000-0000-0000-0000-000000000101"
	good := `{"corrected":"I went yesterday.","spans":[],"relevance":{"status":"relevant","reason":null},"alternatives":{"language":null,"contextual":null},"tip":"past tense"}`

	tests := []struct {
		name        string
		enforce     bool
		speech      string // SpeechSessionID on the owned conversation
		committed   bool
		commitErr   error
		wantStatus  int
		wantCode    string
		expectUsage bool
	}{
		{name: "committed session analyzes", enforce: true, committed: true, expectUsage: true},
		{name: "grandfathered speech session analyzes without an event", enforce: true, speech: "engine-1", expectUsage: true},
		{name: "unpaid conversation is 403", enforce: true, wantStatus: 403, wantCode: common.CodePremiumRequired, expectUsage: true},
		{name: "usage check failure is 503", enforce: true, commitErr: errors.New("boom"), wantStatus: 503, wantCode: common.CodePolicyUnavailable, expectUsage: true},
		{name: "shadow mode skips the check", enforce: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := accessmocks.NewMockPolicyDecider(t)
			usage := accessmocks.NewMockUsageMeter(t)
			deps := &services.AccessDependencies{Policy: policy, Usage: usage, Enforce: tt.enforce}

			conversations := repomocks.NewMockConversationRepository(t)
			conversations.EXPECT().GetByID(mock.Anything, convID).Return(&models.Conversation{
				ID: convID, UserID: "user_1", Status: enums.ConversationStatusPending,
				SpeechSessionID: tt.speech,
			}, nil).Once()

			usageEvents := repomocks.NewMockUsageEventRepository(t)
			if tt.expectUsage {
				usageEvents.EXPECT().HasCommitted(mock.Anything, "user_1", []string{
					"conversation:" + convID + ":ai_credit",
					"conversation:" + convID + ":voice_session",
				}).Return(tt.committed, tt.commitErr).Once()
			}

			turns := repomocks.NewMockConversationTurnRepository(t)
			feedbacks := repomocks.NewMockFeedbackRepository(t)
			var engine fakeVoice
			if tt.wantStatus == 0 {
				turns.EXPECT().GetTurnByPosition(mock.Anything, convID, 1).Return(voiceChargingTurn(), nil).Once()
				turns.EXPECT().ListTurns(mock.Anything, convID).Return([]*models.ConversationTurn{voiceChargingTurn()}, nil).Once()
				engine = fakeVoice{analyze: &voice.AnalyzeTurnResponse{Feedback: json.RawMessage(good)}}
				feedbacks.EXPECT().Upsert(mock.Anything, mock.Anything).Return(nil).Once()
			}

			svc := services.NewConversationService(conversations, turns, engine,
				voicemocks.NewMockPronounceEngine(t), 300, feedbacks,
				repomocks.NewMockScenarioRepository(t), repomocks.NewMockPersonaRepository(t),
				fakeNotPremium{}, usageEvents, deps)

			got, err := svc.AnalyzeTurn(context.Background(), "user_1", convID, 1, "")
			if tt.wantStatus != 0 {
				appErr := &common.AppError{}
				require.ErrorAs(t, err, &appErr)
				require.Equal(t, tt.wantStatus, appErr.Status)
				require.Equal(t, tt.wantCode, appErr.Code)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
