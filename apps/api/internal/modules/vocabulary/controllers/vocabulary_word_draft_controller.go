package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/services"
)

// VocabularyWordDraftController serves the add-word dialog's LLM draft.
type VocabularyWordDraftController struct {
	drafts services.VocabularyWordDraftService
}

// NewVocabularyWordDraftController builds the controller over the draft service.
func NewVocabularyWordDraftController(drafts services.VocabularyWordDraftService) *VocabularyWordDraftController {
	return &VocabularyWordDraftController{drafts: drafts}
}

// GenerateWordDraft godoc
// @Summary      Generate a word draft with the LLM
// @Description  Read-only: returns a draft the add-word dialog shows for review; nothing is stored.
// @Tags         vocabulary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body requests.GenerateWordDraft true "term, optional context sentence, and optional output language"
// @Success      200 {object} common.ApiResponse{data=responses.WordDraft}
// @Router       /vocabulary-items/generate [post]
func (h *VocabularyWordDraftController) GenerateWordDraft(c *gin.Context) {
	var req requests.GenerateWordDraft
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	draft, err := h.drafts.Generate(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "word draft generated", draft)
}
