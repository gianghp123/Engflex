package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/payments/dtos/requests"
	"engflex-api/internal/modules/payments/services"
	"engflex-api/internal/utils"
)

// PremiumFeatureController exposes admin CRUD over the premium-gating
// feature registry.
type PremiumFeatureController struct {
	service services.PremiumFeatureService
}

// NewPremiumFeatureController wires the controller to the service.
func NewPremiumFeatureController(service services.PremiumFeatureService) *PremiumFeatureController {
	return &PremiumFeatureController{service: service}
}

// List godoc
// @Summary      List premium features (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=[]responses.PremiumFeature}
// @Router       /admin/premium/features [get]
func (h *PremiumFeatureController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	offset := utils.Offset(page, pageSize)
	items, total, err := h.service.List(c.Request.Context(), pageSize, offset)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Paginated(c, "premium features listed", items, page, pageSize, total)
}

// Create godoc
// @Summary      Create a premium feature entry (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreatePremiumFeature true "feature payload"
// @Success      201 {object} common.ApiResponse{data=responses.PremiumFeature}
// @Router       /admin/premium/features [post]
func (h *PremiumFeatureController) Create(c *gin.Context) {
	var req requests.CreatePremiumFeature
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	created, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Created(c, "premium feature created", created)
}

// Update godoc
// @Summary      Edit a premium feature entry (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "feature id"
// @Param        request body requests.UpdatePremiumFeature true "feature payload"
// @Success      200 {object} common.ApiResponse{data=responses.PremiumFeature}
// @Router       /admin/premium/features/{id} [put]
func (h *PremiumFeatureController) Update(c *gin.Context) {
	var req requests.UpdatePremiumFeature
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	updated, err := h.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "premium feature updated", updated)
}

// Delete godoc
// @Summary      Delete a premium feature entry (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "feature id"
// @Success      200 {object} common.ApiResponse
// @Router       /admin/premium/features/{id} [delete]
func (h *PremiumFeatureController) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "premium feature deleted", nil)
}
