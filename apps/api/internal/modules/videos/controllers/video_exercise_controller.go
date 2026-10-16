package controllers

import (
	"engflex-api/internal/common"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/modules/videos/dtos/responses"
	"engflex-api/internal/modules/videos/services"
	"engflex-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type VideoExerciseController struct {
	service services.VideoExerciseService
}

func NewVideoExerciseController(service services.VideoExerciseService) *VideoExerciseController {
	return &VideoExerciseController{service: service}
}

// Create godoc
//
//	@Summary		Create lesson
//	@Description	Create a new lesson
//	@Tags			Video Exercises
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CreateVideoExercise								true	"Lesson payload"
//	@Success		201		{object}	common.ApiResponse{data=responses.VideoExerciseResponse}	"Lesson created successfully"
//	@Failure		400		{object}	common.ApiResponse									"Invalid request body"
//	@Failure		409		{object}	common.ApiResponse									"Lesson conflict"
//	@Failure		500		{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-exercises [post]
func (h *VideoExerciseController) Create(c *gin.Context) {
	var req requests.CreateVideoExercise
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VideoExerciseResponse
	_ = utils.Map(&dto, m)
	common.Created(c, "video exercise created", dto)
}

// Update godoc
//
//	@Summary		Update lesson
//	@Description	Update an existing lesson by its ID
//	@Tags			Video Exercises
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string												true	"Lesson ID"
//	@Param			request	body		requests.UpdateVideoExercise								true	"Lesson payload"
//	@Success		200		{object}	common.ApiResponse{data=responses.VideoExerciseResponse}	"Lesson updated successfully"
//	@Failure		400		{object}	common.ApiResponse									"Invalid ID or payload"
//	@Failure		404		{object}	common.ApiResponse									"Lesson not found"
//	@Failure		500		{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-exercises/{id} [put]
func (h *VideoExerciseController) Update(c *gin.Context) {
	id := c.Param("id")
	var req requests.UpdateVideoExercise
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VideoExerciseResponse
	_ = utils.Map(&dto, m)
	common.OK(c, "video exercise updated", dto)
}

// Delete godoc
//
//	@Summary		Delete lesson
//	@Description	Delete a lesson by its ID
//	@Tags			Video Exercises
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string				true	"Lesson ID"
//	@Success		200	{object}	common.ApiResponse	"Lesson deleted successfully"
//	@Failure		400	{object}	common.ApiResponse	"Invalid ID"
//	@Failure		404	{object}	common.ApiResponse	"Lesson not found"
//	@Failure		500	{object}	common.ApiResponse	"Internal server error"
//	@Router			/video-exercises/{id} [delete]
func (h *VideoExerciseController) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "video exercise deleted successfully", nil)
}

// GetByID godoc
//
//	@Summary		Get video exercise by ID
//	@Description	Retrieve one video exercise with its category and its transcripts ordered by sequence
//	@Tags			Video Exercises
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string												true	"Video exercise ID"
//	@Success		200	{object}	common.ApiResponse{data=responses.VideoExerciseResponse}	"Video exercise retrieved successfully"
//	@Failure		400	{object}	common.ApiResponse									"Invalid ID"
//	@Failure		404	{object}	common.ApiResponse									"Video exercise not found"
//	@Failure		500	{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-exercises/{id} [get]
func (h *VideoExerciseController) GetByID(c *gin.Context) {
	id := c.Param("id")
	m, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VideoExerciseResponse
	_ = utils.Map(&dto, m)
	common.OK(c, "video exercise found", dto)
}

// List godoc
//
//	@Summary		List video exercises
//	@Description	Get a paginated list of video exercises, newest first, each with its category
//	@Tags			Video Exercises
//	@Security		BearerAuth
//	@Produce		json
//	@Param			page		query		int													false	"Page number (1-based)"	default(1)
//	@Param			pageSize	query		int													false	"Page size"				default(20)
//	@Param			search		query		string												false	"Case-insensitive match on title or description"
//	@Param			categoryId	query		string												false	"Category ID (uuid)"
//	@Param			cefrLevel	query		string												false	"CEFR level"	Enums(A1, A2, B1, B2, C1, C2)
//	@Success		200			{object}	common.ApiResponse{data=[]responses.VideoExerciseResponse}	"Video exercises retrieved successfully"
//	@Failure		400			{object}	common.ApiResponse									"Invalid filter"
//	@Failure		500			{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-exercises [get]
func (h *VideoExerciseController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	var req requests.ListVideoExercises
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	// The clamped values win over the raw query: pageSize is capped at
	// common.MaxPageSize by ParsePagination, never by the client.
	req.Page, req.PageSize = page, pageSize
	lessons, total, err := h.service.List(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	// Non-nil so an empty page encodes as `"data": []`: a nil slice inside
	// the `any` Data field encodes as `null`, which the typed web client
	// would read as a missing array.
	dtos := make([]responses.VideoExerciseResponse, 0, len(lessons))
	_ = utils.MapSlice(&dtos, lessons)
	common.Paginated(c, "video exercises retrieved", dtos, page, pageSize, total)
}
