package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"engflex-api/config"
	"engflex-api/internal/server"
)

// TestVocabAdminDecksItemsRoutesRequireAuth guards against silently losing
// RequireAuth() on admin-only content routes (vocabulary, videos,
// scenarios/topics/personas, lessons/sections/activities — same curated-
// content pattern). The admin-vs-user 403 split for these, and the owner-
// vs-admin split on the shared /vocabulary-decks/:id PUT/DELETE, are
// covered by the service-layer table tests in each module's services/tests
// — a real Clerk token is needed to exercise the HTTP path end to end,
// which only the manual Swagger pass can provide.
func TestVocabAdminDecksItemsRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uuid := "00000000-0000-0000-0000-000000000000"
	tests := []struct{ method, path string }{
		{http.MethodPost, "/api/v1/admin/vocabulary-decks"},
		{http.MethodPost, "/api/v1/admin/vocabulary-items"},
		{http.MethodPut, "/api/v1/vocabulary-items/" + uuid},
		{http.MethodDelete, "/api/v1/vocabulary-items/" + uuid},
		{http.MethodPost, "/api/v1/vocabulary-categories"},
		{http.MethodPut, "/api/v1/vocabulary-categories/" + uuid},
		{http.MethodDelete, "/api/v1/vocabulary-categories/" + uuid},
		{http.MethodPost, "/api/v1/video-exercises"},
		{http.MethodPut, "/api/v1/video-exercises/" + uuid},
		{http.MethodDelete, "/api/v1/video-exercises/" + uuid},
		{http.MethodPost, "/api/v1/video-categories"},
		{http.MethodPut, "/api/v1/video-categories/" + uuid},
		{http.MethodDelete, "/api/v1/video-categories/" + uuid},
		{http.MethodPost, "/api/v1/video-transcripts"},
		{http.MethodPut, "/api/v1/video-transcripts/" + uuid},
		{http.MethodDelete, "/api/v1/video-transcripts/" + uuid},
		{http.MethodPost, "/api/v1/scenarios"},
		{http.MethodPut, "/api/v1/scenarios/" + uuid},
		{http.MethodDelete, "/api/v1/scenarios/" + uuid},
		{http.MethodPost, "/api/v1/scenario-topics"},
		{http.MethodPut, "/api/v1/scenario-topics/" + uuid},
		{http.MethodDelete, "/api/v1/scenario-topics/" + uuid},
		{http.MethodPost, "/api/v1/personas"},
		{http.MethodPut, "/api/v1/personas/" + uuid},
		{http.MethodDelete, "/api/v1/personas/" + uuid},
		{http.MethodPost, "/api/v1/lessons"},
		{http.MethodPut, "/api/v1/lessons/" + uuid},
		{http.MethodDelete, "/api/v1/lessons/" + uuid},
		{http.MethodPost, "/api/v1/lesson-sections"},
		{http.MethodPut, "/api/v1/lesson-sections/" + uuid},
		{http.MethodDelete, "/api/v1/lesson-sections/" + uuid},
		{http.MethodPost, "/api/v1/lesson-activities"},
		{http.MethodPut, "/api/v1/lesson-activities/" + uuid},
		{http.MethodDelete, "/api/v1/lesson-activities/" + uuid},
		{http.MethodGet, "/api/v1/admin/lesson-activities/" + uuid},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			r := server.NewRouter(nil, config.Config{})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}
