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

// TestLessonAudioSynthesizeRouteRequiresAuth guards against silently
// losing RequireAuth() while wiring RequireAdmin() on the lesson-activity
// audio synthesis route. The admin-vs-user 403 split is already covered
// generically by TestRequireAdmin in middleware/role_test.go; a real
// Clerk-signed token is needed to exercise that distinction end to end,
// which only the manual Swagger pass can provide.
func TestLessonAudioSynthesizeRouteRequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := server.NewRouter(nil, config.Config{})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/lesson-activities/some-id/synthesize-audio", nil))
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
