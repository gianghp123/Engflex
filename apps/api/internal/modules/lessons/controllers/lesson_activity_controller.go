package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"engflex-api/config"
	"engflex-api/internal/common"
	attemptsresponses "engflex-api/internal/modules/attempts/dtos/responses"
	"engflex-api/internal/modules/lessons/dtos/requests"
	"engflex-api/internal/modules/lessons/dtos/responses"
	"engflex-api/internal/modules/lessons/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
	"engflex-api/internal/voice/pronunciation"
)

// LessonActivityController exposes ordered lesson parts.
type LessonActivityController struct {
	activities services.LessonActivityService
	audio      services.AudioSynthesisService
	media      config.MediaConfig
}

// NewLessonActivityController wires the controller to the services.
func NewLessonActivityController(activities services.LessonActivityService, audio services.AudioSynthesisService, media config.MediaConfig) *LessonActivityController {
	return &LessonActivityController{activities: activities, audio: audio, media: media}
}

// List godoc
// @Summary      List activities of one lesson
// @Tags         lesson-activities
// @Security     BearerAuth
// @Produce      json
// @Param        lessonId query string true "lesson id"
// @Success      200 {object} common.ApiResponse{data=[]responses.Activity}
// @Router       /lesson-activities [get]
func (h *LessonActivityController) List(c *gin.Context) {
	var req requests.ListActivities
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	items, err := h.activities.ListByLesson(c.Request.Context(), req.LessonID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.Activity
	_ = utils.MapSlice(&dtos, items)
	for i := range dtos {
		if i < len(items) {
			ApplyPayload(&dtos[i], items[i], h.media)
		}
	}
	common.OK(c, "lesson activities", dtos)
}

// GetByID godoc
// @Summary      Get one lesson activity
// @Tags         lesson-activities
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "activity id"
// @Success      200 {object} common.ApiResponse{data=responses.Activity}
// @Router       /lesson-activities/{id} [get]
func (h *LessonActivityController) GetByID(c *gin.Context) {
	m, err := h.activities.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Activity
	_ = utils.Map(&dto, m)
	ApplyPayload(&dto, m, h.media)
	common.OK(c, "lesson activity", dto)
}

// CheckAnswer godoc
// @Summary      Check one reading/listening question
// @Tags         lesson-activities
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "activity id"
// @Param        body body requests.CheckAnswer true "answer"
// @Success      200 {object} common.ApiResponse{data=responses.CheckResult}
// @Router       /lesson-activities/{id}/check [post]
func (h *LessonActivityController) CheckAnswer(c *gin.Context) {
	var req requests.CheckAnswer
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	res, err := h.activities.CheckAnswer(c.Request.Context(), middleware.UserID(c), c.Param("id"), req.AttemptID, req.QuestionIndex, req.Key)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "activity checked", res)
}

// Pronounce godoc
// @Summary      Score one read-aloud attempt
// @Tags         lesson-activities
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path string true "activity id"
// @Param        itemIndex formData int true "index of the speaking item"
// @Param        audio formData file true "recorded audio"
// @Success      200 {object} common.ApiResponse{data=responses.PronunciationResult}
// @Router       /lesson-activities/{id}/pronounce [post]
func (h *LessonActivityController) Pronounce(c *gin.Context) {
	file, err := c.FormFile("audio")
	if err != nil {
		common.Fail(c, common.BadRequest("audio is required"))
		return
	}
	if file.Size > int64(pronunciation.MaxAudioBytes) {
		common.Fail(c, common.New(http.StatusRequestEntityTooLarge, "audio too large"))
		return
	}
	itemIndex, err := strconv.Atoi(c.PostForm("itemIndex"))
	if err != nil {
		common.Fail(c, common.BadRequest("itemIndex is required"))
		return
	}
	attemptID := c.PostForm("attemptId")
	if strings.TrimSpace(attemptID) == "" {
		common.Fail(c, common.BadRequest("attemptId is required"))
		return
	}
	f, err := file.Open()
	if err != nil {
		common.Fail(c, common.BadRequest("audio is required"))
		return
	}
	defer f.Close()
	audio, err := io.ReadAll(io.LimitReader(f, int64(pronunciation.MaxAudioBytes)+1))
	if err != nil {
		common.Fail(c, common.BadRequest("audio is required"))
		return
	}
	if len(audio) > pronunciation.MaxAudioBytes {
		common.Fail(c, common.New(http.StatusRequestEntityTooLarge, "audio too large"))
		return
	}
	res, err := h.activities.Pronounce(c.Request.Context(), middleware.UserID(c), c.Param("id"), attemptID, itemIndex, audio, file.Header.Get("Content-Type"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "pronunciation scored", res)
}

// StartAttempt godoc
// @Summary      Open one lesson attempt
// @Tags         lesson-attempts
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "lesson id"
// @Success      200 {object} common.ApiResponse{data=responses.Attempt}
// @Router       /lessons/{id}/attempts [post]
func (h *LessonActivityController) StartAttempt(c *gin.Context) {
	m, err := h.activities.StartAttempt(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto attemptsresponses.Attempt
	_ = utils.Map(&dto, m)
	// datatypes.JSON does not Map into map[string]any; decode explicitly.
	_ = json.Unmarshal(m.Result, &dto.Result)
	if dto.Result == nil {
		dto.Result = map[string]any{}
	}
	common.OK(c, "attempt started", dto)
}

// GetOpenAttempt godoc
// @Summary      Open attempt with graded questions for resume
// @Tags         lesson-attempts
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "lesson id"
// @Success      200 {object} common.ApiResponse{data=responses.AttemptProgress}
// @Router       /lessons/{id}/attempts/open [get]
func (h *LessonActivityController) GetOpenAttempt(c *gin.Context) {
	res, err := h.activities.GetOpenProgress(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "open attempt", res)
}

// ScoreWriting godoc
// @Summary      Score one writing part into the open attempt
// @Tags         lesson-activities
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "activity id"
// @Param        body body requests.ScoreWriting true "attempt and text"
// @Success      200 {object} common.ApiResponse{data=responses.WritingScoreResponse}
// @Router       /lesson-activities/{id}/score [post]
func (h *LessonActivityController) ScoreWriting(c *gin.Context) {
	var req requests.ScoreWriting
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	res, err := h.activities.ScoreWriting(c.Request.Context(), middleware.UserID(c), c.Param("id"), req.AttemptID, req.Text)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "writing scored", res)
}

// SynthesizeAudio godoc
// @Summary      Generate and upload this activity's listening/speaking audio (admin)
// @Tags         lesson-activities
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "activity id"
// @Success      200 {object} common.ApiResponse
// @Router       /lesson-activities/{id}/synthesize-audio [post]
func (h *LessonActivityController) SynthesizeAudio(c *gin.Context) {
	if err := h.audio.SynthesizeForActivity(c.Request.Context(), c.Param("id")); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "audio synthesized", nil)
}
