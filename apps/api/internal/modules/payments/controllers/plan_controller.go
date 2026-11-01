package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/modules/payments/dtos/requests"
	"engflex-api/internal/modules/payments/services"
)

// PlanController exposes admin CRUD over the three fixed-duration
// packages' prices.
type PlanController struct {
	service services.PlanService
}

// NewPlanController wires the controller to the service.
func NewPlanController(service services.PlanService) *PlanController {
	return &PlanController{service: service}
}

// List godoc
// @Summary      List plans, including inactive ones (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=[]responses.Plan}
// @Router       /admin/premium/plans [get]
func (h *PlanController) List(c *gin.Context) {
	items, err := h.service.List(c.Request.Context())
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "plans listed", items)
}

// Update godoc
// @Summary      Edit a plan's price (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        plan path string true "plan code (1m, 3m, 12m)"
// @Param        request body requests.UpdatePlan true "price payload"
// @Success      200 {object} common.ApiResponse{data=responses.Plan}
// @Router       /admin/premium/plans/{plan} [put]
func (h *PlanController) Update(c *gin.Context) {
	var req requests.UpdatePlan
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	plan := enums.SubscriptionPlan(c.Param("plan"))
	updated, err := h.service.Update(c.Request.Context(), plan, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "plan updated", updated)
}
