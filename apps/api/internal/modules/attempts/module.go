// Package attempts registers routes over dependencies supplied by composition.
package attempts

import (
	"engflex-api/internal/modules/attempts/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, ctl *controllers.ProgressController) {
	rg.GET("/progress", middleware.RequireAuth(), ctl.Get)
}
