// Package users wires the admin-only Clerk user directory plus subscription
// management. There is no local users table — Clerk is the source of
// truth for account/role data — so this module talks to the Clerk Backend
// API for that, and to the subscriptions table (via the payments module's
// SubscriptionService) for premium status and manual grants.
package users

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/internal/database/repositories"
	paymentservices "engflex-api/internal/modules/payments/services"
	"engflex-api/internal/modules/users/controllers"
	"engflex-api/internal/modules/users/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts the user-directory endpoints under /api/v1/admin.
// All routes are admin-only: this is account/role/subscription management,
// never a learner-facing surface. SubscriptionService is wired here (not
// inside the payments module) so this module stays the single owner of the
// /admin/users route group — the collision this branch closes was two
// modules both registering that group.
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	subSvc := paymentservices.NewSubscriptionService(
		repositories.NewSubscriptionRepository(db), repositories.NewPlanRepository(db))
	svc := services.NewUserService(subSvc)
	ctl := controllers.NewUserController(svc, subSvc)

	admin := rg.Group("/admin/users")
	admin.GET("", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.List)
	admin.PATCH("/:id/role", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.UpdateRole)
	admin.GET("/:id/subscriptions", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.ListSubscriptions)
	admin.POST("/:id/subscriptions", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.GrantSubscription)
	admin.DELETE("/:id/subscriptions", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.RevokeSubscription)
}
