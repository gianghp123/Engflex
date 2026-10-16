// Package users wires the admin-only Clerk user directory. There is no
// local users table — Clerk is the source of truth — so this module talks
// to the Clerk Backend API directly instead of a repository.
package users

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/modules/users/controllers"
	"engflex-api/internal/modules/users/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts the user-directory endpoints under /api/v1/admin.
// Both routes are admin-only: this is account/role management, never a
// learner-facing surface.
func RegisterRoutes(rg *gin.RouterGroup) {
	svc := services.NewUserService()
	ctl := controllers.NewUserController(svc)

	admin := rg.Group("/admin/users")
	admin.GET("", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.List)
	admin.PATCH("/:id/role", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.UpdateRole)
}
