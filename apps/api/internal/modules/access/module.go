// Package access wires the entitlement module: the policy decision point,
// usage metering, /premium/me, trial start, and the admin limits + usage
// surfaces (Task 13), plus paid-AI orchestration (Task 12).
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
type Services struct {
	Policy       *accessservices.PolicyService
	Usage        *accessservices.UsageService
	Entitlements *accessservices.EntitlementsService
	Admin        *accessservices.AdminAccessService
	Trial        *accessservices.TrialService
	Pricing      *accessservices.PricingService
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
	creditCosts := repositories.NewActionCreditCostRepository(db)
	trialSettings := repositories.NewTrialSettingRepository(db)

	policy := accessservices.NewPolicyService(subscriptions, trials, counters, limits, creditCosts, trialSettings, cfg.Access)
	txManager := repositories.NewTransactionManager(db)
	usage := accessservices.NewUsageService(counters, events, txManager, policy)
	if db != nil {
		// Route-table tests boot the router with a nil DB and never serve
		// traffic; the snapshot and sweeper need a live handle.
		_ = policy.SnapshotCatalog(context.Background())
		usage.StartSweeper(context.Background())
	}
	return &Services{
		Policy:       policy,
		Usage:        usage,
		Entitlements: accessservices.NewEntitlementsService(policy, events),
		Admin:        accessservices.NewAdminAccessService(limits, counters, events, creditCosts, trialSettings, txManager, policy),
		Trial:        accessservices.NewTrialService(trials, policy),
		Pricing:      accessservices.NewPricingService(repositories.NewPlanRepository(db), policy),
		Enforce:      cfg.Access.Enforce,
	}
}

// RegisterRoutes mounts the entitlement endpoints under /api/v1.
func RegisterRoutes(rg *gin.RouterGroup, svc *Services) {
	ctl := controllers.NewEntitlementController(svc.Entitlements)
	trialCtl := controllers.NewTrialController(svc.Trial)
	adminCtl := controllers.NewAdminAccessController(svc.Admin, svc.Trial)
	premium := rg.Group("/premium")
	premium.GET("/me", middleware.RequireAuth(), ctl.GetMy)
	premium.GET("/usage", middleware.RequireAuth(), ctl.GetUsage)
	premium.POST("/trial", middleware.RequireAuth(), trialCtl.StartMy)
	rg.GET("/pricing", controllers.NewPricingController(svc.Pricing).Get)
	admin := rg.Group("/admin", middleware.RequireAuth(), middleware.RequireAdmin())
	admin.GET("/premium/limits", adminCtl.ListLimits)
	admin.PUT("/premium/limits/:planKey", adminCtl.UpdatePlanLimits)
	admin.GET("/premium/action-costs", adminCtl.ListActionCosts)
	admin.PUT("/premium/action-costs", adminCtl.UpdateActionCosts)
	admin.GET("/premium/trial-terms", adminCtl.GetTrialTerms)
	admin.PUT("/premium/trial-terms", adminCtl.UpdateTrialTerms)
	admin.GET("/users/:id/usage", adminCtl.UserUsage)
	admin.POST("/users/:id/trial", adminCtl.GrantTrial)
}
