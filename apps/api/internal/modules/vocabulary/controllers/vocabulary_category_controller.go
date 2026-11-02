package controllers

import (
	"engflex-api/internal/common"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/dtos/responses"
	"engflex-api/internal/modules/vocabulary/services"
	"engflex-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type VocabularyCategoryController struct {
	service services.VocabularyCategoryService
}

func NewVocabularyCategoryController(service services.VocabularyCategoryService) *VocabularyCategoryController {
	return &VocabularyCategoryController{service: service}
}

// Create godoc
//
//	@Summary		Create vocabulary category
//	@Description	Create a new vocabulary category
//	@Tags			Vocabulary Categories
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CreateVocabularyCategory								true	"Vocabulary category payload"
//	@Success		201		{object}	common.ApiResponse{data=responses.VocabularyCategoryResponse}	"Vocabulary category created successfully"
//	@Failure		400		{object}	common.ApiResponse												"Invalid request body"
//	@Failure		409		{object}	common.ApiResponse												"Vocabulary category name already exists"
//	@Failure		500		{object}	common.ApiResponse												"Internal server error"
//	@Router			/vocabulary-categories [post]
func (h *VocabularyCategoryController) Create(c *gin.Context) {
	var req requests.CreateVocabularyCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	m, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyCategoryResponse
	_ = utils.Map(&dto, m)
	common.Created(c, "vocabulary category created", dto)
}

// Update godoc
//
//	@Summary		Update vocabulary category
//	@Description	Update an existing vocabulary category by its ID
//	@Tags			Vocabulary Categories
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string															true	"Vocabulary Category ID"
//	@Param			request	body		requests.UpdateVocabularyCategory								true	"Vocabulary category payload"
//	@Success		200		{object}	common.ApiResponse{data=responses.VocabularyCategoryResponse}	"Vocabulary category updated successfully"
//	@Failure		400		{object}	common.ApiResponse												"Invalid ID or payload"
//	@Failure		404		{object}	common.ApiResponse												"Vocabulary category not found"
//	@Failure		409		{object}	common.ApiResponse												"Vocabulary category name already exists"
//	@Failure		500		{object}	common.ApiResponse												"Internal server error"
//	@Router			/vocabulary-categories/{id} [put]
func (h *VocabularyCategoryController) Update(c *gin.Context) {
	id := c.Param("id")
	var req requests.UpdateVocabularyCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	m, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyCategoryResponse
	_ = utils.Map(&dto, m)
	common.OK(c, "vocabulary category updated", dto)
}

// Delete godoc
//
//	@Summary		Delete vocabulary category
//	@Description	Delete a vocabulary category by its ID
//	@Tags			Vocabulary Categories
//	@Produce		json
//	@Param			id	path		string				true	"Vocabulary Category ID"
//	@Success		200	{object}	common.ApiResponse	"Vocabulary category deleted successfully"
//	@Failure		400	{object}	common.ApiResponse	"Invalid ID"
//	@Failure		404	{object}	common.ApiResponse	"Vocabulary category not found"
//	@Failure		500	{object}	common.ApiResponse	"Internal server error"
//	@Router			/vocabulary-categories/{id} [delete]
func (h *VocabularyCategoryController) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "vocabulary category deleted successfully", nil)
}

// GetByID godoc
//
//	@Summary		Get vocabulary category by ID
//	@Description	Retrieve details of a vocabulary category by its ID
//	@Tags			Vocabulary Categories
//	@Produce		json
//	@Param			id	path		string															true	"Vocabulary Category ID"
//	@Success		200	{object}	common.ApiResponse{data=responses.VocabularyCategoryResponse}	"Vocabulary category retrieved successfully"
//	@Failure		400	{object}	common.ApiResponse												"Invalid ID"
//	@Failure		404	{object}	common.ApiResponse												"Vocabulary category not found"
//	@Failure		500	{object}	common.ApiResponse												"Internal server error"
//	@Router			/vocabulary-categories/{id} [get]
func (h *VocabularyCategoryController) GetByID(c *gin.Context) {
	id := c.Param("id")
	m, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyCategoryResponse
	_ = utils.Map(&dto, m)
	common.OK(c, "vocabulary category found", dto)
}

// List godoc
//
//	@Summary		List vocabulary categories
//	@Description	Get paginated list of vocabulary categories
//	@Tags			Vocabulary Categories
//	@Produce		json
//	@Param			page		query		int																false	"Page number (1-based)"	default(1)
//	@Param			pageSize	query		int																false	"Page size"				default(20)
//	@Success		200			{object}	common.ApiResponse{data=[]responses.VocabularyCategoryResponse}	"Vocabulary categories retrieved successfully"
//	@Failure		500			{object}	common.ApiResponse												"Internal server error"
//	@Router			/vocabulary-categories [get]
func (h *VocabularyCategoryController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	offset := utils.Offset(page, pageSize)
	categories, total, err := h.service.List(c.Request.Context(), pageSize, offset)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.VocabularyCategoryResponse
	_ = utils.MapSlice(&dtos, categories)
	common.Paginated(c, "vocabulary categories retrieved", dtos, page, pageSize, total)
}
