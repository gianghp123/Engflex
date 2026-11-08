package bookmarks

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/bookmarks/controllers"
	"engflex-api/internal/modules/bookmarks/services"
	paymentservices "engflex-api/internal/modules/payments/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts the bookmarks endpoints under /api/v1.
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	repo := repositories.NewBookmarkRepository(db)
	premium := paymentservices.NewSubscriptionService(repositories.NewSubscriptionRepository(db), repositories.NewPlanRepository(db))
	svc := services.NewBookmarkService(repo, repositories.NewSavedItemRepository(db), premium)
	ctl := controllers.NewBookmarkController(svc)

	g := rg.Group("/bookmarks")
	g.GET("", middleware.RequireAuth(), ctl.List)
	g.GET("/library", middleware.RequireAuth(), ctl.Library)
	g.POST("", middleware.RequireAuth(), ctl.Save)
	g.PATCH("/:contentType/:contentId/note", middleware.RequireAuth(), ctl.UpdateNote)
	g.DELETE("/:contentType/:contentId", middleware.RequireAuth(), ctl.Unsave)
}
