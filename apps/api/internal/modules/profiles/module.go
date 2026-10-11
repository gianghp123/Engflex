// Package profiles wires the admin user-management surface: a read-only
// list/detail view joining Clerk identity with local onboarding profile and
// premium (subscription) state. Account mutation (ban/lock) is not exposed
// here — that stays a manual Clerk Dashboard action.
package profiles

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/internal/clerkusers"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/profiles/controllers"
	"engflex-api/internal/modules/profiles/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts the profiles/admin endpoints under /api/v1.
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	subSvc := services.NewSubscriptionService(repositories.NewSubscriptionRepository(db))
	adminSvc := services.NewAdminUserService(clerkusers.NewHTTPClient(), repositories.NewUserProfileRepository(db), subSvc)
	adminCtl := controllers.NewAdminUserController(adminSvc, subSvc)

	au := rg.Group("/admin/users")
	au.GET("", middleware.RequireAuth(), middleware.RequireAdmin(), adminCtl.List)
	au.GET("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), adminCtl.GetByID)
	au.POST("/:id/subscriptions", middleware.RequireAuth(), middleware.RequireAdmin(), adminCtl.GrantSubscription)
}
