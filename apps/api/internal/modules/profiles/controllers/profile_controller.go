package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/profiles/dtos/requests"
	"engflex-api/internal/modules/profiles/dtos/responses"
	"engflex-api/internal/modules/profiles/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// ProfileController exposes the caller's profile over HTTP.
type ProfileController struct {
	service *services.ProfileService
}

// NewProfileController wires the controller to the service.
func NewProfileController(service *services.ProfileService) *ProfileController {
	return &ProfileController{service: service}
}

// GetMe godoc
// @Summary      Get the caller's profile
// @Tags         profiles
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=responses.Profile}
// @Failure      401 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Router       /profiles/me [get]
func (h *ProfileController) GetMe(c *gin.Context) {
	userID := middleware.UserID(c)
	p, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Profile
	_ = utils.Map(&dto, p)
	common.OK(c, "profile", dto)
}

// UpdateMe godoc
// @Summary      Complete onboarding / update profile
// @Tags         profiles
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body requests.UpdateProfile true "Profile preferences"
// @Success      200 {object} common.ApiResponse{data=responses.Profile}
// @Failure      400 {object} common.ApiResponse
// @Failure      401 {object} common.ApiResponse
// @Failure      500 {object} common.ApiResponse
// @Router       /profiles/me [put]
func (h *ProfileController) UpdateMe(c *gin.Context) {
	var req requests.UpdateProfile
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	userID := middleware.UserID(c)
	p, err := h.service.CompleteOnboarding(c.Request.Context(), userID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Profile
	_ = utils.Map(&dto, p)
	common.OK(c, "onboarding completed", dto)
}
