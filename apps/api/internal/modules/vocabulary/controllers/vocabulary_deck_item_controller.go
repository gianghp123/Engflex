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

// VocabularyDeckItemController exposes flashcard rows.
type VocabularyDeckItemController struct {
	items services.VocabularyDeckItemService
}

// NewVocabularyDeckItemController wires the controller to the service.
func NewVocabularyDeckItemController(items services.VocabularyDeckItemService) *VocabularyDeckItemController {
	return &VocabularyDeckItemController{items: items}
}

// List godoc
// @Summary      List items of one deck
// @Tags         vocabulary-deck-items
// @Security     BearerAuth
// @Produce      json
// @Param        deckId query string true "deck id"
// @Success      200 {object} common.ApiResponse{data=[]responses.VocabularyDeckItemResponse}
// @Router       /vocabulary-deck-items [get]
func (h *VocabularyDeckItemController) List(c *gin.Context) {
	deckID := c.Query("deckId")
	if deckID == "" {
		common.Fail(c, common.BadRequest("deckId is required"))
		return
	}
	items, err := h.items.ListByDeck(c.Request.Context(), middleware.UserID(c), deckID, middleware.Role(c) == enums.UserRoleAdmin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.VocabularyDeckItemResponse
	_ = utils.MapSlice(&dtos, items)
	common.OK(c, "vocabulary deck items", dtos)
}

// GetByID godoc
// @Summary      Get one vocabulary deck item
// @Tags         vocabulary-deck-items
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "item id"
// @Success      200 {object} common.ApiResponse{data=responses.VocabularyDeckItemResponse}
// @Router       /vocabulary-deck-items/{id} [get]
func (h *VocabularyDeckItemController) GetByID(c *gin.Context) {
	m, err := h.items.GetByID(c.Request.Context(), middleware.UserID(c), c.Param("id"), middleware.Role(c) == enums.UserRoleAdmin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyDeckItemResponse
	_ = utils.Map(&dto, m)
	common.OK(c, "vocabulary deck item", dto)
}

// Create godoc
// @Summary      Add a flashcard to a deck
// @Tags         vocabulary-deck-items
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateVocabularyDeckItem true "item payload"
// @Success      201 {object} common.ApiResponse{data=responses.VocabularyDeckItemResponse}
// @Router       /vocabulary-deck-items [post]
func (h *VocabularyDeckItemController) Create(c *gin.Context) {
	var req requests.CreateVocabularyDeckItem
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.items.Create(c.Request.Context(), middleware.UserID(c), middleware.Role(c) == enums.UserRoleAdmin, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyDeckItemResponse
	_ = utils.Map(&dto, m)
	common.Created(c, "vocabulary deck item created", dto)
}

// Update godoc
// @Summary      Update a flashcard
// @Tags         vocabulary-deck-items
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "item id"
// @Param        request body requests.UpdateVocabularyDeckItem true "item payload"
// @Success      200 {object} common.ApiResponse{data=responses.VocabularyDeckItemResponse}
// @Router       /vocabulary-deck-items/{id} [put]
func (h *VocabularyDeckItemController) Update(c *gin.Context) {
	var req requests.UpdateVocabularyDeckItem
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.items.Update(c.Request.Context(), middleware.UserID(c), c.Param("id"), middleware.Role(c) == enums.UserRoleAdmin, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyDeckItemResponse
	_ = utils.Map(&dto, m)
	common.OK(c, "vocabulary deck item updated", dto)
}

// Delete godoc
// @Summary      Delete a flashcard
// @Tags         vocabulary-deck-items
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "item id"
// @Success      200 {object} common.ApiResponse
// @Router       /vocabulary-deck-items/{id} [delete]
func (h *VocabularyDeckItemController) Delete(c *gin.Context) {
	if err := h.items.Delete(c.Request.Context(), middleware.UserID(c), c.Param("id"), middleware.Role(c) == enums.UserRoleAdmin); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "vocabulary deck item deleted", nil)
}
