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

// TestProfilesRoutesRequireAuth pins that both profile routes are mounted and
// stay behind RequireAuth: an unauthenticated caller is rejected before the
// handler ever runs. The DB is nil because RegisterRoutes only stores the
// handle and the 401 is decided before any query.
func TestProfilesRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct{ method, path string }{
		{http.MethodGet, "/api/v1/profiles/me"},
		{http.MethodPut, "/api/v1/profiles/me"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			cfg := config.Config{Cors: config.CorsConfig{AllowedOrigins: []string{"http://localhost:3000"}}}
			r := server.NewRouter(nil, cfg)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}
