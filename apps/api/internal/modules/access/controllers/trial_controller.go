package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/access/services"
	"engflex-api/internal/server/middleware"
)

// TrialController serves the self-serve trial start.
type TrialController struct {
	trials *services.TrialService
}

// NewTrialController builds the controller over the trial service.
func NewTrialController(trials *services.TrialService) *TrialController {
	return &TrialController{trials: trials}
}

// StartMy godoc
// @Summary      Start my trial (self-serve, idempotent)
// @Tags         premium
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=responses.TrialStatus}
// @Success      201 {object} common.ApiResponse{data=responses.TrialStatus}
// @Failure      401 {object} common.ApiResponse
// @Router       /premium/trial [post]
func (h *TrialController) StartMy(c *gin.Context) {
	started, err := h.trials.Start(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	writeTrialStatus(c, started)
}
