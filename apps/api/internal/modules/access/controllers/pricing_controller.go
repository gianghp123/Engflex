package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/access/services"
)

// PricingController serves the public pricing catalog.
type PricingController struct {
	pricing *services.PricingService
}

// NewPricingController builds the controller over the pricing service.
func NewPricingController(pricing *services.PricingService) *PricingController {
	return &PricingController{pricing: pricing}
}

// Get godoc
// @Summary      Public pricing catalog
// @Description  Active plans with monthly limits, free entitlements, trial terms, and per-action costs
// @Tags         pricing
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=responses.PricingCatalog}
// @Failure      503 {object} common.ApiResponse
// @Router       /pricing [get]
func (h *PricingController) Get(c *gin.Context) {
	out, err := h.pricing.Get(c.Request.Context())
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.Header("Cache-Control", "public, max-age=30")
	common.OK(c, "ok", out)
}
