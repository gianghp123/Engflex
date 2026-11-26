// Package access wires the entitlement module: the policy decision point,
// usage metering, /premium/me, and (Task 12/13) paid-AI orchestration.
package access

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/config"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/access/controllers"
	accessservices "engflex-api/internal/modules/access/services"
	"engflex-api/internal/server/middleware"
)

// Services bundles the singletons the router hands to other modules.
// Admin and Trial join this bundle in Task 13.
type Services struct {
	Policy       *accessservices.PolicyService
	Usage        *accessservices.UsageService
	Entitlements *accessservices.EntitlementsService
	// Enforce mirrors cfg.Access.Enforce so consuming modules (exam AI)
	// never re-read config: one flag, one source.
	Enforce bool
}

// NewServices builds the access singletons: one policy snapshot cache and
// one sweeper per process, shared by every consuming module.
func NewServices(db *gorm.DB, cfg config.Config) *Services {
	subscriptions := repositories.NewSubscriptionRepository(db)
	trials := repositories.NewUserTrialRepository(db)
	counters := repositories.NewUsageCounterRepository(db)
	limits := repositories.NewPlanFeatureLimitRepository(db)
	events := repositories.NewUsageEventRepository(db)

	policy := accessservices.NewPolicyService(subscriptions, trials, counters, limits, cfg.Access)
	usage := accessservices.NewUsageService(counters, events, repositories.NewTransactionManager(db), policy)
	if db != nil {
		// Route-table tests boot the router with a nil DB and never serve
		// traffic; the snapshot and sweeper need a live handle.
		_ = policy.SnapshotLimits(context.Background())
		usage.StartSweeper(context.Background())
	}
	return &Services{Policy: policy, Usage: usage, Entitlements: accessservices.NewEntitlementsService(policy), Enforce: cfg.Access.Enforce}
}

// RegisterRoutes mounts the entitlement endpoints under /api/v1.
// Task 13 adds /premium/trial and the /admin/premium/limits group here.
func RegisterRoutes(rg *gin.RouterGroup, svc *Services) {
	ctl := controllers.NewEntitlementController(svc.Entitlements)
	premium := rg.Group("/premium")
	premium.GET("/me", middleware.RequireAuth(), ctl.GetMy)
}
