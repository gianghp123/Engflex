package tests

import (
	"context"
	"encoding/json"
	"engflex-api/internal/voice"
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
	"engflex-api/internal/modules/lessons/services"
	voicemocks "engflex-api/internal/voice/mocks"
	"engflex-api/internal/voice/pronunciation"
)

const (
	attachUserID   = "user-1"
	attachLessonID = "11111111-1111-1111-1111-111111111111"
	attachOpenID   = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	attachOtherAct = "44444444-4444-4444-4444-444444444444"
)

func attachCheckFixture() *models.LessonActivity {
	return &models.LessonActivity{
		LessonID: attachLessonID,
		ID:       "33333333-3333-3333-3333-333333333333",
		Type:     enums.ActivityTypeReading,
		Config: datatypes.JSON(`{"text":"t","questions":[` +
			`{"stem":"s0","instruction":"i","options":[],"answerKey":"B","explanation":"why0"},` +
			`{"stem":"s1","instruction":"i","options":[],"answerKey":"C","explanation":"why1"}]}`),
	}
}

func attachOpenAttempt() *models.Attempt {
	lessonID := attachLessonID
	return &models.Attempt{
		ID:       attachOpenID,
		UserID:   attachUserID,
		LessonID: &lessonID,
		Type:     enums.AttemptTypeLesson,
		Status:   enums.AttemptStatusInProgress,
		Result:   []byte(`{}`),
	}
}

func attachTwoActivities(actID string) []*models.LessonActivity {
	return []*models.LessonActivity{{ID: actID}, {ID: attachOtherAct}}
}

