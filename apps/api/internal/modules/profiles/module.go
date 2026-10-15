// Package profiles wires the profiles-domain endpoints under /api/v1.
// Onboarding completion lives here; the progress rollup reads the same table
// through the shared repository.
package profiles

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/profiles/controllers"
	"engflex-api/internal/modules/profiles/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts the caller's profile endpoints.
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	svc := services.NewProfileService(repositories.NewUserProfileRepository(db))
	ctl := controllers.NewProfileController(svc)
	g := rg.Group("/profiles")
	g.GET("/me", middleware.RequireAuth(), ctl.GetMe)
	g.PUT("/me", middleware.RequireAuth(), ctl.UpdateMe)
}
