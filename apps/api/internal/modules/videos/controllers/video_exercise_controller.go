package controllers

import (
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/modules/videos/dtos/responses"
	"engflex-api/internal/modules/videos/services"
	"engflex-api/internal/server/middleware"
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
//	@Summary		Get lesson by ID
//	@Description	Retrieve details of a lesson by its ID
//	@Tags			Video Exercises
//	@Produce		json
//	@Param			id	path		string												true	"Lesson ID"
//	@Success		200	{object}	common.ApiResponse{data=responses.VideoExerciseResponse}	"Lesson retrieved successfully"
//	@Failure		400	{object}	common.ApiResponse									"Invalid ID"
//	@Failure		404	{object}	common.ApiResponse									"Lesson not found"
//	@Failure		500	{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-exercises/{id} [get]
func (h *VideoExerciseController) GetByID(c *gin.Context) {
	id := c.Param("id")
	m, err := h.service.GetByID(c.Request.Context(), id, middleware.Role(c) == enums.UserRoleAdmin)
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
//	@Summary		List lessons
//	@Description	Get paginated list of lessons
//	@Tags			Video Exercises
//	@Produce		json
//	@Param			page		query		int													false	"Page number (1-based)"	default(1)
//	@Param			pageSize	query		int													false	"Page size"				default(20)
//	@Success		200			{object}	common.ApiResponse{data=[]responses.VideoExerciseResponse}	"Lessons retrieved successfully"
//	@Failure		500			{object}	common.ApiResponse									"Internal server error"
//	@Router			/video-exercises [get]
func (h *VideoExerciseController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	offset := utils.Offset(page, pageSize)
	lessons, total, err := h.service.List(c.Request.Context(), pageSize, offset, middleware.Role(c) == enums.UserRoleAdmin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.VideoExerciseResponse
	_ = utils.MapSlice(&dtos, lessons)
	common.Paginated(c, "video exercises retrieved", dtos, page, pageSize, total)
}