func TestCheckAnswerAttach(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "two outcomes merge at 50 then overwrite to 100",
			run: func(t *testing.T) {
				activity := attachCheckFixture()
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				repo.EXPECT().GetByID(mock.Anything, activity.ID).Return(activity, nil).Times(3)
				open := attachOpenAttempt()
				attempts.EXPECT().GetForUpdate(mock.Anything, attachOpenID).Return(open, nil).Times(3)
				repo.EXPECT().ListByLesson(mock.Anything, attachLessonID).Return(attachTwoActivities(activity.ID), nil).Times(3)
				var saved []*models.Attempt
				attempts.EXPECT().Update(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, m *models.Attempt) { saved = append(saved, m) }).
					Return(nil).Times(3)
				svc := newActivityService(t, repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llmmocks.NewMockLLMClient(t), attempts)

				got, err := svc.CheckAnswer(context.Background(), attachUserID, activity.ID, attachOpenID, 0, "B")
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.True(t, got.Correct)

				got, err = svc.CheckAnswer(context.Background(), attachUserID, activity.ID, attachOpenID, 1, "A")
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.False(t, got.Correct)

				require.Len(t, saved, 2)
				var parts map[string]struct {
					Type   enums.AttemptType        `json:"type"`
					Score  float64                  `json:"score"`
					Detail services.CheckPartDetail `json:"detail"`
				}
				require.NoError(t, json.Unmarshal(saved[1].Result, &parts))
				require.Contains(t, parts, activity.ID)
				assert.Equal(t, enums.AttemptTypeReading, parts[activity.ID].Type)
				assert.Equal(t, 50.0, parts[activity.ID].Score)
				require.Len(t, parts[activity.ID].Detail.Outcomes, 2)
				assert.Equal(t, enums.AttemptStatusInProgress, saved[1].Status)

				got, err = svc.CheckAnswer(context.Background(), attachUserID, activity.ID, attachOpenID, 1, "C")
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.True(t, got.Correct)

				require.Len(t, saved, 3)
				require.NoError(t, json.Unmarshal(saved[2].Result, &parts))
				require.Contains(t, parts, activity.ID)
				assert.Equal(t, 100.0, parts[activity.ID].Score)
				require.Len(t, parts[activity.ID].Detail.Outcomes, 2)
			},
		},
		{
			name: "missing attempt maps to 404",
			run: func(t *testing.T) {
				activity := attachCheckFixture()
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				repo.EXPECT().GetByID(mock.Anything, activity.ID).Return(activity, nil).Once()
				attempts.EXPECT().GetForUpdate(mock.Anything, attachOpenID).Return(nil, gorm.ErrRecordNotFound).Once()
				svc := newActivityService(t, repo, pronunciation.NewScorer(voicemocks.NewMockPronounceEngine(t)), llmmocks.NewMockLLMClient(t), attempts)

				got, err := svc.CheckAnswer(context.Background(), attachUserID, activity.ID, attachOpenID, 0, "B")

				requireAppError(t, err, http.StatusNotFound)
				assert.Nil(t, got)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func attachSpeakingFixture() *models.LessonActivity {
	return &models.LessonActivity{
		LessonID: attachLessonID,
		ID:       "66666666-6666-6666-6666-666666666666",
		Type:     enums.ActivityTypeSpeaking,
		Config: datatypes.JSON(`{"items":[` +
			`{"text":"Please call me back.","modelAudioKey":"speaking/x/0.mp3","modelVoice":"en-GB-SoniaNeural"},` +
			`{"text":"Good morning.","modelAudioKey":"speaking/x/1.mp3","modelVoice":"en-GB-SoniaNeural"}]}`),
	}
}

func cannedScore(score float64) *voice.PronounceResult {
	return &voice.PronounceResult{Score: score, Transcription: "heard words"}
}

func TestPronounceAttach(t *testing.T) {
	const audioMime = "audio/webm"
	audio := []byte("fake-webm")

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "latest wins per item with mean score",
			run: func(t *testing.T) {
				activity := attachSpeakingFixture()
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				voice := voicemocks.NewMockPronounceEngine(t)
				voice.On("Pronounce", mock.Anything, "Please call me back.", "en", audio, audioMime).
					Return(cannedScore(80), nil).Once()
				voice.On("Pronounce", mock.Anything, "Good morning.", "en", audio, audioMime).
					Return(cannedScore(60), nil).Once()
				voice.On("Pronounce", mock.Anything, "Please call me back.", "en", audio, audioMime).
					Return(cannedScore(95), nil).Once()
				repo.EXPECT().GetByID(mock.Anything, activity.ID).Return(activity, nil).Times(3)
				open := attachOpenAttempt()
				attempts.EXPECT().GetForUpdate(mock.Anything, attachOpenID).Return(open, nil).Times(6)
				repo.EXPECT().ListByLesson(mock.Anything, attachLessonID).Return(attachTwoActivities(activity.ID), nil).Times(3)
				var saved []*models.Attempt
				attempts.EXPECT().Update(mock.Anything, mock.Anything).
					Run(func(ctx context.Context, m *models.Attempt) { saved = append(saved, m) }).
					Return(nil).Times(3)
				svc := newActivityService(t, repo, pronunciation.NewScorer(voice), llmmocks.NewMockLLMClient(t), attempts)

				got, err := svc.Pronounce(context.Background(), attachUserID, activity.ID, attachOpenID, 0, audio, audioMime)
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, 80.0, got.Score)

				got, err = svc.Pronounce(context.Background(), attachUserID, activity.ID, attachOpenID, 1, audio, audioMime)
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, 60.0, got.Score)

				require.Len(t, saved, 2)
				var parts map[string]struct {
					Type   enums.AttemptType            `json:"type"`
					Score  float64                      `json:"score"`
					Detail services.PronouncePartDetail `json:"detail"`
				}
				require.NoError(t, json.Unmarshal(saved[1].Result, &parts))
				require.Contains(t, parts, activity.ID)
				assert.Equal(t, enums.AttemptTypeSpeaking, parts[activity.ID].Type)
				assert.Equal(t, 70.0, parts[activity.ID].Score)
				assert.Equal(t, map[string]float64{"0": 80, "1": 60}, parts[activity.ID].Detail.Items)
				assert.Equal(t, enums.AttemptStatusInProgress, saved[1].Status)

				got, err = svc.Pronounce(context.Background(), attachUserID, activity.ID, attachOpenID, 0, audio, audioMime)
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, 95.0, got.Score)

				require.Len(t, saved, 3)
				require.NoError(t, json.Unmarshal(saved[2].Result, &parts))
				require.Contains(t, parts, activity.ID)
				assert.Equal(t, 77.5, parts[activity.ID].Score)
				assert.Equal(t, map[string]float64{"0": 95, "1": 60}, parts[activity.ID].Detail.Items)
			},
		},
		{
			name: "missing attempt maps to 404 without engine",
			run: func(t *testing.T) {
				activity := attachSpeakingFixture()
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				voice := voicemocks.NewMockPronounceEngine(t)
				repo.EXPECT().GetByID(mock.Anything, activity.ID).Return(activity, nil).Once()
				attempts.EXPECT().GetForUpdate(mock.Anything, attachOpenID).Return(nil, gorm.ErrRecordNotFound).Once()
				svc := newActivityService(t, repo, pronunciation.NewScorer(voice), llmmocks.NewMockLLMClient(t), attempts)

				got, err := svc.Pronounce(context.Background(), attachUserID, activity.ID, attachOpenID, 0, audio, audioMime)

				requireAppError(t, err, http.StatusNotFound)
				assert.Nil(t, got)
				voice.AssertNotCalled(t, "Pronounce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
			},
		},
		{
			name: "foreign attempt maps to 404 without engine",
			run: func(t *testing.T) {
				activity := attachSpeakingFixture()
				repo := mocks.NewMockLessonActivityRepository(t)
				attempts := mocks.NewMockAttemptRepository(t)
				voice := voicemocks.NewMockPronounceEngine(t)
				repo.EXPECT().GetByID(mock.Anything, activity.ID).Return(activity, nil).Once()
				foreign := attachOpenAttempt()
				foreign.UserID = "user_999"
				attempts.EXPECT().GetForUpdate(mock.Anything, attachOpenID).Return(foreign, nil).Once()
				svc := newActivityService(t, repo, pronunciation.NewScorer(voice), llmmocks.NewMockLLMClient(t), attempts)

				got, err := svc.Pronounce(context.Background(), attachUserID, activity.ID, attachOpenID, 0, audio, audioMime)

				requireAppError(t, err, http.StatusNotFound)
				assert.Nil(t, got)
				voice.AssertNotCalled(t, "Pronounce", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				attempts.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
