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

// TestVideoAdminRoutesRequireAuth guards against silently losing
// RequireAuth() while wiring RequireAdmin() on the video admin routes: an
// unauthenticated caller must still be rejected before role is ever
// evaluated. The admin-vs-user 403 split is already covered generically by
// TestRequireAdmin in middleware/role_test.go; a real Clerk-signed token is
// needed to exercise that distinction end to end, which only the manual
// Swagger pass can provide.
func TestVideoAdminRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uuid := "00000000-0000-0000-0000-000000000000"
	tests := []struct{ method, path string }{
		{http.MethodPost, "/api/v1/video-exercises"},
		{http.MethodPut, "/api/v1/video-exercises/" + uuid},
		{http.MethodDelete, "/api/v1/video-exercises/" + uuid},
		{http.MethodPost, "/api/v1/video-transcripts"},
		{http.MethodPut, "/api/v1/video-transcripts/" + uuid},
		{http.MethodDelete, "/api/v1/video-transcripts/" + uuid},
		{http.MethodPost, "/api/v1/video-categories"},
		{http.MethodPut, "/api/v1/video-categories/" + uuid},
		{http.MethodDelete, "/api/v1/video-categories/" + uuid},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			// Same CORS-construction rule as router_test.go: gin-contrib/cors
			// panics on an empty allowlist, so pass one origin; these routes
			// carry no CORS assertions (401 is decided before CORS matters).
			cfg := config.Config{Cors: config.CorsConfig{AllowedOrigins: []string{"http://localhost:3000"}}}
			r := server.NewRouter(nil, cfg)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}
