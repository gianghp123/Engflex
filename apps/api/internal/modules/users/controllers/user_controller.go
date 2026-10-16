package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/users/dtos/requests"
	"engflex-api/internal/modules/users/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// UserController exposes the Clerk-backed admin user directory.
type UserController struct {
	service services.UserService
}

// NewUserController wires the controller to the service.
func NewUserController(service services.UserService) *UserController {
	return &UserController{service: service}
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
