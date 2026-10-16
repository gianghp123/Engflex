// Package attempts wires the attempts-domain endpoints under /api/v1. The
// attempt lifecycle itself is driven by the lessons module; progress is the
// first endpoint owned here.
package attempts

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/attempts/controllers"
	"engflex-api/internal/modules/attempts/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts the progress endpoint.
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	svc := services.NewProgressService(
		repositories.NewAttemptRepository(db),
		repositories.NewConversationRepository(db),
		repositories.NewUserProfileRepository(db),
	)
	ctl := controllers.NewProgressController(svc)
	rg.GET("/progress", middleware.RequireAuth(), ctl.Get)
}
