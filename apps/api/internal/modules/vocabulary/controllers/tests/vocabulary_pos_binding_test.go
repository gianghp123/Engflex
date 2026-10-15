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

// TestCreateVocabularyItemBinding pins the part-of-speech binding. The column is
// nullable and CHECK-constrained, so an absent/null value binds (nil), and any
// token outside the 12 members must fail at the wire boundary — the controller
// maps a bind error to 400, so an unknown token never reaches the DB CHECK as a
// hidden 500.
func TestCreateVocabularyItemBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		body    string
		wantOK  bool
		wantPOS *string
	}{
		{"member binds", `{"term":"give up","cefr":"B2","partOfSpeech":"phrasal_verb"}`, true, posPtr("phrasal_verb")},
		{"noun binds", `{"term":"bank","cefr":"A1","partOfSpeech":"noun"}`, true, posPtr("noun")},
		{"absent binds nil", `{"term":"bank","cefr":"A1"}`, true, nil},
		{"null binds nil", `{"term":"bank","cefr":"A1","partOfSpeech":null}`, true, nil},
		{"unknown token rejected", `{"term":"bank","cefr":"A1","partOfSpeech":"banana"}`, false, nil},
		{"legacy free-text token rejected", `{"term":"bank","cefr":"A1","partOfSpeech":"adj"}`, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("POST", "/vocabulary-items", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			var got requests.CreateVocabularyItem
			err := c.ShouldBindJSON(&got)

			if !tt.wantOK {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantPOS == nil {
				assert.Nil(t, got.PartOfSpeech)
			} else {
				require.NotNil(t, got.PartOfSpeech)
				assert.Equal(t, *tt.wantPOS, *got.PartOfSpeech)
			}
		})
	}
}

// TestAddUserVocabularyWordBinding pins the same boundary on the add-word route.
func TestAddUserVocabularyWordBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		body   string
		wantOK bool
	}{
		{"member binds", `{"term":"bank","cefr":"A1","partOfSpeech":"idiom"}`, true},
		{"absent binds", `{"term":"bank","cefr":"A1"}`, true},
		{"unknown token rejected", `{"term":"bank","cefr":"A1","partOfSpeech":"banana"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("POST", "/user-vocabulary/words", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			var got requests.AddUserVocabularyWord
			err := c.ShouldBindJSON(&got)

			if tt.wantOK {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func posPtr(s string) *string { return &s }
