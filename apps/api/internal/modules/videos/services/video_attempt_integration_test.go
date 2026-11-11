package services_test

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"testing"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/modules/videos/services"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestDictationAttemptPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a migrated PostgreSQL test database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	pool, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, pool.Close()) })

	video := models.VideoExercise{Title: "Dictation attempt fixture", VideoURL: "fixture"}
	require.NoError(t, db.Create(&video).Error)
	t.Cleanup(func() { require.NoError(t, db.Delete(&video).Error) })
	lines := []models.VideoTranscript{
		{VideoExerciseID: video.ID, Sequence: 1, Content: "First line"},
		{VideoExerciseID: video.ID, Sequence: 2, Content: "Second line"},
	}
	require.NoError(t, db.Create(&lines).Error)

	svc := services.NewVideoAttemptService(
		repositories.NewAttemptRepository(db),
		repositories.NewVideoExerciseRepository(db),
		repositories.NewVideoTranscriptRepository(db),
		repositories.NewTransactionManager(db),
	)
	ctx := t.Context()
	user := "dictation-test-" + uuid.NewString()
	start := requests.StartVideoAttempt{Type: enums.AttemptTypeDictation}
	var attempts [2]*models.Attempt
	var startErrors [2]error
	var starts sync.WaitGroup
	for i := range attempts {
		starts.Go(func() {
			attempts[i], startErrors[i] = svc.StartAttempt(ctx, user, video.ID, start.Type, start)
		})
	}
	starts.Wait()
	require.NoError(t, startErrors[0])
	require.NoError(t, startErrors[1])
	id := attempts[0].ID
	require.Equal(t, id, attempts[1].ID)

	requireStatus := func(err error, status int) {
		t.Helper()
		var appErr *common.AppError
		require.ErrorAs(t, err, &appErr)
		require.Equal(t, status, appErr.Status)
	}
	foreign := uuid.NewString()
	for _, progress := range []requests.SaveVideoProgress{
		{ActiveTranscriptID: &foreign},
		{CompletedTranscriptIDs: []string{foreign}},
		{Transcripts: map[string]any{foreign: "draft"}},
	} {
		_, err := svc.SaveProgress(ctx, user, video.ID, id, progress)
		requireStatus(err, http.StatusBadRequest)
	}
	_, err = svc.SaveProgress(ctx, "other-user", video.ID, id, requests.SaveVideoProgress{})
	requireStatus(err, http.StatusNotFound)

	for i, line := range lines {
		_, err := svc.SaveProgress(ctx, user, video.ID, id, requests.SaveVideoProgress{
			ActiveSequence: i, ActiveTranscriptID: &line.ID,
			CompletedTranscriptIDs: []string{line.ID, line.ID},
			Transcripts:            map[string]any{line.ID: map[string]any{"draft": line.Content}},
		})
		require.NoError(t, err)
		if i == 0 {
			_, err := svc.CompleteAttempt(ctx, user, video.ID, id, requests.CompleteVideoAttempt{})
			requireStatus(err, http.StatusConflict)
		}
	}
	open, err := svc.GetOpenAttempt(ctx, user, video.ID, start.Type)
	require.NoError(t, err)
	require.Equal(t, id, open.ID)
	var restored requests.SaveVideoProgress
	require.NoError(t, json.Unmarshal(open.Result, &restored))
	require.Equal(t, []string{lines[0].ID, lines[1].ID}, restored.CompletedTranscriptIDs)
	require.Equal(t, &lines[1].ID, restored.ActiveTranscriptID)
	require.Len(t, restored.Transcripts, 2)

	var writes sync.WaitGroup
	var saveErr, completeErr error
	writes.Go(func() {
		_, saveErr = svc.SaveProgress(ctx, user, video.ID, id, requests.SaveVideoProgress{
			ActiveTranscriptID: &lines[0].ID,
		})
	})
	writes.Go(func() {
		_, completeErr = svc.CompleteAttempt(ctx, user, video.ID, id, requests.CompleteVideoAttempt{})
	})
	writes.Wait()
	require.NoError(t, saveErr)
	require.NoError(t, completeErr)
	closed, err := svc.CompleteAttempt(ctx, user, video.ID, id, requests.CompleteVideoAttempt{})
	require.NoError(t, err)
	require.Equal(t, enums.AttemptStatusCompleted, closed.Status)
	require.NotNil(t, closed.SubmittedAt)
	late, err := svc.SaveProgress(ctx, user, video.ID, id, requests.SaveVideoProgress{
		Transcripts: map[string]any{lines[0].ID: "late write"},
	})
	require.NoError(t, err)
	require.Equal(t, enums.AttemptStatusCompleted, late.Status)
	require.JSONEq(t, string(closed.Result), string(late.Result))
	_, err = svc.GetOpenAttempt(ctx, user, video.ID, start.Type)
	requireStatus(err, http.StatusNotFound)

	level := "normal"
	scored, err := svc.StartAttempt(ctx, user, video.ID, start.Type, requests.StartVideoAttempt{Type: start.Type, Level: &level})
	require.NoError(t, err)
	require.NotEqual(t, id, scored.ID)
	easy := "easy"
	_, err = svc.SaveProgress(ctx, user, video.ID, scored.ID, requests.SaveVideoProgress{Level: &easy})
	requireStatus(err, http.StatusConflict)
	_, err = svc.SaveProgress(ctx, user, video.ID, scored.ID, requests.SaveVideoProgress{
		Transcripts: map[string]any{lines[0].ID: requests.DictationLineInput{Revealed: []int{1}}},
	})
	requireStatus(err, http.StatusBadRequest)
	_, err = svc.SaveProgress(ctx, user, video.ID, scored.ID, requests.SaveVideoProgress{
		CompletedTranscriptIDs: []string{lines[0].ID, lines[1].ID},
		Transcripts: map[string]any{
			lines[0].ID: requests.DictationLineInput{Answers: []string{" FIRST! "}, Completed: true},
		},
	})
	require.NoError(t, err)
	_, err = svc.CompleteAttempt(ctx, user, video.ID, scored.ID, requests.CompleteVideoAttempt{})
	requireStatus(err, http.StatusConflict)
	_, err = svc.SaveProgress(ctx, user, video.ID, scored.ID, requests.SaveVideoProgress{
		Transcripts: map[string]any{
			lines[0].ID: requests.DictationLineInput{Revealed: []int{0}},
			lines[1].ID: requests.DictationLineInput{Completed: true, RevealAll: true},
		},
	})
	require.NoError(t, err)
	fakeScore := 100.0
	finished, err := svc.CompleteAttempt(ctx, user, video.ID, scored.ID, requests.CompleteVideoAttempt{Score: &fakeScore})
	require.NoError(t, err)
	require.NotNil(t, finished.Score)
	require.Equal(t, 50.0, *finished.Score)
}
