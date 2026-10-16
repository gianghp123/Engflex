package tests

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	repomocks "engflex-api/internal/database/repositories/mocks"
)

func TestListConversations(t *testing.T) {
	scenarioID := "22222222-2222-2222-2222-222222222222"
	ended := &models.Conversation{
		ID: convID, UserID: userID,
		Mode: enums.ConversationModeRoleplay, Status: enums.ConversationStatusEnded,
		ScenarioID: &scenarioID,
	}
	free := &models.Conversation{
		ID: "33333333-3333-3333-3333-333333333333", UserID: userID,
		Mode: enums.ConversationModeFreeTalk, Status: enums.ConversationStatusEnded,
	}

	tests := []struct {
		name       string
		scenarioID *string
		setup      func(*repomocks.MockConversationRepository)
		wantLen    int
		wantTotal  int64
		wantStatus int // 0 = success
	}{
		{
			name: "lists the caller's finished conversations with the total",
			setup: func(repo *repomocks.MockConversationRepository) {
				repo.On("ListByUser", mock.Anything, userID, (*string)(nil), 20, 0).
					Return([]*models.Conversation{ended, free}, nil).Once()
				repo.On("CountByUser", mock.Anything, userID, (*string)(nil)).
					Return(int64(2), nil).Once()
			},
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name:       "passes the scenario filter through unchanged",
			scenarioID: &scenarioID,
			setup: func(repo *repomocks.MockConversationRepository) {
				repo.On("ListByUser", mock.Anything, userID, &scenarioID, 20, 0).
					Return([]*models.Conversation{ended}, nil).Once()
				repo.On("CountByUser", mock.Anything, userID, &scenarioID).
					Return(int64(1), nil).Once()
			},
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name: "list failure is a hidden 500",
			setup: func(repo *repomocks.MockConversationRepository) {
				repo.On("ListByUser", mock.Anything, userID, (*string)(nil), 20, 0).
					Return(nil, errors.New("boom")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "count failure is a hidden 500",
			setup: func(repo *repomocks.MockConversationRepository) {
				repo.On("ListByUser", mock.Anything, userID, (*string)(nil), 20, 0).
					Return(nil, nil).Once()
				repo.On("CountByUser", mock.Anything, userID, (*string)(nil)).
					Return(int64(0), errors.New("boom")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, _, _ := newService(t)
			tt.setup(repo)

			got, total, err := svc.List(context.Background(), userID, tt.scenarioID, 20, 0)
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
			require.Len(t, got, tt.wantLen)
			require.Equal(t, tt.wantTotal, total)
		})
	}
}
