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

// TestAdminUserRoutesRequireAuth guards against silently losing
// RequireAuth() while wiring RequireAdmin() on the admin user endpoints. The
// admin-vs-user 403 split is already covered generically by TestRequireAdmin
// in middleware/role_test.go; a real Clerk-signed token is needed to
// exercise that distinction end to end, which only the manual Swagger pass
// can provide.
func TestAdminUserRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct{ method, path string }{
		{http.MethodGet, "/api/v1/admin/users"},
		{http.MethodGet, "/api/v1/admin/users/some-user-id"},
		{http.MethodPost, "/api/v1/admin/users/some-user-id/subscriptions"},
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
