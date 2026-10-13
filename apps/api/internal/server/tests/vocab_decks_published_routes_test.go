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
// RequireAuth() on the new admin-only vocabulary routes (system deck
// creation, system entry creation, dictionary entry update/delete). The
// admin-vs-user 403 split for these, and the owner-vs-admin split on the
// shared /vocabulary-decks/:id PUT/DELETE, are covered by the service-layer
// table tests in modules/vocabulary/services/tests — a real Clerk token is
// needed to exercise the HTTP path end to end, which only the manual
// Swagger pass can provide.
func TestVocabAdminDecksItemsRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uuid := "00000000-0000-0000-0000-000000000000"
	tests := []struct{ method, path string }{
		{http.MethodPost, "/api/v1/admin/vocabulary-decks"},
		{http.MethodPost, "/api/v1/admin/vocabulary-items"},
		{http.MethodPut, "/api/v1/vocabulary-items/" + uuid},
		{http.MethodDelete, "/api/v1/vocabulary-items/" + uuid},
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
