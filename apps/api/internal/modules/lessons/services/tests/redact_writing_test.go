package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/database/models"
	"engflex-api/internal/database/repositories/mocks"
	llmmocks "engflex-api/internal/llm/mocks"
	"engflex-api/internal/modules/lessons/services"
	pronunciationmocks "engflex-api/internal/voice/pronunciation/mocks"
)

const redactAttemptID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

func writingAttempt(userID string, rawText string) *models.Attempt {
	detail := map[string]any{
		"feedback":  map[string]any{},
		"rawHash":   "deadbeef",
		"wordCount": 3,
	}
	if rawText != "" {
		detail["rawText"] = rawText
	}
	raw, _ := json.Marshal(map[string]any{
		"writing-1": map[string]any{"type": "writing", "score": 75, "detail": detail},
		"reading-1": map[string]any{"type": "reading", "score": 100, "detail": map[string]any{"outcomes": []any{}}},
	})
	return &models.Attempt{ID: redactAttemptID, UserID: userID, Result: raw}
}

func TestRedactWriting(t *testing.T) {
	tests := []struct {
		name       string
		attempt    *models.Attempt
		userID     string
		wantUpdate bool
		wantStatus int
		check      func(*testing.T, *models.Attempt)
	}{
		{
			name:       "clears raw text and keeps the receipt",
			attempt:    writingAttempt("user-1", "my private draft"),
			userID:     "user-1",
			wantUpdate: true,
			check: func(t *testing.T, got *models.Attempt) {
				t.Helper()
				assert.NotContains(t, string(got.Result), "my private draft")
				var parts map[string]struct {
					Detail services.WritingPartDetail `json:"detail"`
				}
				require.NoError(t, json.Unmarshal(got.Result, &parts))
				assert.Empty(t, parts["writing-1"].Detail.RawText)
				assert.Equal(t, "deadbeef", parts["writing-1"].Detail.RawHash)
				// Non-writing parts are untouched.
				assert.Contains(t, string(got.Result), `"reading-1"`)
			},
		},
		{
			name:       "another user's attempt is 404 and writes nothing",
			attempt:    writingAttempt("someone-else", "secret"),
			userID:     "user-1",
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "nothing to redact does not write",
			attempt: writingAttempt("user-1", ""),
			userID:  "user-1",
			check: func(t *testing.T, got *models.Attempt) {
				t.Helper()
				assert.NotContains(t, string(got.Result), "rawText")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockLessonActivityRepository(t)
			attempts := mocks.NewMockAttemptRepository(t)
			attempts.EXPECT().GetByID(mock.Anything, redactAttemptID).Return(tt.attempt, nil).Once()
			if tt.wantUpdate {
				attempts.EXPECT().Update(mock.Anything, mock.Anything).Return(nil).Once()
			}
			svc := services.NewLessonActivityService(
				repo,
				pronunciationmocks.NewMockScorer(t),
				llmmocks.NewMockLLMClient(t),
				attempts,
				mocks.NewMockLessonRepository(t),
				fakePremiumChecker{},
			)

			got, err := svc.RedactWriting(context.Background(), tt.userID, redactAttemptID)

			if tt.wantStatus != 0 {
				requireAppError(t, err, tt.wantStatus)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}
