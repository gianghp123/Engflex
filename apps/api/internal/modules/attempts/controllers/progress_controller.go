package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/attempts/dtos/responses"
	"engflex-api/internal/modules/attempts/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// ProgressController exposes the measurable-only progress rollup plus the
// display-only mistake aggregation over HTTP.
type ProgressController struct {
	service *services.ProgressService
	errors  *services.ErrorStatsService
}

// NewProgressController wires the controller to the service.
func NewProgressController(service *services.ProgressService, errors *services.ErrorStatsService) *ProgressController {
	return &ProgressController{service: service, errors: errors}
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

// Errors godoc
// @Summary      Get the caller's top error types
// @Tags         progress
// @Security     BearerAuth
// @Produce      json
// @Param        limit query int false "max buckets"
// @Success      200 {object} common.ApiResponse{data=[]services.ErrorStat}
// @Failure      500 {object} common.ApiResponse
// @Router       /progress/errors [get]
func (h *ProgressController) Errors(c *gin.Context) {
	stats, err := h.errors.TopErrors(c.Request.Context(), middleware.UserID(c), queryLimit(c, "limit", 5))
	if err != nil {
		common.Fail(c, err)
		return
	}
	if stats == nil {
		stats = []services.ErrorStat{}
	}
	common.OK(c, "top errors", stats)
}

// Mistakes godoc
// @Summary      List the caller's recent mistakes
// @Tags         progress
// @Security     BearerAuth
// @Produce      json
// @Param        limit query int false "max rows"
// @Param        errorType query string false "error type filter"
// @Success      200 {object} common.ApiResponse{data=[]services.Mistake}
// @Failure      500 {object} common.ApiResponse
// @Router       /progress/mistakes [get]
func (h *ProgressController) Mistakes(c *gin.Context) {
	rows, err := h.errors.Mistakes(c.Request.Context(), middleware.UserID(c), c.Query("errorType"), queryLimit(c, "limit", 20))
	if err != nil {
		common.Fail(c, err)
		return
	}
	if rows == nil {
		rows = []services.Mistake{}
	}
	common.OK(c, "mistakes", rows)
}

// queryLimit parses a positive query int, falling back to def on any
// missing/invalid value.
func queryLimit(c *gin.Context, key string, def int) int {
	if raw := c.Query(key); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			return n
		}
	}
	return def
}
