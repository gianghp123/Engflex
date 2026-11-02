package controllers

import (
	"github.com/gin-gonic/gin"

	"context"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/users/dtos/requests"
	"engflex-api/internal/modules/users/dtos/responses"
	"engflex-api/internal/modules/users/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// UserController exposes the Clerk-backed admin user directory, plus the
// admin-only subscription history/grant endpoints — delegated straight to
// SubscriptionService, not wrapped by UserService.
type SubscriptionAdministration interface {
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.Subscription, int64, error)
	GrantUser(ctx context.Context, userID string, plan enums.SubscriptionPlan, note string) (*models.Subscription, error)
	Revoke(ctx context.Context, userID string) error
}

type UserController struct {
	service services.UserService
	subs    SubscriptionAdministration
}

// NewUserController wires the controller to its services.
func NewUserController(service services.UserService, subs SubscriptionAdministration) *UserController {
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
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	req.Page, req.PageSize = page, pageSize
	items, total, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto []*responses.User
	_ = utils.MapSlice(&dto, items)
	common.Paginated(c, "users", dto, page, pageSize, total)
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
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	updated, err := h.service.UpdateRole(c.Request.Context(), middleware.UserID(c), c.Param("id"), req.Role)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.User
	_ = utils.Map(&dto, updated)
	common.OK(c, "user role updated", dto)
}

// ListSubscriptions godoc
// @Summary      List a user's subscription history (admin)
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Clerk user id"
// @Param        page query int false "page"
// @Param        pageSize query int false "page size"
// @Success      200 {object} common.ApiResponse{data=[]UserSubscription}
// @Router       /admin/users/{id}/subscriptions [get]
func (h *UserController) ListSubscriptions(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	items, total, err := h.subs.ListByUserID(c.Request.Context(), c.Param("id"), pageSize, (page-1)*pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto []*responses.Subscription
	_ = utils.MapSlice(&dto, items)
	common.Paginated(c, "subscriptions", dto, page, pageSize, total)
}

// GrantSubscription godoc
// @Summary      Manually grant a premium subscription (admin)
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Clerk user id"
// @Param        request body UserGrantSubscription true "plan to grant"
// @Success      200 {object} common.ApiResponse{data=UserSubscription}
// @Router       /admin/users/{id}/subscriptions [post]
func (h *UserController) GrantSubscription(c *gin.Context) {
	var req requests.GrantSubscription
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	sub, err := h.subs.GrantUser(c.Request.Context(), c.Param("id"), req.Plan, req.Note)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Subscription
	_ = utils.Map(&dto, sub)
	common.OK(c, "subscription granted", dto)
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

// ListAttempts godoc
// @Summary      List a user's learning attempts (admin)
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Clerk user id"
// @Param        page query int false "page"
// @Param        pageSize query int false "page size"
// @Param        type query string false "attempt type filter"
// @Success      200 {object} common.ApiResponse{data=[]responses.UserAttempt}
// @Router       /admin/users/{id}/attempts [get]
func (h *UserController) ListAttempts(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	var req requests.ListUserAttempts
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	req.Page, req.PageSize = page, pageSize
	items, total, err := h.service.ListAttempts(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto []*responses.UserAttempt
	_ = utils.MapSlice(&dto, items)
	common.Paginated(c, "attempts", dto, page, pageSize, total)
}
