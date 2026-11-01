package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	paymentrequests "engflex-api/internal/modules/payments/dtos/requests"
	paymentservices "engflex-api/internal/modules/payments/services"
	"engflex-api/internal/modules/users/dtos/requests"
	"engflex-api/internal/modules/users/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// UserController exposes the Clerk-backed admin user directory, plus the
// admin-only subscription history/grant endpoints — delegated straight to
// SubscriptionService, not wrapped by UserService.
type UserController struct {
	service services.UserService
	subs    paymentservices.SubscriptionService
}

// NewUserController wires the controller to its services.
func NewUserController(service services.UserService, subs paymentservices.SubscriptionService) *UserController {
	return &UserController{service: service, subs: subs}
}

// List godoc
// @Summary      List Clerk users (admin)
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "page"
// @Param        pageSize query int false "page size"
// @Param        search query string false "name/email/username search"
// @Success      200 {object} common.ApiResponse{data=[]responses.User}
// @Router       /admin/users [get]
func (h *UserController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	var req requests.ListUsers
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	req.Page, req.PageSize = page, pageSize
	items, total, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Paginated(c, "users", items, page, pageSize, total)
}

// UpdateRole godoc
// @Summary      Change a user's role (admin)
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Clerk user id"
// @Param        request body requests.UpdateUserRole true "role payload"
// @Success      200 {object} common.ApiResponse{data=responses.User}
// @Router       /admin/users/{id}/role [patch]
func (h *UserController) UpdateRole(c *gin.Context) {
	var req requests.UpdateUserRole
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	updated, err := h.service.UpdateRole(c.Request.Context(), middleware.UserID(c), c.Param("id"), req.Role)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "user role updated", updated)
}

// ListSubscriptions godoc
// @Summary      List a user's subscription history (admin)
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Clerk user id"
// @Param        page query int false "page"
// @Param        pageSize query int false "page size"
// @Success      200 {object} common.ApiResponse{data=[]responses.Subscription}
// @Router       /admin/users/{id}/subscriptions [get]
func (h *UserController) ListSubscriptions(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	items, total, err := h.subs.ListByUserID(c.Request.Context(), c.Param("id"), pageSize, (page-1)*pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Paginated(c, "subscriptions", items, page, pageSize, total)
}

// GrantSubscription godoc
// @Summary      Manually grant a premium subscription (admin)
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Clerk user id"
// @Param        request body requests.GrantSubscription true "plan to grant"
// @Success      200 {object} common.ApiResponse{data=responses.Subscription}
// @Router       /admin/users/{id}/subscriptions [post]
func (h *UserController) GrantSubscription(c *gin.Context) {
	var req paymentrequests.GrantSubscription
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	sub, err := h.subs.Grant(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "subscription granted", sub)
}

// RevokeSubscription godoc
// @Summary      Revoke a user's active premium subscription (admin)
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Clerk user id"
// @Success      200 {object} common.ApiResponse
// @Router       /admin/users/{id}/subscriptions [delete]
func (h *UserController) RevokeSubscription(c *gin.Context) {
	if err := h.subs.Revoke(c.Request.Context(), c.Param("id")); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "subscription revoked", nil)
}
