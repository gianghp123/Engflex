package controllers

import (
	"engflex-api/internal/common"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/modules/videos/dtos/responses"
	"engflex-api/internal/modules/videos/services"
	"engflex-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type VideoTranscriptController struct {
	service services.VideoTranscriptService
}

func NewVideoTranscriptController(service services.VideoTranscriptService) *VideoTranscriptController {
	return &VideoTranscriptController{service: service}
}

// Create godoc
//
//	@Summary		Create transcript
//	@Description	Create a new transcript
//	@Tags			Video Transcripts
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CreateVideoTranscript								true	"Transcript payload"
//	@Success		201		{object}	common.ApiResponse{data=responses.VideoTranscriptResponse}	"Transcript created successfully"
//	@Failure		400		{object}	common.ApiResponse										"Invalid request body"
//	@Failure		409		{object}	common.ApiResponse										"Transcript conflict"
//	@Failure		500		{object}	common.ApiResponse										"Internal server error"
//	@Router			/video-transcripts [post]
func (h *VideoTranscriptController) Create(c *gin.Context) {
	var req requests.CreateVideoTranscript
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VideoTranscriptResponse
	_ = utils.Map(&dto, m)
	common.Created(c, "video transcript created", dto)
}

// Update godoc
//
//	@Summary		Update transcript
//	@Description	Update an existing transcript by its ID
//	@Tags			Video Transcripts
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string													true	"Transcript ID"
//	@Param			request	body		requests.UpdateVideoTranscript								true	"Transcript payload"
//	@Success		200		{object}	common.ApiResponse{data=responses.VideoTranscriptResponse}	"Transcript updated successfully"
//	@Failure		400		{object}	common.ApiResponse										"Invalid ID or payload"
//	@Failure		404		{object}	common.ApiResponse										"Transcript not found"
//	@Failure		500		{object}	common.ApiResponse										"Internal server error"
//	@Router			/video-transcripts/{id} [put]
func (h *VideoTranscriptController) Update(c *gin.Context) {
	id := c.Param("id")
	var req requests.UpdateVideoTranscript
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VideoTranscriptResponse
	_ = utils.Map(&dto, m)
	common.OK(c, "video transcript updated", dto)
}

// Delete godoc
//
//	@Summary		Delete transcript
//	@Description	Delete a transcript by its ID
//	@Tags			Video Transcripts
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string				true	"Transcript ID"
//	@Success		200	{object}	common.ApiResponse	"Transcript deleted successfully"
//	@Failure		400	{object}	common.ApiResponse	"Invalid ID"
//	@Failure		404	{object}	common.ApiResponse	"Transcript not found"
//	@Failure		500	{object}	common.ApiResponse	"Internal server error"
//	@Router			/video-transcripts/{id} [delete]
func (h *VideoTranscriptController) Delete(c *gin.Context) {
	id := c.Param("id")
	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "video transcript deleted successfully", nil)
}

// GetByID godoc
//
//	@Summary		Get transcript by ID
//	@Description	Retrieve details of a transcript by its ID
//	@Tags			Video Transcripts
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id	path		string													true	"Transcript ID"
//	@Success		200	{object}	common.ApiResponse{data=responses.VideoTranscriptResponse}	"Transcript retrieved successfully"
//	@Failure		400	{object}	common.ApiResponse										"Invalid ID"
//	@Failure		404	{object}	common.ApiResponse										"Transcript not found"
//	@Failure		500	{object}	common.ApiResponse										"Internal server error"
//	@Router			/video-transcripts/{id} [get]
func (h *VideoTranscriptController) GetByID(c *gin.Context) {
	id := c.Param("id")
	m, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VideoTranscriptResponse
	_ = utils.Map(&dto, m)
	common.OK(c, "video transcript found", dto)
}

// GetByVideoExerciseID godoc
//
//	@Summary		Get transcripts by lesson ID
//	@Description	Retrieve list of transcripts belonging to a lesson
//	@Tags			Video Transcripts
//	@Security		BearerAuth
//	@Produce		json
//	@Param			video_exercise_id	path		string													true	"Lesson ID"
//	@Success		200			{object}	common.ApiResponse{data=[]responses.VideoTranscriptResponse}	"Transcripts retrieved successfully"
//	@Failure		400			{object}	common.ApiResponse										"Invalid lesson ID"
//	@Failure		500			{object}	common.ApiResponse										"Internal server error"
//	@Router			/video-transcripts/video-exercise/{video_exercise_id} [get]
func (h *VideoTranscriptController) GetByVideoExerciseID(c *gin.Context) {
	videoExerciseID := c.Param("video_exercise_id")
	list, err := h.service.GetByVideoExerciseID(c.Request.Context(), videoExerciseID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.VideoTranscriptResponse
	_ = utils.MapSlice(&dtos, list)
	common.OK(c, "video transcripts retrieved", dtos)
}

// List godoc
//
//	@Summary		List transcripts
//	@Description	Get paginated list of transcripts
//	@Tags			Video Transcripts
//	@Security		BearerAuth
//	@Produce		json
//	@Param			page		query		int														false	"Page number (1-based)"	default(1)
//	@Param			pageSize	query		int														false	"Page size"				default(20)
//	@Success		200			{object}	common.ApiResponse{data=[]responses.VideoTranscriptResponse}	"Transcripts retrieved successfully"
//	@Failure		500			{object}	common.ApiResponse										"Internal server error"
//	@Router			/video-transcripts [get]
func (h *VideoTranscriptController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	offset := utils.Offset(page, pageSize)
	list, total, err := h.service.List(c.Request.Context(), pageSize, offset)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.VideoTranscriptResponse
	_ = utils.MapSlice(&dtos, list)
	common.Paginated(c, "video transcripts retrieved", dtos, page, pageSize, total)
}
