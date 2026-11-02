// Package bookmarks registers routes over dependencies supplied by composition.
package bookmarks

import (
	"engflex-api/internal/modules/bookmarks/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, ctl *controllers.BookmarkController) {
	g := rg.Group("/bookmarks")
	g.GET("", middleware.RequireAuth(), ctl.List)
	g.POST("", middleware.RequireAuth(), ctl.Save)
	g.DELETE("/:contentType/:contentId", middleware.RequireAuth(), ctl.Unsave)
}
