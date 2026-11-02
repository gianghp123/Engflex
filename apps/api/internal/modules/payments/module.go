// Package payments registers routes over dependencies supplied by composition.
package payments

import (
	"engflex-api/internal/modules/payments/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, paymentCtl *controllers.PaymentController, planCtl *controllers.PlanController, featureCtl *controllers.PremiumFeatureController, subAdminCtl *controllers.SubscriptionAdminController) {
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
