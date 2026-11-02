package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	dtoscallbacks "engflex-api/internal/modules/conversations/dtos/callbacks"
	"engflex-api/internal/modules/conversations/services"
)

// VoiceCallbackController handles engine -> Go callbacks.
type VoiceCallbackController struct {
	service *services.ConversationService
}

// NewVoiceCallbackController wires the controller to the service.
func NewVoiceCallbackController(service *services.ConversationService) *VoiceCallbackController {
	return &VoiceCallbackController{service: service}
}

// Finalize godoc
// @Summary      Finalize a conversation (engine callback)
// @Tags         internal
// @Accept       json
// @Produce      json
// @Param        id path string true "conversation id"
// @Param        body body callbacks.FinalizeConversation true "duration report"
// @Success      200 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Router       /internal/conversations/{id}/finalize [post]
func (h *VoiceCallbackController) Finalize(c *gin.Context) {
	var req dtoscallbacks.FinalizeConversation
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	if err := h.service.Finalize(c.Request.Context(), c.Param("id"), req); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "conversation finalized", nil)
}
