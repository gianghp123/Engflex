package tests

import (
	"context"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"net/http"
	"os"
	"testing"
)

func TestPostgresConversationSnapshots(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to migrated PostgreSQL")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlog.Discard})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	defer pool.Close()
	conv := &models.Conversation{UserID: "snapshot-" + uuid.NewString(), Mode: enums.ConversationModeFreeTalk, Status: enums.ConversationStatusLive, SpeechSessionID: "engine-test"}
	require.NoError(t, db.Create(conv).Error)
	defer db.Delete(conv)
	ctx := context.Background()
	repo := repositories.NewConversationRepository(db)
	turns := []*models.ConversationTurn{{ConversationID: conv.ID, Position: 1, Role: enums.TurnRoleUser, Text: "wrong"}, {ConversationID: conv.ID, Position: 2, Role: enums.TurnRoleAI, Text: "stale"}}
	_, err = repo.ApplySnapshot(ctx, conv.ID, "engine-test", 1, "original", turns, nil)
	require.NoError(t, err)
	rows, err := repositories.NewConversationTurnRepository(db).ListTurns(ctx, conv.ID)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	f := &models.Feedback{UserID: conv.UserID, SubjectType: enums.FeedbackSubjectConversationTurn, SubjectID: rows[0].ID, InputFingerprint: "input"}
	require.NoError(t, repo.StoreAnalysis(ctx, conv.ID, 1, f))
	defer db.Where("user_id = ?", conv.UserID).Delete(&models.Feedback{})
	corrected := []*models.ConversationTurn{{ConversationID: conv.ID, Position: 1, Role: enums.TurnRoleUser, Text: "correct"}}
	_, err = repo.ApplySnapshot(ctx, conv.ID, "engine-test", 2, "corrected", corrected, nil)
	require.NoError(t, err)
	rows, err = repositories.NewConversationTurnRepository(db).ListTurnsWithFeedback(ctx, conv.ID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "correct", rows[0].Text)
	require.Nil(t, rows[0].Feedback)
	_, err = repo.ApplySnapshot(ctx, conv.ID, "engine-test", 1, "original", turns, nil)
	require.NoError(t, err)
	_, err = repo.ApplySnapshot(ctx, conv.ID, "engine-test", 2, "corrected", corrected, nil)
	require.NoError(t, err)
	var app *common.AppError
	_, err = repo.ApplySnapshot(ctx, conv.ID, "engine-test", 2, "different", corrected, nil)
	require.ErrorAs(t, err, &app)
	require.Equal(t, http.StatusConflict, app.Status)
	require.Error(t, repo.StoreAnalysis(ctx, conv.ID, 1, f))
	for _, session := range []string{"", "wrong-engine"} {
		t.Run(fmt.Sprint("session", session), func(t *testing.T) {
			_, err := repo.ApplySnapshot(ctx, conv.ID, session, 3, "corrected", corrected, nil)
			require.ErrorAs(t, err, &app)
			require.Equal(t, http.StatusConflict, app.Status)
		})
	}
	duration := 0
	_, err = repo.ApplySnapshot(ctx, conv.ID, "engine-test", 2, "corrected", corrected, &duration)
	require.NoError(t, err)
	_, err = repo.ApplySnapshot(ctx, conv.ID, "engine-test", 2, "corrected", corrected, &duration)
	require.NoError(t, err)
	finished, err := repo.GetByID(ctx, conv.ID)
	require.NoError(t, err)
	require.True(t, finished.TranscriptFinalized)
	require.Equal(t, enums.ConversationStatusEnded, finished.Status)
	require.NotNil(t, finished.DurationSec)
	require.Zero(t, *finished.DurationSec)
	_, err = repo.ApplySnapshot(ctx, conv.ID, "engine-test", 3, "new", turns, nil)
	require.ErrorAs(t, err, &app)
	require.Equal(t, http.StatusConflict, app.Status)
}
