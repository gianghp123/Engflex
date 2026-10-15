package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	repomocks "engflex-api/internal/database/repositories/mocks"
	"engflex-api/internal/modules/profiles/controllers"
	"engflex-api/internal/modules/profiles/services"
)

const controllerUserID = "user_controller_1"

// newRouter mounts the profile controller with a stub auth middleware. The
// string "userID" must match middleware.userIDKey (unexported), which is the
// gin context key production's RequireAuth stamps.
func newRouter(repo *repomocks.MockUserProfileRepository) *gin.Engine {
	svc := services.NewProfileService(repo)
	ctl := controllers.NewProfileController(svc)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", controllerUserID)
		c.Next()
	})
	r.GET("/profiles/me", ctl.GetMe)
	r.PUT("/profiles/me", ctl.UpdateMe)
	return r
}

// TestUpdateMeRejectsInvalidBody pins Review Focus #3 at the HTTP layer: a body
// that fails binding returns 400 before the service (and thus the DB) is
// reached. The mock has no expectations, so any repository call would fail the
// mock — a 400 with no panic proves nothing downstream ran.
func TestUpdateMeRejectsInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := repomocks.NewMockUserProfileRepository(t)

	r := newRouter(repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/profiles/me", strings.NewReader(`{"topics":["travel"]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "GetByUserID", mock.Anything, mock.Anything)
}

// TestUpdateMeRejectsEmptyGoals pins the goals binding (omitempty,min=1): a
// provided-but-empty goals list is rejected before the service is reached.
func TestUpdateMeRejectsEmptyGoals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := repomocks.NewMockUserProfileRepository(t)

	r := newRouter(repo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/profiles/me", strings.NewReader(`{"goals":[]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	repo.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "GetByUserID", mock.Anything, mock.Anything)
}

// TestGetMeNoProfileReturns404 pins Review Focus #5 at the HTTP layer: a
// signed-in caller with no profile row gets 404, which the web guard reads as
// "not onboarded".
func TestGetMeNoProfileReturns404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := repomocks.NewMockUserProfileRepository(t)
	repo.On("GetByUserID", mock.Anything, controllerUserID).Return(nil, gorm.ErrRecordNotFound)

	r := newRouter(repo)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/profiles/me", nil))

	assert.Equal(t, http.StatusNotFound, w.Code)
}
