package controllers

import (
	"encoding/json"

	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/database/models"
	"engflex-api/internal/modules/videos/dtos/requests"
	"engflex-api/internal/modules/videos/dtos/responses"
	"engflex-api/internal/modules/videos/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

type VideoAttemptController struct {
	service services.VideoAttemptService
}

func NewVideoAttemptController(service services.VideoAttemptService) *VideoAttemptController {
	return &VideoAttemptController{service: service}
}

func toVideoAttemptDTO(m *models.Attempt) responses.VideoAttemptResponse {
	var dto responses.VideoAttemptResponse
	_ = utils.Map(&dto, m)
	if len(m.Result) > 0 {
		var payload responses.VideoProgressPayload
		if err := json.Unmarshal(m.Result, &payload); err == nil {
			if payload.CompletedTranscriptIDs == nil {
				payload.CompletedTranscriptIDs = []string{}
			}
			if payload.Transcripts == nil {
				payload.Transcripts = map[string]any{}
			}
			dto.Result = &payload
		}
	}
	if dto.Result == nil {
		dto.Result = &responses.VideoProgressPayload{
			CompletedTranscriptIDs: []string{},
			Transcripts:            map[string]any{},
		}
	}
	return dto
}

// StartAttempt godoc
//
//	@Summary		Start or resume video practice attempt
//	@Description	Starts a new attempt or resumes an existing in-progress attempt for a video exercise (dictation or shadowing)
//	@Tags			Video Attempts
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Video Exercise ID (uuid)"
//	@Param			request	body		requests.StartVideoAttempt	true	"Start video attempt payload"
//	@Success		201		{object}	common.ApiResponse{data=responses.VideoAttemptResponse}	"Attempt started successfully"
//	@Failure		400		{object}	common.ApiResponse			"Invalid request body or ID"
//	@Failure		404		{object}	common.ApiResponse			"Video exercise not found"
//	@Failure		500		{object}	common.ApiResponse			"Internal server error"
//	@Router			/video-exercises/{id}/attempts [post]
func (h *VideoAttemptController) StartAttempt(c *gin.Context) {
	id := c.Param("id")
	var req requests.StartVideoAttempt
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	userID := middleware.UserID(c)
	m, err := h.service.StartAttempt(c.Request.Context(), userID, id, req.Type, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Created(c, "attempt started", toVideoAttemptDTO(m))
}

// GetOpenAttempt godoc
//
//	@Summary		Get open video practice attempt
//	@Description	Retrieve the in-progress practice checkpoint for dictation or shadowing
//	@Tags			Video Attempts
//	@Security		BearerAuth
//	@Produce		json
//	@Param			id		path		string	true	"Video Exercise ID (uuid)"
//	@Param			type	query		string	true	"Practice type"	Enums(dictation, shadowing)
//	@Success		200		{object}	common.ApiResponse{data=responses.VideoAttemptResponse}	"Open attempt found"
//	@Failure		400		{object}	common.ApiResponse	"Invalid ID or practice type"
//	@Failure		404		{object}	common.ApiResponse	"No open attempt found"
//	@Failure		500		{object}	common.ApiResponse	"Internal server error"
//	@Router			/video-exercises/{id}/attempts/open [get]
func (h *VideoAttemptController) GetOpenAttempt(c *gin.Context) {
	id := c.Param("id")
	var req requests.StartVideoAttempt
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	userID := middleware.UserID(c)
	m, err := h.service.GetOpenAttempt(c.Request.Context(), userID, id, req.Type)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "open attempt found", toVideoAttemptDTO(m))
}

// SaveProgress godoc
//
//	@Summary		Save video practice progress
//	@Description	Save incremental progress checkpoint for an in-progress video attempt
//	@Tags			Video Attempts
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string						true	"Video Exercise ID (uuid)"
//	@Param			attemptId	path		string						true	"Attempt ID (uuid)"
//	@Param			request		body		requests.SaveVideoProgress	true	"Progress payload"
//	@Success		200			{object}	common.ApiResponse{data=responses.VideoAttemptResponse}	"Progress saved successfully"
//	@Failure		400			{object}	common.ApiResponse			"Invalid ID or request payload"
//	@Failure		404			{object}	common.ApiResponse			"Attempt not found"
//	@Failure		409			{object}	common.ApiResponse			"Attempt is closed"
//	@Failure		500			{object}	common.ApiResponse			"Internal server error"
//	@Router			/video-exercises/{id}/attempts/{attemptId}/progress [put]
func (h *VideoAttemptController) SaveProgress(c *gin.Context) {
	id := c.Param("id")
	attemptID := c.Param("attemptId")
	var req requests.SaveVideoProgress
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	userID := middleware.UserID(c)
	m, err := h.service.SaveProgress(c.Request.Context(), userID, id, attemptID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "progress saved", toVideoAttemptDTO(m))
}

// CompleteAttempt godoc
//
//	@Summary		Complete video practice attempt
//	@Description	Mark a video practice attempt as completed with final score and duration
//	@Tags			Video Attempts
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string							true	"Video Exercise ID (uuid)"
//	@Param			attemptId	path		string							true	"Attempt ID (uuid)"
//	@Param			request		body		requests.CompleteVideoAttempt	true	"Completion payload"
//	@Success		200			{object}	common.ApiResponse{data=responses.VideoAttemptResponse}	"Attempt completed successfully"
//	@Failure		400			{object}	common.ApiResponse				"Invalid ID or payload"
//	@Failure		404			{object}	common.ApiResponse				"Attempt not found"
//	@Failure		409			{object}	common.ApiResponse				"Attempt is closed"
//	@Failure		500			{object}	common.ApiResponse				"Internal server error"
//	@Router			/video-exercises/{id}/attempts/{attemptId}/complete [post]
func (h *VideoAttemptController) CompleteAttempt(c *gin.Context) {
	id := c.Param("id")
	attemptID := c.Param("attemptId")
	var req requests.CompleteVideoAttempt
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	userID := middleware.UserID(c)
	m, err := h.service.CompleteAttempt(c.Request.Context(), userID, id, attemptID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "attempt completed", toVideoAttemptDTO(m))
}
