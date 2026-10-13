package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/dtos/responses"
	"engflex-api/internal/modules/vocabulary/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// VocabularyDeckController exposes user flashcard decks.
type VocabularyDeckController struct {
	decks services.VocabularyDeckService
}

// NewVocabularyDeckController wires the controller to the service.
func NewVocabularyDeckController(decks services.VocabularyDeckService) *VocabularyDeckController {
	return &VocabularyDeckController{decks: decks}
}

// List godoc
// @Summary      List my vocabulary decks
// @Tags         vocabulary-decks
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=[]responses.VocabularyDeckDetail}
// @Router       /vocabulary-decks [get]
func (h *VocabularyDeckController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	items, total, err := h.decks.ListByUser(c.Request.Context(), middleware.UserID(c), pageSize, utils.Offset(page, pageSize))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.VocabularyDeckDetail
	_ = utils.MapSlice(&dtos, items)
	common.Paginated(c, "vocabulary decks", dtos, page, pageSize, total)
}

// GetByID godoc
// @Summary      Get one vocabulary deck with category
// @Tags         vocabulary-decks
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "deck id"
// @Success      200 {object} common.ApiResponse{data=responses.VocabularyDeckDetail}
// @Router       /vocabulary-decks/{id} [get]
func (h *VocabularyDeckController) GetByID(c *gin.Context) {
	m, err := h.decks.GetDetail(c.Request.Context(), middleware.UserID(c), c.Param("id"), middleware.Role(c) == enums.UserRoleAdmin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyDeckDetail
	_ = utils.Map(&dto, m)
	common.OK(c, "vocabulary deck", dto)
}

// Create godoc
// @Summary      Create a vocabulary deck
// @Tags         vocabulary-decks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateVocabularyDeck true "deck payload"
// @Success      201 {object} common.ApiResponse{data=responses.VocabularyDeckDetail}
// @Router       /vocabulary-decks [post]
func (h *VocabularyDeckController) Create(c *gin.Context) {
	var req requests.CreateVocabularyDeck
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.decks.Create(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyDeckDetail
	_ = utils.Map(&dto, m)
	common.Created(c, "vocabulary deck created", dto)
}

// Update godoc
// @Summary      Update a vocabulary deck
// @Tags         vocabulary-decks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "deck id"
// @Param        request body requests.UpdateVocabularyDeck true "deck payload"
// @Success      200 {object} common.ApiResponse{data=responses.VocabularyDeckDetail}
// @Router       /vocabulary-decks/{id} [put]
func (h *VocabularyDeckController) Update(c *gin.Context) {
	var req requests.UpdateVocabularyDeck
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.decks.Update(c.Request.Context(), middleware.UserID(c), c.Param("id"), middleware.Role(c) == enums.UserRoleAdmin, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyDeckDetail
	_ = utils.Map(&dto, m)
	common.OK(c, "vocabulary deck updated", dto)
}

// Delete godoc
// @Summary      Delete a vocabulary deck
// @Tags         vocabulary-decks
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "deck id"
// @Success      200 {object} common.ApiResponse
// @Router       /vocabulary-decks/{id} [delete]
func (h *VocabularyDeckController) Delete(c *gin.Context) {
	if err := h.decks.Delete(c.Request.Context(), middleware.UserID(c), c.Param("id"), middleware.Role(c) == enums.UserRoleAdmin); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "vocabulary deck deleted", nil)
}

// CreateSystemDeck godoc
// @Summary      Create a system-curated vocabulary deck (admin)
// @Tags         vocabulary-decks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateVocabularyDeck true "deck payload"
// @Success      201 {object} common.ApiResponse{data=responses.VocabularyDeckDetail}
// @Router       /admin/vocabulary-decks [post]
func (h *VocabularyDeckController) CreateSystemDeck(c *gin.Context) {
	var req requests.CreateVocabularyDeck
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.decks.CreateSystemDeck(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyDeckDetail
	_ = utils.Map(&dto, m)
	common.Created(c, "system vocabulary deck created", dto)
}
