package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	llmmocks "engflex-api/internal/llm/mocks"
	conversationsresponses "engflex-api/internal/modules/conversations/dtos/responses"
	"engflex-api/internal/modules/lessons/services"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
)

func pronounceSpeakingFixture() *models.LessonActivity {
	return &models.LessonActivity{
		ID:     "66666666-6666-6666-6666-666666666666",
		Type:   enums.ActivityTypeSpeaking,
		Config: datatypes.JSON(`{"items":[{"text":"Please call me back.","modelAudioKey":"speaking/x/0.mp3","modelVoice":"en-GB-SoniaNeural"}]}`),
	}
}

func TestPronounce(t *testing.T) {
	speaking := pronounceSpeakingFixture()
	reading := checkReadingFixture()
	audio := []byte("fake-webm")
	want := &conversationsresponses.PronounceResult{
		Score:            88.5,
		Transcription:    "please call me back",
		PhonemeErrorRate: 0.1,
		WordErrorRate:    0.2,
		AcousticDistance: 2.5,
		Errors: []conversationsresponses.MispronouncedWord{
			{Word: "think", Expected: "θɪŋk", Heard: "tɪŋk", Confidence: 0.9},
		},
		ReferencePhones: []conversationsresponses.ReferencePhone{
			{Word: "call", Phones: "kɔːl"},
			{Word: "back", Phones: "bæk"},
		},
	}

	tests := []struct {
		name       string
		activity   *models.LessonActivity
		repoErr    error
		itemIndex  int
		setupVoice func(*voicemocks.MockPronounceEngine)
		wantStatus int
	}{
		{
			name:      "happy path maps scorer result",
			activity:  speaking,
			itemIndex: 0,
			setupVoice: func(voice *voicemocks.MockPronounceEngine) {
				voice.On("Pronounce", mock.Anything, "Please call me back.", "en", audio, "audio/webm").
					Return(want, nil).Once()
			},
		},
		{name: "item index negative", activity: speaking, itemIndex: -1, setupVoice: func(voice *voicemocks.MockPronounceEngine) {}, wantStatus: http.StatusBadRequest},
		{name: "item index out of range", activity: speaking, itemIndex: 1, setupVoice: func(voice *voicemocks.MockPronounceEngine) {}, wantStatus: http.StatusBadRequest},
		{name: "reading not pronounceable", activity: reading, itemIndex: 0, setupVoice: func(voice *voicemocks.MockPronounceEngine) {}, wantStatus: http.StatusUnprocessableEntity},
		{name: "not found", activity: nil, repoErr: gorm.ErrRecordNotFound, itemIndex: 0, setupVoice: func(voice *voicemocks.MockPronounceEngine) {}, wantStatus: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockLessonActivityRepository(t)
			id := "66666666-6666-6666-6666-666666666666"
			if tt.activity != nil {
				id = tt.activity.ID
			}
			if tt.name == "not found" {
				id = "99999999-9999-9999-9999-999999999999"
			}
			repo.EXPECT().GetByID(mock.Anything, id).Return(tt.activity, tt.repoErr).Once()
			voice := voicemocks.NewMockPronounceEngine(t)
			tt.setupVoice(voice)
			attempts := mocks.NewMockAttemptRepository(t)
			const userID = "user-1"
			const attemptID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
			if tt.wantStatus == 0 {
				lessonID := "11111111-1111-1111-1111-111111111111"
				attempts.EXPECT().GetByID(mock.Anything, attemptID).Return(&models.Attempt{
					ID: attemptID, UserID: userID, LessonID: &lessonID,
					Type: enums.AttemptTypeLesson, Status: enums.AttemptStatusInProgress,
					Result: []byte(`{}`),
				}, nil).Twice()
				repo.EXPECT().ListByLesson(mock.Anything, lessonID).Return([]*models.LessonActivity{{ID: id}}, nil).Once()
				attempts.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			}
			svc := services.NewLessonActivityService(repo, pronunciation.NewScorer(voice), llmmocks.NewMockLLMClient(t), attempts)
			got, err := svc.Pronounce(context.Background(), userID, id, attemptID, tt.itemIndex, audio, "audio/webm")
			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				assert.Nil(t, got)
				voice.AssertNotCalled(t, "Pronounce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, 88.5, got.Score)
			assert.Equal(t, "please call me back", got.Transcription)
			require.Len(t, got.Errors, 1)
			assert.Equal(t, "think", got.Errors[0].Word)
			assert.Equal(t, 1, got.TotalErrors)
			assert.Contains(t, got.Errors[0].Tip, "θ")
			// Reference phones survive both maps: conversations DTO → scorer
			// Result → lessons DTO (web copies whole graphs).
			require.Len(t, got.ReferencePhones, 2)
			assert.Equal(t, "call", got.ReferencePhones[0].Word)
			assert.Equal(t, "kɔːl", got.ReferencePhones[0].Phones)
			assert.Equal(t, "bæk", got.ReferencePhones[1].Phones)
		})
	}
}
