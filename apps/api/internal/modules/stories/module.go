// Package stories registers routes over dependencies supplied by composition.
package stories

import (
	"engflex-api/internal/modules/stories/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, ctl *controllers.StoryController) {
	g := rg.Group("/stories")
	g.GET("", middleware.RequireAuth(), ctl.List)
	g.GET("/topics", middleware.RequireAuth(), ctl.ListTopics)
	g.GET("/:id", middleware.RequireAuth(), ctl.GetByID)
	g.GET("/by-slug/:slug", middleware.RequireAuth(), ctl.GetBySlug)
}
