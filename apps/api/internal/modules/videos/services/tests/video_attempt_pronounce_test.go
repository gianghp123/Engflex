package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	conversationsresponses "engflex-api/internal/modules/conversations/dtos/responses"
	videoresponses "engflex-api/internal/modules/videos/dtos/responses"
	"engflex-api/internal/modules/videos/services"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
)

const (
	pronounceVideoID      = "11111111-1111-1111-1111-111111111111"
	pronounceAttemptID    = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	pronounceTranscriptID = "33333333-3333-3333-3333-333333333333"
	pronounceUserID       = "user-1"
)

func pronounceTranscriptFixture() *models.VideoTranscript {
	return &models.VideoTranscript{
		ID:              pronounceTranscriptID,
		VideoExerciseID: pronounceVideoID,
		Content:         "Please call me back.",
	}
}

func pronounceAttemptFixture(attemptType enums.AttemptType, status enums.AttemptStatus) *models.Attempt {
	videoID := pronounceVideoID
	return &models.Attempt{
		ID: pronounceAttemptID, UserID: pronounceUserID, VideoExerciseID: &videoID,
		Type: attemptType, Status: status,
	}
}

func TestVideoAttemptService_Pronounce(t *testing.T) {
	audio := []byte("fake-webm")
	want := &conversationsresponses.PronounceResult{
		Score:         88.5,
		Transcription: "please call me back",
		Feedback:      "You need to better pronounce: call",
		Errors: []conversationsresponses.MispronouncedWord{
			{Word: "call", Expected: "kɔːl", Heard: "kɑːl", Confidence: 0.4},
		},
	}

	tests := []struct {
		name         string
		setup        func(transcripts *mocks.MockVideoTranscriptRepository, attempts *mocks.MockAttemptRepository, voice *voicemocks.MockPronounceEngine)
		wantStatus   int
		wantPersists bool
	}{
		{
			name: "happy path scores against the transcript line's text and persists it onto the attempt",
			setup: func(transcripts *mocks.MockVideoTranscriptRepository, attempts *mocks.MockAttemptRepository, voice *voicemocks.MockPronounceEngine) {
				transcripts.EXPECT().GetByID(mock.Anything, pronounceTranscriptID).
					Return(pronounceTranscriptFixture(), nil).Once()
				attempts.EXPECT().GetByID(mock.Anything, pronounceAttemptID).
					Return(pronounceAttemptFixture(enums.AttemptTypeShadowing, enums.AttemptStatusInProgress), nil).Once()
				voice.On("Pronounce", mock.Anything, "Please call me back.", "en", audio, "audio/webm").
					Return(want, nil).Once()
				attempts.EXPECT().WithTx(mock.Anything).Return(attempts).Once()
				attempts.EXPECT().GetVideoAttemptForUpdate(mock.Anything, pronounceUserID, pronounceVideoID, pronounceAttemptID).
					Return(pronounceAttemptFixture(enums.AttemptTypeShadowing, enums.AttemptStatusInProgress), nil).Once()
				attempts.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantPersists: true,
		},
		{
			name: "transcript not found",
			setup: func(transcripts *mocks.MockVideoTranscriptRepository, attempts *mocks.MockAttemptRepository, voice *voicemocks.MockPronounceEngine) {
				transcripts.EXPECT().GetByID(mock.Anything, pronounceTranscriptID).
					Return(nil, gorm.ErrRecordNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "transcript belongs to a different video",
			setup: func(transcripts *mocks.MockVideoTranscriptRepository, attempts *mocks.MockAttemptRepository, voice *voicemocks.MockPronounceEngine) {
				other := pronounceTranscriptFixture()
				other.VideoExerciseID = "22222222-2222-2222-2222-222222222222"
				transcripts.EXPECT().GetByID(mock.Anything, pronounceTranscriptID).Return(other, nil).Once()
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "attempt not found",
			setup: func(transcripts *mocks.MockVideoTranscriptRepository, attempts *mocks.MockAttemptRepository, voice *voicemocks.MockPronounceEngine) {
				transcripts.EXPECT().GetByID(mock.Anything, pronounceTranscriptID).
					Return(pronounceTranscriptFixture(), nil).Once()
				attempts.EXPECT().GetByID(mock.Anything, pronounceAttemptID).
					Return(nil, gorm.ErrRecordNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "attempt belongs to a different user",
			setup: func(transcripts *mocks.MockVideoTranscriptRepository, attempts *mocks.MockAttemptRepository, voice *voicemocks.MockPronounceEngine) {
				transcripts.EXPECT().GetByID(mock.Anything, pronounceTranscriptID).
					Return(pronounceTranscriptFixture(), nil).Once()
				foreign := pronounceAttemptFixture(enums.AttemptTypeShadowing, enums.AttemptStatusInProgress)
				foreign.UserID = "someone-else"
				attempts.EXPECT().GetByID(mock.Anything, pronounceAttemptID).Return(foreign, nil).Once()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "attempt is not a shadowing exercise",
			setup: func(transcripts *mocks.MockVideoTranscriptRepository, attempts *mocks.MockAttemptRepository, voice *voicemocks.MockPronounceEngine) {
				transcripts.EXPECT().GetByID(mock.Anything, pronounceTranscriptID).
					Return(pronounceTranscriptFixture(), nil).Once()
				attempts.EXPECT().GetByID(mock.Anything, pronounceAttemptID).
					Return(pronounceAttemptFixture(enums.AttemptTypeDictation, enums.AttemptStatusInProgress), nil).Once()
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "attempt is closed",
			setup: func(transcripts *mocks.MockVideoTranscriptRepository, attempts *mocks.MockAttemptRepository, voice *voicemocks.MockPronounceEngine) {
				transcripts.EXPECT().GetByID(mock.Anything, pronounceTranscriptID).
					Return(pronounceTranscriptFixture(), nil).Once()
				attempts.EXPECT().GetByID(mock.Anything, pronounceAttemptID).
					Return(pronounceAttemptFixture(enums.AttemptTypeShadowing, enums.AttemptStatusCompleted), nil).Once()
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transcripts := mocks.NewMockVideoTranscriptRepository(t)
			attempts := mocks.NewMockAttemptRepository(t)
			voice := voicemocks.NewMockPronounceEngine(t)
			var transactions *mocks.MockTransactionManager
			if tt.wantPersists {
				transactions = mocks.NewMockTransactionManager(t)
				transactions.EXPECT().Run(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, fn func(*gorm.DB) error) error {
						return fn(nil)
					}).Once()
			}
			tt.setup(transcripts, attempts, voice)
			svc := services.NewVideoAttemptService(
				attempts, mocks.NewMockVideoExerciseRepository(t), transcripts,
				transactions, pronunciation.NewScorer(voice), nil,
			)

			got, err := svc.Pronounce(context.Background(), pronounceUserID, pronounceVideoID, pronounceAttemptID, pronounceTranscriptID, audio, "audio/webm")

			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, 88.5, got.Score)
			assert.Equal(t, "please call me back", got.Transcription)
			assert.Equal(t, "You need to better pronounce: call", got.Feedback)
			require.Len(t, got.Errors, 1)
			assert.Equal(t, "call", got.Errors[0].Word)
		})
	}
}

func TestVideoAttemptService_Pronounce_PersistsScoreOntoAttempt(t *testing.T) {
	audio := []byte("fake-webm")
	want := &conversationsresponses.PronounceResult{Score: 73, Transcription: "please call me back", Feedback: "nice work"}

	transcripts := mocks.NewMockVideoTranscriptRepository(t)
	attempts := mocks.NewMockAttemptRepository(t)
	voice := voicemocks.NewMockPronounceEngine(t)
	transactions := mocks.NewMockTransactionManager(t)

	transcripts.EXPECT().GetByID(mock.Anything, pronounceTranscriptID).
		Return(pronounceTranscriptFixture(), nil).Once()
	attempts.EXPECT().GetByID(mock.Anything, pronounceAttemptID).
		Return(pronounceAttemptFixture(enums.AttemptTypeShadowing, enums.AttemptStatusInProgress), nil).Once()
	voice.On("Pronounce", mock.Anything, "Please call me back.", "en", audio, "audio/webm").
		Return(want, nil).Once()
	transactions.EXPECT().Run(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, fn func(*gorm.DB) error) error {
			return fn(nil)
		}).Once()
	attempts.EXPECT().WithTx(mock.Anything).Return(attempts).Once()
	persisted := pronounceAttemptFixture(enums.AttemptTypeShadowing, enums.AttemptStatusInProgress)
	attempts.EXPECT().GetVideoAttemptForUpdate(mock.Anything, pronounceUserID, pronounceVideoID, pronounceAttemptID).
		Return(persisted, nil).Once()
	attempts.EXPECT().Update(mock.Anything, persisted).Return(nil).Once()

	svc := services.NewVideoAttemptService(
		attempts, mocks.NewMockVideoExerciseRepository(t), transcripts,
		transactions, pronunciation.NewScorer(voice), nil,
	)

	got, err := svc.Pronounce(context.Background(), pronounceUserID, pronounceVideoID, pronounceAttemptID, pronounceTranscriptID, audio, "audio/webm")
	require.NoError(t, err)
	require.NotNil(t, got)

	var payload videoresponses.VideoProgressPayload
	require.NoError(t, json.Unmarshal(persisted.Result, &payload))
	raw, ok := payload.Transcripts[pronounceTranscriptID]
	require.True(t, ok, "expected a persisted line result for the scored transcript")
	encoded, err := json.Marshal(raw)
	require.NoError(t, err)
	var line videoresponses.ShadowingLineResult
	require.NoError(t, json.Unmarshal(encoded, &line))
	assert.InDelta(t, 73, line.Score, 0.001)
	assert.Equal(t, "nice work", line.Feedback)
	assert.Equal(t, "please call me back", line.Transcription)
	assert.False(t, line.AttemptedAt.IsZero())
}
