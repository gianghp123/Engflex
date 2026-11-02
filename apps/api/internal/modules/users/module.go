// Package users registers routes over dependencies supplied by composition.
package users

import (
	"engflex-api/internal/modules/users/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, ctl *controllers.UserController) {
	admin := rg.Group("/admin/users")
	admin.GET("", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.List)
	admin.PATCH("/:id/role", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.UpdateRole)
	admin.GET("/:id/subscriptions", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.ListSubscriptions)
	admin.POST("/:id/subscriptions", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.GrantSubscription)
	admin.DELETE("/:id/subscriptions", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.RevokeSubscription)
	admin.GET("/:id/attempts", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.ListAttempts)
}
