package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/modules/access/dtos/requests"
	"engflex-api/internal/modules/access/dtos/responses"
	"engflex-api/internal/modules/access/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// AdminAccessController serves the admin limits, usage, and trial-grant
// surfaces. The trial grant reuses TrialService.Start — the same single
// path as self-serve, idempotent by the user_trials primary key.
type AdminAccessController struct {
	admin  *services.AdminAccessService
	trials *services.TrialService
}

// NewAdminAccessController builds the controller over both services.
func NewAdminAccessController(admin *services.AdminAccessService, trials *services.TrialService) *AdminAccessController {
	return &AdminAccessController{admin: admin, trials: trials}
}

// ListLimits godoc
// @Summary      List the entitlement catalog (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=responses.AdminPlanLimits}
// @Failure      401 {object} common.ApiResponse
// @Failure      403 {object} common.ApiResponse
// @Router       /admin/premium/limits [get]
func (h *AdminAccessController) ListLimits(c *gin.Context) {
	rows, err := h.admin.ListLimits(c.Request.Context())
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "limits listed", responses.AdminPlanLimits(rows))
}

// UpdatePlanLimits godoc
// @Summary      Batch-upsert one plan's limits (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        planKey path string true "plan code (free, trial, 1m, 3m, 12m)"
// @Param        request body requests.UpdatePlanLimits true "limits payload"
// @Success      200 {object} common.ApiResponse
// @Failure      400 {object} common.ApiResponse
// @Failure      401 {object} common.ApiResponse
// @Failure      403 {object} common.ApiResponse
// @Failure      422 {object} common.ApiResponse
// @Router       /admin/premium/limits/{planKey} [put]
func (h *AdminAccessController) UpdatePlanLimits(c *gin.Context) {
	var req requests.UpdatePlanLimits
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	plan := enums.PlanKey(c.Param("planKey"))
	if err := h.admin.UpdatePlanLimits(c.Request.Context(), plan, req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "limits updated", nil)
}

// UserUsage godoc
// @Summary      Read a user's counters and recent events (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "user id"
// @Success      200 {object} common.ApiResponse{data=responses.UserUsage}
// @Failure      401 {object} common.ApiResponse
// @Failure      403 {object} common.ApiResponse
// @Router       /admin/users/{id}/usage [get]
func (h *AdminAccessController) UserUsage(c *gin.Context) {
	out, err := h.admin.UserUsage(c.Request.Context(), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "usage listed", out)
}

// GrantTrial godoc
// @Summary      Grant a trial to a user (admin, idempotent)
// @Tags         admin-premium
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "user id"
// @Success      200 {object} common.ApiResponse{data=responses.TrialStatus}
// @Success      201 {object} common.ApiResponse{data=responses.TrialStatus}
// @Failure      401 {object} common.ApiResponse
// @Failure      403 {object} common.ApiResponse
// @Router       /admin/users/{id}/trial [post]
func (h *AdminAccessController) GrantTrial(c *gin.Context) {
	started, err := h.trials.Start(c.Request.Context(), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	writeTrialStatus(c, started)
}

func writeTrialStatus(c *gin.Context, started bool) {
	if started {
		common.Created(c, "trial started", responses.TrialStatus{Started: true})
		return
	}
	common.OK(c, "trial already used", responses.TrialStatus{Started: false})
}

// ListActionCosts godoc
// @Summary      List the per-action credit prices (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=responses.AdminActionCosts}
// @Failure      401 {object} common.ApiResponse
// @Failure      403 {object} common.ApiResponse
// @Router       /admin/premium/action-costs [get]
func (h *AdminAccessController) ListActionCosts(c *gin.Context) {
	rows, err := h.admin.ListActionCosts(c.Request.Context())
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto []responses.AdminActionCost
	_ = utils.MapSlice(&dto, rows)
	common.OK(c, "action costs listed", responses.AdminActionCosts(dto))
}

// UpdateActionCosts godoc
// @Summary      Batch-upsert the per-action credit prices (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.UpdateActionCosts true "action costs payload"
// @Success      200 {object} common.ApiResponse
// @Failure      400 {object} common.ApiResponse
// @Failure      401 {object} common.ApiResponse
// @Failure      403 {object} common.ApiResponse
// @Failure      422 {object} common.ApiResponse
// @Router       /admin/premium/action-costs [put]
func (h *AdminAccessController) UpdateActionCosts(c *gin.Context) {
	var req requests.UpdateActionCosts
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	if err := h.admin.UpdateActionCosts(c.Request.Context(), middleware.UserID(c), req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "action costs updated", nil)
}

// GetTrialTerms godoc
// @Summary      Read the trial duration and caps (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=responses.AdminTrialTerms}
// @Failure      401 {object} common.ApiResponse
// @Failure      403 {object} common.ApiResponse
// @Router       /admin/premium/trial-terms [get]
func (h *AdminAccessController) GetTrialTerms(c *gin.Context) {
	out, err := h.admin.GetTrialTerms(c.Request.Context())
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "trial terms", out)
}

// UpdateTrialTerms godoc
// @Summary      Rewrite the trial duration and caps (admin)
// @Tags         admin-premium
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.UpdateTrialTerms true "trial terms payload"
// @Success      200 {object} common.ApiResponse
// @Failure      400 {object} common.ApiResponse
// @Failure      401 {object} common.ApiResponse
// @Failure      403 {object} common.ApiResponse
// @Router       /admin/premium/trial-terms [put]
func (h *AdminAccessController) UpdateTrialTerms(c *gin.Context) {
	var req requests.UpdateTrialTerms
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	if err := h.admin.UpdateTrialTerms(c.Request.Context(), middleware.UserID(c), req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "trial terms updated", nil)
}
