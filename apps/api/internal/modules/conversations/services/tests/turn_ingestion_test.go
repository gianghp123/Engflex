package tests

import (
	"context"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/conversations/dtos/callbacks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestIngestSnapshot(t *testing.T) {
	for _, tt := range []struct {
		name   string
		req    callbacks.IngestTurns
		status int
	}{
		{"empty acknowledged", callbacks.IngestTurns{EngineSessionID: "engine-1", Revision: 1, Turns: []callbacks.IngestTurn{}}, 0},
		{"turn", callbacks.IngestTurns{EngineSessionID: "engine-1", Revision: 2, Turns: []callbacks.IngestTurn{{Position: 1, Role: enums.TurnRoleUser, Text: "hello"}}}, 0},
		{"missing session", callbacks.IngestTurns{Revision: 1, Turns: []callbacks.IngestTurn{}}, http.StatusBadRequest},
		{"missing revision", callbacks.IngestTurns{EngineSessionID: "engine-1", Turns: []callbacks.IngestTurn{}}, http.StatusBadRequest},
		{"missing turns", callbacks.IngestTurns{EngineSessionID: "engine-1", Revision: 1}, http.StatusBadRequest},
		{"position gap", callbacks.IngestTurns{EngineSessionID: "engine-1", Revision: 1, Turns: []callbacks.IngestTurn{{Position: 2, Role: enums.TurnRoleUser, Text: "hello"}}}, http.StatusBadRequest},
	} {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, _, _ := newService(t)
			if tt.status == 0 {
				repo.On("ApplySnapshot", mock.Anything, convID, tt.req.EngineSessionID, tt.req.Revision, mock.Anything, mock.Anything, (*int)(nil)).Return(&models.Conversation{ID: convID}, nil).Once()
			}
			n, err := svc.IngestTurns(context.Background(), convID, tt.req)
			if tt.status != 0 {
				requireAppError(t, err, tt.status)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tt.req.Turns), n)
			}
		})
	}
	t.Run("repository conflict stays 409", func(t *testing.T) {
		svc, repo, _, _, _ := newService(t)
		repo.On("ApplySnapshot", mock.Anything, convID, "engine-1", int64(1), mock.Anything, mock.Anything, (*int)(nil)).Return(nil, common.Conflict()).Once()
		_, err := svc.IngestTurns(context.Background(), convID, callbacks.IngestTurns{EngineSessionID: "engine-1", Revision: 1, Turns: []callbacks.IngestTurn{}})
		requireAppError(t, err, http.StatusConflict)
	})
}
