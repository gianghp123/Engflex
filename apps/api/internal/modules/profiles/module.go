// Package profiles registers routes over dependencies supplied by composition.
package profiles

import (
	"engflex-api/internal/modules/profiles/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, ctl *controllers.ProfileController) {
	g := rg.Group("/profiles")
	g.GET("/me", middleware.RequireAuth(), ctl.GetMe)
	g.PUT("/me", middleware.RequireAuth(), ctl.UpdateMe)
}
