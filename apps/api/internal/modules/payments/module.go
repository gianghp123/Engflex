// Package payments wires the VNPay checkout/IPN endpoints, the admin
// premium console (plan pricing, feature registry, all-user subscriptions
// view), and the subscription service other modules (e.g. users) pull
// premium status from.
package payments

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/config"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/payments/controllers"
	"engflex-api/internal/modules/payments/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts /payments/vnpay/* and /admin/premium/* under
// /api/v1. Checkout requires auth (learner-initiated); ipn/return/plans are
// public — ipn's security is the HMAC signature alone (VNPay calls it
// server-to-server, no Clerk token exists there), and return is a UX-only
// status page. Every /admin/premium/* route is admin-gated outright.
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, cfg config.PaymentConfig) {
	subRepo := repositories.NewSubscriptionRepository(db)
	txnRepo := repositories.NewPaymentTransactionRepository(db)
	planRepo := repositories.NewPlanRepository(db)
	featureRepo := repositories.NewPremiumFeatureRepository(db)

	subSvc := services.NewSubscriptionService(subRepo, planRepo)
	planSvc := services.NewPlanService(planRepo)
	featureSvc := services.NewPremiumFeatureService(featureRepo)
	vnpaySvc := services.NewVNPayService(services.VNPayConfig{
		TmnCode:    cfg.VNPayTmnCode,
		HashSecret: cfg.VNPayHashSecret,
		PayURL:     cfg.VNPayPayURL,
		ReturnURL:  cfg.VNPayReturnURL,
	}, db, subRepo, txnRepo, planRepo)

	paymentCtl := controllers.NewPaymentController(planSvc, subSvc, vnpaySvc)
	planCtl := controllers.NewPlanController(planSvc)
	featureCtl := controllers.NewPremiumFeatureController(featureSvc)
	subAdminCtl := controllers.NewSubscriptionAdminController(subSvc)

	g := rg.Group("/payments/vnpay")
	g.GET("/plans", paymentCtl.ListPlans)
	g.POST("/checkout", middleware.RequireAuth(), paymentCtl.Checkout)
	g.GET("/ipn", paymentCtl.IPN)
	g.GET("/return", paymentCtl.Return)

	rg.GET("/premium/me", middleware.RequireAuth(), paymentCtl.Me)

	admin := rg.Group("/admin/premium", middleware.RequireAuth(), middleware.RequireAdmin())
	admin.GET("/plans", planCtl.List)
	admin.PUT("/plans/:plan", planCtl.Update)
	admin.GET("/features", featureCtl.List)
	admin.POST("/features", featureCtl.Create)
	admin.PUT("/features/:id", featureCtl.Update)
	admin.DELETE("/features/:id", featureCtl.Delete)
	admin.GET("/subscriptions", subAdminCtl.List)
}
