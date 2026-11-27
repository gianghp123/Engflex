package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
	llmmocks "engflex-api/internal/llm/mocks"
	"engflex-api/internal/modules/conversations/dtos/requests"
	"engflex-api/internal/modules/conversations/services"
	voicemocks "engflex-api/internal/voice/mocks"
)

// fakeNotPremium satisfies common.PremiumChecker as "not premium" so the
// gating branch under test is deterministic.
type fakeNotPremium struct{}

func (fakeNotPremium) IsPremium(ctx context.Context, userID string) (bool, error) {
	return false, nil
}

func TestCreateGatingFollowsFlag(t *testing.T) {
	scenarioID := "22222222-2222-2222-2222-222222222222"
	roleplay := func() requests.StartConversation {
		return requests.StartConversation{Mode: enums.ConversationModeRoleplay, ScenarioID: &scenarioID}
	}
	tests := []struct {
		name     string
		gatingOn bool
		req      requests.StartConversation
		scenario *models.Scenario // nil = the scenario must never load
		loads    int              // GetByID calls: gating probe + objectives snapshot
		wantErr  int              // 0 = nil
	}{
		{name: "gating off still loads the scenario for the objectives snapshot", gatingOn: false, req: roleplay(), scenario: &models.Scenario{ID: scenarioID}, loads: 1},
		{name: "gating on blocks a premium scenario for a free user", gatingOn: true, req: roleplay(), scenario: &models.Scenario{ID: scenarioID, IsPremium: true}, loads: 1, wantErr: 403},
		{name: "gating on passes free scenarios", gatingOn: true, req: roleplay(), scenario: &models.Scenario{ID: scenarioID, IsPremium: false}, loads: 2},
		{name: "free talk never loads the scenario", gatingOn: true, req: requests.StartConversation{Mode: enums.ConversationModeFreeTalk}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			common.SetContentGating(tt.gatingOn)
			t.Cleanup(func() { common.SetContentGating(false) })

			scenarios := repomocks.NewMockScenarioRepository(t)
			if tt.scenario != nil {
				scenarios.EXPECT().GetByID(mock.Anything, scenarioID).Return(tt.scenario, nil).Times(tt.loads)
			}
			convos := repomocks.NewMockConversationRepository(t)
			if tt.wantErr == 0 {
				convos.EXPECT().CreateClosingActive(mock.Anything, mock.Anything).Return(nil).Once()
			}
			svc := services.NewConversationService(
				convos,
				repomocks.NewMockConversationTurnRepository(t),
				voicemocks.NewMockVoiceClient(t),
				voicemocks.NewMockPronounceEngine(t),
				llmmocks.NewMockLLMClient(t),
				300,
				repomocks.NewMockFeedbackRepository(t),
				scenarios,
				repomocks.NewMockPersonaRepository(t),
				fakeNotPremium{},
				repomocks.NewMockUsageEventRepository(t),
				shadowAccessBundle(t),
			)

			m, err := svc.Create(context.Background(), userID, tt.req)
			if tt.wantErr == 0 {
				require.NoError(t, err)
				require.NotNil(t, m)
			} else {
				requireAppError(t, err, tt.wantErr)
				require.Nil(t, m)
			}
		})
	}
}
