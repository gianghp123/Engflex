package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/attempts/dtos/responses"
	"engflex-api/internal/modules/attempts/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// ProgressController exposes the measurable-only progress rollup over HTTP.
type ProgressController struct {
	service *services.ProgressService
}

// NewProgressController wires the controller to the service.
func NewProgressController(service *services.ProgressService) *ProgressController {
	return &ProgressController{service: service}
}

// Get godoc
// @Summary      Get the caller's progress rollup
// @Tags         progress
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=responses.Progress}
// @Failure      401 {object} common.ApiResponse
// @Failure      500 {object} common.ApiResponse
// @Router       /progress [get]
func (h *ProgressController) Get(c *gin.Context) {
	userID := middleware.UserID(c)
	progress, err := h.service.GetProgress(c.Request.Context(), userID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Progress
	_ = utils.Map(&dto, progress)
	common.OK(c, "progress", dto)
}
