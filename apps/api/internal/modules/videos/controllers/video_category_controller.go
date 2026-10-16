package controllers

import (
	"engflex-api/internal/common"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/modules/videos/dtos/responses"
	"engflex-api/internal/modules/videos/services"
	"engflex-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type VideoCategoryController struct {
	service services.VideoCategoryService
}

func NewVideoCategoryController(service services.VideoCategoryService) *VideoCategoryController {
	return &VideoCategoryController{service: service}
}

// Create godoc
//
//	@Summary		Create category
//	@Description	Create a new category with the specified name
//	@Tags			Video Categories
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CreateVideoCategory								true	"Category payload"
//	@Success		201		{object}	common.ApiResponse{data=responses.VideoCategoryResponse}	"Category created successfully"
//	@Failure		400		{object}	common.ApiResponse									"Invalid request body"
//	@Failure		409		{object}	common.ApiResponse									"Category name already exists"
//	@Failure		500		{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-categories [post]
func (h *VideoCategoryController) Create(c *gin.Context) {
	var req requests.CreateVideoCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VideoCategoryResponse
	_ = utils.Map(&dto, m)
	common.Created(c, "video category created", dto)
}

// Update godoc
//
//	@Summary		Update category
//	@Description	Update an existing category's name by its ID
//	@Tags			Video Categories
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string												true	"Category ID"
//	@Param			request	body		requests.UpdateVideoCategory								true	"Category payload"
//	@Success		200		{object}	common.ApiResponse{data=responses.VideoCategoryResponse}	"Category updated successfully"
//	@Failure		400		{object}	common.ApiResponse									"Invalid ID or payload"
//	@Failure		404		{object}	common.ApiResponse									"Category not found"
//	@Failure		409		{object}	common.ApiResponse									"Category name already exists"
//	@Failure		500		{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-categories/{id} [put]
func (h *VideoCategoryController) Update(c *gin.Context) {
	var req requests.UpdateVideoCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VideoCategoryResponse
	_ = utils.Map(&dto, m)
	common.OK(c, "video category updated", dto)
}

// Delete godoc
//
//	@Summary		Delete category
//	@Description	Delete a category by its ID
//	@Tags			Video Categories
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string				true	"Category ID"
//	@Success		200	{object}	common.ApiResponse	"Category deleted successfully"
//	@Failure		400	{object}	common.ApiResponse	"Invalid ID"
//	@Failure		404	{object}	common.ApiResponse	"Category not found"
//	@Failure		500	{object}	common.ApiResponse	"Internal server error"
//	@Router			/video-categories/{id} [delete]
func (h *VideoCategoryController) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "video category deleted successfully", nil)
}

// GetByID godoc
//
//	@Summary		Get category by ID
//	@Description	Retrieve details of a category by its ID
//	@Tags			Video Categories
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string												true	"Category ID"
//	@Success		200	{object}	common.ApiResponse{data=responses.VideoCategoryResponse}	"Category retrieved successfully"
//	@Failure		400	{object}	common.ApiResponse									"Invalid ID"
//	@Failure		404	{object}	common.ApiResponse									"Category not found"
//	@Failure		500	{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-categories/{id} [get]
func (h *VideoCategoryController) GetByID(c *gin.Context) {
	id := c.Param("id")
	m, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VideoCategoryResponse
	_ = utils.Map(&dto, m)
	common.OK(c, "video category found", dto)
}

// List godoc
//
//	@Summary		List categories
//	@Description	Get paginated list of categories
//	@Tags			Video Categories
//	@Security		BearerAuth
//	@Produce		json
//	@Param			page		query		int													false	"Page number (1-based)"	default(1)
//	@Param			pageSize	query		int													false	"Page size"				default(20)
//	@Success		200			{object}	common.ApiResponse{data=[]responses.VideoCategoryResponse}	"Categories retrieved successfully"
//	@Failure		500			{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-categories [get]
func (h *VideoCategoryController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	offset := utils.Offset(page, pageSize)
	categories, total, err := h.service.List(c.Request.Context(), pageSize, offset)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.VideoCategoryResponse
	_ = utils.MapSlice(&dtos, categories)
	common.Paginated(c, "video categories retrieved", dtos, page, pageSize, total)
}

// ListWithPreview godoc
//
//	@Summary		List categories with a video preview
//	@Description	Every category with its total video count and its previewK newest videos. Not paginated: the video hub renders every category.
//	@Tags			Video Categories
//	@Security		BearerAuth
//	@Produce		json
//	@Param			previewK	query		int		true	"Videos per category (1-32)"
//	@Success		200			{object}	common.ApiResponse{data=[]responses.VideoCategoryPreview}	"Categories with preview retrieved successfully"
//	@Failure		400			{object}	common.ApiResponse									"previewK missing or out of range"
//	@Failure		500			{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-categories/preview [get]
func (h *VideoCategoryController) ListWithPreview(c *gin.Context) {
	var req requests.PreviewVideoCategories
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	categories, err := h.service.ListWithPreview(c.Request.Context(), req.PreviewK)
	if err != nil {
		common.Fail(c, err)
		return
	}
	dtos := make([]responses.VideoCategoryPreview, 0, len(categories))
	_ = utils.MapSlice(&dtos, categories)
	common.OK(c, "video categories with preview retrieved", dtos)
}
