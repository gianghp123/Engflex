package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	dtoscallbacks "engflex-api/internal/modules/conversations/dtos/callbacks"
	"engflex-api/internal/modules/conversations/services"
)

// TurnCallbackController handles the engine's finalize-time turn batch.
type TurnCallbackController struct {
	service *services.ConversationService
}

// NewTurnCallbackController wires the controller to the service.
func NewTurnCallbackController(service *services.ConversationService) *TurnCallbackController {
	return &TurnCallbackController{service: service}
}

// IngestTurns godoc
// @Summary      Ingest conversation turns (engine callback)
// @Tags         internal
// @Accept       json
// @Produce      json
// @Param        id path string true "conversation id"
// @Param        body body callbacks.IngestTurns true "turn batch"
// @Success      200 {object} common.ApiResponse
// @Failure      400 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Router       /internal/conversations/{id}/turns [post]
func (h *TurnCallbackController) IngestTurns(c *gin.Context) {
	var req dtoscallbacks.IngestTurns
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	stored, err := h.service.IngestTurns(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "turns ingested", gin.H{"stored": stored})
}
