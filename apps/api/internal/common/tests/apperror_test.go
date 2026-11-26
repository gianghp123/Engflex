// Package tests black-boxes the shared error contract: With* builders must
// copy (never mutate the base error, matching withCause), and Fail must
// carry Details in the envelope only when present so legacy bodies are
// byte-identical.
package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/common"
)

func TestWithCodeAndDetailsReturnCopies(t *testing.T) {
	tests := []struct {
		name string
		base *common.AppError
	}{
		{name: "plain forbidden", base: common.Forbidden("gated")},
		{name: "db mapped", base: common.NotFound("lesson not found")},
		{name: "service unavailable", base: common.ServiceUnavailable("scoring unavailable")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mutated := tt.base.WithCode(common.CodeAIUnavailable).WithDetails(map[string]int{"remaining": 0})
			require.Equal(t, common.CodeAIUnavailable, mutated.Code)
			require.NotNil(t, mutated.Details)
			require.Empty(t, tt.base.Code, "the base error must stay unmutated")
			require.Nil(t, tt.base.Details, "the base error must stay unmutated")
			// Error() keeps its existing contract: message (+ cause), never the code.
			assert.Equal(t, tt.base.Error(), common.New(tt.base.Status, tt.base.Message).Error())
		})
	}
}

func TestFailWritesDetailsOnlyWhenPresent(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   map[string]any
	}{
		{
			name:       "quota details cross the wire",
			err:        common.TooManyRequests("insufficient AI credits").WithDetails(map[string]any{"code": common.CodeInsufficientCredits, "remaining": 0}),
			wantStatus: http.StatusTooManyRequests,
			wantBody: map[string]any{
				"message": "insufficient AI credits",
				"data":    map[string]any{"code": common.CodeInsufficientCredits, "remaining": float64(0)},
			},
		},
		{
			name:       "legacy errors keep the exact old envelope",
			err:        common.NotFound("lesson not found"),
			wantStatus: http.StatusNotFound,
			wantBody:   map[string]any{"message": "lesson not found"},
		},
		{
			name:       "hidden internals stay hidden",
			err:        common.Internal().WithCode(common.CodePolicyUnavailable),
			wantStatus: http.StatusInternalServerError,
			wantBody:   map[string]any{"message": "Internal Server Error"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			common.Fail(c, tt.err)
			require.Equal(t, tt.wantStatus, w.Code)
			var got map[string]any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
			assert.Equal(t, tt.wantBody, got)
		})
	}
}
