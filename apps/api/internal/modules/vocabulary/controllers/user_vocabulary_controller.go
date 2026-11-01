package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/dtos/responses"
	"engflex-api/internal/modules/vocabulary/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// UserVocabularyController exposes per-user word state.
type UserVocabularyController struct {
	state services.UserVocabularyService
}

// NewUserVocabularyController wires the controller to the service.
func NewUserVocabularyController(state services.UserVocabularyService) *UserVocabularyController {
	return &UserVocabularyController{state: state}
}

// List godoc
// @Summary      List my saved vocabulary
// @Tags         user-vocabulary
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=[]responses.UserVocabularyState}
// @Router       /user-vocabulary [get]
func (h *UserVocabularyController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	items, err := h.state.ListByUser(c.Request.Context(), middleware.UserID(c), pageSize, utils.Offset(page, pageSize))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.UserVocabularyState
	_ = utils.MapSlice(&dtos, items)
	common.OK(c, "user vocabulary", dtos)
}

// Save godoc
// @Summary      Save a word to my vocabulary
// @Tags         user-vocabulary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.SaveUserVocabulary true "save payload"
// @Success      201 {object} common.ApiResponse{data=responses.UserVocabularyState}
// @Router       /user-vocabulary [post]
func (h *UserVocabularyController) Save(c *gin.Context) {
	var req requests.SaveUserVocabulary
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.state.Save(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.UserVocabularyState
	_ = utils.Map(&dto, m)
	common.Created(c, "word saved", dto)
}

// Unsave godoc
// @Summary      Remove a word from my vocabulary
// @Tags         user-vocabulary
// @Security     BearerAuth
// @Produce      json
// @Param        itemId path string true "item id"
// @Success      200 {object} common.ApiResponse
// @Router       /user-vocabulary/{itemId} [delete]
func (h *UserVocabularyController) Unsave(c *gin.Context) {
	if err := h.state.Unsave(c.Request.Context(), middleware.UserID(c), c.Param("itemId")); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "word removed", nil)
}
