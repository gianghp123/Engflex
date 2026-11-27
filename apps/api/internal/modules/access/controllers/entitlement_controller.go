// Package controllers holds the access module's HTTP handlers.
package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/access/dtos/responses"
	"engflex-api/internal/modules/access/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// EntitlementController serves the caller's entitlement view.
type EntitlementController struct {
	entitlements *services.EntitlementsService
}

// NewEntitlementController builds the controller over the view service.
func NewEntitlementController(entitlements *services.EntitlementsService) *EntitlementController {
	return &EntitlementController{entitlements: entitlements}
}

// GetMy godoc
// @Summary      My entitlements
// @Description  Plan, credit and voice usage, trial state, and per-action costs
// @Tags         premium
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=responses.MyEntitlements}
// @Failure      401 {object} common.ApiResponse
// @Failure      503 {object} common.ApiResponse
// @Router       /premium/me [get]
func (h *EntitlementController) GetMy(c *gin.Context) {
	userID := middleware.UserID(c)
	// Typed so the responses import resolves for swagger's type lookup.
	var dto *responses.MyEntitlements
	dto, err := h.entitlements.MyEntitlements(c.Request.Context(), userID, middleware.IsAdmin(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "ok", dto)
}

// GetUsage godoc
// @Summary      My usage history
// @Description  Plan, counter usage, per-action costs, and the caller's settled charge history (committed + reserved, newest first)
// @Tags         premium
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "page number"
// @Param        pageSize query int false "page size"
// @Success      200 {object} common.ApiResponse{data=responses.MyUsage}
// @Failure      401 {object} common.ApiResponse
// @Failure      503 {object} common.ApiResponse
// @Router       /premium/usage [get]
func (h *EntitlementController) GetUsage(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	// Typed so the responses import resolves for swagger's type lookup.
	var dto *responses.MyUsage
	dto, err := h.entitlements.MyUsage(c.Request.Context(), middleware.UserID(c), middleware.IsAdmin(c), page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "ok", dto)
}
