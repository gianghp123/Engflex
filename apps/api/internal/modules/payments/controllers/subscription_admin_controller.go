package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/modules/payments/services"
	"engflex-api/internal/utils"
)

// SubscriptionAdminController exposes the admin-wide subscriptions view
// (every user), distinct from the per-user history already on
// /admin/users/:id/subscriptions.
type SubscriptionAdminController struct {
	service services.SubscriptionService
}

// NewSubscriptionAdminController wires the controller to the service.
func NewSubscriptionAdminController(service services.SubscriptionService) *SubscriptionAdminController {
	return &SubscriptionAdminController{service: service}
}

// List godoc
// @Summary      List subscriptions across every user (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Produce      json
// @Param        status query string false "pending|active|expired|cancelled|failed"
// @Param        plan query string false "1m|3m|12m"
// @Param        page query int false "page"
// @Param        pageSize query int false "page size"
// @Success      200 {object} common.ApiResponse{data=[]responses.Subscription}
// @Router       /admin/premium/subscriptions [get]
func (h *SubscriptionAdminController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	offset := utils.Offset(page, pageSize)
	status := enums.SubscriptionStatus(c.Query("status"))
	plan := enums.SubscriptionPlan(c.Query("plan"))
	items, total, err := h.service.ListAll(c.Request.Context(), status, plan, pageSize, offset)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Paginated(c, "subscriptions listed", items, page, pageSize, total)
}
