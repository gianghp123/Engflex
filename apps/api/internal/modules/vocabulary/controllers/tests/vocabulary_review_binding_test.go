package tests

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"engflex-api/internal/modules/vocabulary/dtos/requests"
)

// TestReviewUserVocabularyBinding pins the grade binding: 0 ("again") is a
// valid grade, so the field is a *int — `required` rejects a missing grade
// while `oneof` still bounds the value. A plain int with `required` would
// reject grade 0 (validator treats zero as missing).
func TestReviewUserVocabularyBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		body      string
		wantOK    bool
		wantGrade int
	}{
		{"grade 0 binds", `{"grade":0}`, true, 0},
		{"grade 1 binds", `{"grade":1}`, true, 1},
		{"grade 2 binds", `{"grade":2}`, true, 2},
		{"grade 3 binds", `{"grade":3}`, true, 3},
		{"grade 4 rejected", `{"grade":4}`, false, 0},
		{"grade -1 rejected", `{"grade":-1}`, false, 0},
		{"missing grade rejected", `{}`, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("POST", "/user-vocabulary/item-1/review", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			var got requests.ReviewUserVocabulary
			err := c.ShouldBindJSON(&got)

			if tt.wantOK {
				require.NoError(t, err)
				require.NotNil(t, got.Grade)
				assert.Equal(t, tt.wantGrade, *got.Grade)
			} else {
				require.Error(t, err)
			}
		})
	}
}
