package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/profiles/dtos/requests"
	"engflex-api/internal/modules/profiles/dtos/responses"
	"engflex-api/internal/modules/profiles/services"
	"engflex-api/internal/utils"
)

// AdminUserController exposes the admin, read-only view over user accounts
// plus manual premium grants (no payment flow exists yet, so this is the
// only way to put a subscription row on a test or support account).
type AdminUserController struct {
	service       services.AdminUserService
	subscriptions services.SubscriptionService
}

// NewAdminUserController wires the controller to the services.
func NewAdminUserController(service services.AdminUserService, subscriptions services.SubscriptionService) *AdminUserController {
	return &AdminUserController{service: service, subscriptions: subscriptions}
}

// List godoc
// @Summary      List users (admin)
// @Tags         admin-users
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=[]responses.AdminUserSummary}
// @Router       /admin/users [get]
func (h *AdminUserController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	items, total, err := h.service.List(c.Request.Context(), pageSize, utils.Offset(page, pageSize))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Paginated(c, "users", items, page, pageSize, total)
}

// GetByID godoc
// @Summary      Get user detail (admin)
// @Tags         admin-users
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Clerk user id"
// @Success      200 {object} common.ApiResponse{data=responses.AdminUserDetail}
// @Router       /admin/users/{id} [get]
func (h *AdminUserController) GetByID(c *gin.Context) {
	detail, err := h.service.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "user", detail)
}

// GrantSubscription godoc
// @Summary      Manually grant premium to a user (admin)
// @Tags         admin-users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Clerk user id"
// @Param        request body requests.GrantSubscription true "grant payload"
// @Success      201 {object} common.ApiResponse{data=responses.Subscription}
// @Router       /admin/users/{id}/subscriptions [post]
func (h *AdminUserController) GrantSubscription(c *gin.Context) {
	var req requests.GrantSubscription
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.subscriptions.Grant(c.Request.Context(), c.Param("id"), req.Plan, req.ExpiresAt)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Subscription
	_ = utils.Map(&dto, m)
	common.Created(c, "premium granted", dto)
}
