package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/dtos/responses"
	"engflex-api/internal/modules/vocabulary/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// VocabularyItemController exposes the shared dictionary.
type VocabularyItemController struct {
	items services.VocabularyItemService
}

// NewVocabularyItemController wires the controller to the service.
func NewVocabularyItemController(items services.VocabularyItemService) *VocabularyItemController {
	return &VocabularyItemController{items: items}
}

// List godoc
// @Summary      List vocabulary items
// @Tags         vocabulary-items
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=[]responses.VocabularyItem}
// @Router       /vocabulary-items [get]
func (h *VocabularyItemController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	f := repositories.VocabularyItemFilter{
		Search: c.Query("search"), CEFR: c.Query("cefr"),
		Limit: pageSize, Offset: utils.Offset(page, pageSize),
	}
	items, total, err := h.items.Catalog(c.Request.Context(), middleware.UserID(c), f)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.VocabularyItem
	_ = utils.MapSlice(&dtos, items)
	common.Paginated(c, "vocabulary items", dtos, page, pageSize, total)
}

// GetByID godoc
// @Summary      Get one vocabulary item
// @Tags         vocabulary-items
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "item id"
// @Success      200 {object} common.ApiResponse{data=responses.VocabularyItem}
// @Router       /vocabulary-items/{id} [get]
func (h *VocabularyItemController) GetByID(c *gin.Context) {
	m, err := h.items.GetByID(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyItem
	_ = utils.Map(&dto, m)
	common.OK(c, "vocabulary item", dto)
}

// Create godoc
// @Summary      Add a custom vocabulary word
// @Tags         vocabulary-items
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateVocabularyItem true "word payload"
// @Success      201 {object} common.ApiResponse{data=responses.VocabularyItem}
// @Router       /vocabulary-items [post]
func (h *VocabularyItemController) Create(c *gin.Context) {
	var req requests.CreateVocabularyItem
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.items.Create(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyItem
	_ = utils.Map(&dto, m)
	common.Created(c, "vocabulary item created", dto)
}
