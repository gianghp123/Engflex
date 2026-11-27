package tests

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/videos/dtos/requests"
	videoresponses "engflex-api/internal/modules/videos/dtos/responses"
	"engflex-api/internal/modules/videos/services"
)

// A checkpoint save that doesn't touch a line must not erase a pronunciation
// score Pronounce already persisted for it: use-shadowing-attempt.ts always
// sends an empty Transcripts map on line navigation, so SaveProgress has to
// merge forward (previous entries survive unless this call overwrites them)
// instead of replacing the whole map with what the request carried.
func TestVideoAttemptService_SaveProgress_PreservesShadowingScoreOnUnrelatedCheckpoint(t *testing.T) {
	videoID := "11111111-1111-1111-1111-111111111111"
	attemptID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	userID := "user-1"
	lineA := "33333333-3333-3333-3333-333333333333"
	lineB := "44444444-4444-4444-4444-444444444444"

	existing := videoresponses.ShadowingLineResult{Score: 90, Feedback: "nice", Transcription: "please call me back"}
	previousResult, err := json.Marshal(videoresponses.VideoProgressPayload{
		ActiveSequence:         0,
		CompletedTranscriptIDs: []string{},
		Transcripts:            map[string]any{lineA: existing},
	})
	require.NoError(t, err)

	attemptFixture := func() *models.Attempt {
		vID := videoID
		return &models.Attempt{
			ID: attemptID, UserID: userID, VideoExerciseID: &vID,
			Type: enums.AttemptTypeShadowing, Status: enums.AttemptStatusInProgress,
			Result: previousResult,
		}
	}

	transcripts := mocks.NewMockVideoTranscriptRepository(t)
	attempts := mocks.NewMockAttemptRepository(t)
	transactions := mocks.NewMockTransactionManager(t)

	transcripts.EXPECT().WithTx(mock.Anything).Return(transcripts).Once()
	transcripts.EXPECT().GetByVideoExerciseID(mock.Anything, videoID).Return([]*models.VideoTranscript{
		{ID: lineA, VideoExerciseID: videoID, Sequence: 1, Content: "Please call me back."},
		{ID: lineB, VideoExerciseID: videoID, Sequence: 2, Content: "I will call you back."},
	}, nil).Once()
	transactions.EXPECT().Run(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(*gorm.DB) error) error {
			return fn(nil)
		}).Once()
	attempts.EXPECT().WithTx(mock.Anything).Return(attempts).Once()
	persisted := attemptFixture()
	attempts.EXPECT().GetVideoAttemptForUpdate(mock.Anything, userID, videoID, attemptID).
		Return(persisted, nil).Once()
	attempts.EXPECT().Update(mock.Anything, persisted).Return(nil).Once()

	svc := services.NewVideoAttemptService(
		attempts, mocks.NewMockVideoExerciseRepository(t), transcripts, transactions, nil, nil,
	)

	_, err = svc.SaveProgress(context.Background(), userID, videoID, attemptID, requests.SaveVideoProgress{
		ActiveSequence:         1,
		ActiveTranscriptID:     &lineB,
		CompletedTranscriptIDs: []string{lineA},
		Transcripts:            map[string]any{},
	})
	require.NoError(t, err)

	var payload videoresponses.VideoProgressPayload
	require.NoError(t, json.Unmarshal(persisted.Result, &payload))
	raw, ok := payload.Transcripts[lineA]
	require.True(t, ok, "expected the previously persisted shadowing score for lineA to survive")
	encoded, err := json.Marshal(raw)
	require.NoError(t, err)
	var line videoresponses.ShadowingLineResult
	require.NoError(t, json.Unmarshal(encoded, &line))
	assert.InDelta(t, 90, line.Score, 0.001)
	assert.Equal(t, "nice", line.Feedback)
}
