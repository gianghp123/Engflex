package controllers

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/conversations/dtos/requests"
	"engflex-api/internal/modules/conversations/dtos/responses"
	"engflex-api/internal/modules/conversations/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// ConversationController exposes the conversation lifecycle over HTTP.
type ConversationController struct {
	service     *services.ConversationService
	maxDuration int
}

// NewConversationController wires the controller to the service.
func NewConversationController(service *services.ConversationService, maxDuration int) *ConversationController {
	return &ConversationController{service: service, maxDuration: maxDuration}
}

// Create godoc
// @Summary      Start a conversation
// @Tags         conversations
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body requests.StartConversation true "conversation to start"
// @Success      201 {object} common.ApiResponse{data=responses.Conversation}
// @Failure      400 {object} common.ApiResponse
// @Failure      401 {object} common.ApiResponse
// @Router       /conversations [post]
func (h *ConversationController) Create(c *gin.Context) {
	var req requests.StartConversation
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	userID := middleware.UserID(c)
	m, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Conversation
	_ = utils.Map(&dto, m)
	dto.Turns = []responses.Turn{}
	common.Created(c, "conversation created", dto)
}

// Get godoc
// @Summary      Get a conversation
// @Tags         conversations
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "conversation id"
// @Success      200 {object} common.ApiResponse{data=responses.Conversation}
// @Failure      404 {object} common.ApiResponse
// @Router       /conversations/{id} [get]
func (h *ConversationController) Get(c *gin.Context) {
	userID := middleware.UserID(c)
	m, turns, err := h.service.GetWithFeedback(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Conversation
	_ = utils.Map(&dto, m)
	var turnDtos []responses.Turn
	_ = utils.MapSlice(&turnDtos, turns)
	dto.Turns = turnDtos
	common.OK(c, "conversation", dto)
}

// List godoc
// @Summary      List the caller's finished conversations
// @Description  Ended and failed conversations, newest first, with the scenario and partner attached. Optionally scoped to one scenario.
// @Tags         conversations
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "page"
// @Param        pageSize query int false "page size"
// @Param        scenarioId query string false "scenario filter"
// @Success      200 {object} common.ApiResponse{data=[]responses.ConversationSummary}
// @Failure      400 {object} common.ApiResponse
// @Failure      401 {object} common.ApiResponse
// @Router       /conversations [get]
func (h *ConversationController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	var req requests.ListConversations
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	rows, total, err := h.service.List(c.Request.Context(), middleware.UserID(c), req.ScenarioID, pageSize, utils.Offset(page, pageSize))
	if err != nil {
		common.Fail(c, err)
		return
	}
	dtos := make([]responses.ConversationSummary, 0, len(rows))
	_ = utils.MapSlice(&dtos, rows)
	common.Paginated(c, "conversations", dtos, page, pageSize, total)
}

// Start godoc
// @Summary      Provision the voice session
// @Tags         conversations
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "conversation id"
// @Success      200 {object} common.ApiResponse{data=responses.ConversationSession}
// @Failure      409 {object} common.ApiResponse
// @Failure      503 {object} common.ApiResponse
// @Router       /conversations/{id}/start [post]
func (h *ConversationController) Start(c *gin.Context) {
	userID := middleware.UserID(c)
	conv, ice, err := h.service.Start(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "session ready", responses.ConversationSession{
		SessionID:   conv.ID,
		IceConfig:   ice,
		MaxDuration: h.maxDuration,
	})
}

// Offer godoc
// @Summary      Proxy an SDP offer or ICE candidates to the voice engine
// @Tags         conversations
// @Security     BearerAuth
// @Accept       json
// @Param        id path string true "conversation id"
// @Success      200 {string} string "raw engine response"
// @Failure      409 {object} common.ApiResponse
// @Failure      503 {object} common.ApiResponse
// @Router       /conversations/{id}/offer [post]
// @Router       /conversations/{id}/offer [patch]
func (h *ConversationController) Offer(c *gin.Context) {
	userID := middleware.UserID(c)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		common.Fail(c, common.BadRequest("unreadable request body"))
		return
	}
	respBody, status, err := h.service.Offer(c.Request.Context(), userID, c.Param("id"), c.Request.Method, body)
	if err != nil {
		common.Fail(c, err)
		return
	}
	c.Data(status, "application/json", respBody)
}

// End godoc
// @Summary      End a conversation
// @Tags         conversations
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "conversation id"
// @Success      200 {object} common.ApiResponse{data=responses.Conversation}
// @Router       /conversations/{id}/end [post]
func (h *ConversationController) End(c *gin.Context) {
	userID := middleware.UserID(c)
	m, err := h.service.End(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Conversation
	_ = utils.Map(&dto, m)
	dto.Turns = []responses.Turn{}
	common.OK(c, "conversation ended", dto)
}

// AnalyzeTurn godoc
// @Summary      Analyze one learner turn by position
// @Tags         conversations
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "conversation id"
// @Param        position path int true "turn position (1-based)"
// @Success      200 {object} common.ApiResponse{data=responses.Feedback}
// @Failure      400 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Failure      503 {object} common.ApiResponse
// @Router       /conversations/{id}/turns/{position}/analyze [post]
func (h *ConversationController) AnalyzeTurn(c *gin.Context) {
	userID := middleware.UserID(c)
	position, err := strconv.Atoi(c.Param("position"))
	if err != nil || position < 1 {
		common.Fail(c, common.BadRequest("invalid position"))
		return
	}
	feedback, err := h.service.AnalyzeTurn(c.Request.Context(), userID, c.Param("id"), position)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Feedback
	_ = utils.Map(&dto, feedback)
	common.OK(c, "turn analyzed", dto)
}

// Transcript applies one learner-initiated correction command. The engine owns
// the reviewed turn and resolves which turn is corrected; Go only proves the
// conversation is the caller's and forwards.
//
// @Summary      Apply a transcript correction command
// @Description  One endpoint for the correction modal's three actions (review, send, dismiss). The engine owns the reviewed turn.
// @Tags         conversations
// @Accept       json
// @Produce      json
// @Param        id path string true "conversation id"
// @Param        body body responses.TranscriptCommand true "command"
// @Success      200 {object} common.ApiResponse{data=responses.TranscriptResult}
// @Failure      400 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Failure      409 {object} common.ApiResponse
// @Failure      422 {object} common.ApiResponse
// @Failure      503 {object} common.ApiResponse
// @Router       /conversations/{id}/transcript [post]
func (h *ConversationController) Transcript(c *gin.Context) {
	var req responses.TranscriptCommand
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest("invalid request body"))
		return
	}
	result, err := h.service.Transcript(c.Request.Context(), middleware.UserID(c), c.Param("id"), req.Action, req.Text)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "transcript command applied", result)
}

// Transcribe transcribes one modal re-speak. Multipart audio in, sentence out;
// the engine owns transcription, Go proves ownership and enforces the cap.
//
// @Summary      Transcribe a correction retake
// @Description  Recorded re-speak audio for the correction modal; returns the sentence for the field. Nothing is committed.
// @Tags         conversations
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path string true "conversation id"
// @Param        audio formData file true "recorded re-speak"
// @Success      200 {object} common.ApiResponse{data=responses.TranscribeResult}
// @Failure      400 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Failure      413 {object} common.ApiResponse
// @Failure      422 {object} common.ApiResponse
// @Failure      502 {object} common.ApiResponse
// @Failure      503 {object} common.ApiResponse
// @Router       /conversations/{id}/transcribe [post]
func (h *ConversationController) Transcribe(c *gin.Context) {
	file, err := c.FormFile("audio")
	if err != nil {
		common.Fail(c, common.BadRequest("audio is required"))
		return
	}
	if file.Size > int64(services.MaxTranscribeBytes) {
		common.Fail(c, common.New(http.StatusRequestEntityTooLarge, "audio too large"))
		return
	}
	f, err := file.Open()
	if err != nil {
		common.Fail(c, common.BadRequest("audio is required"))
		return
	}
	defer f.Close()
	audio, err := io.ReadAll(io.LimitReader(f, int64(services.MaxTranscribeBytes)+1))
	if err != nil {
		common.Fail(c, common.BadRequest("audio is required"))
		return
	}
	result, err := h.service.Transcribe(c.Request.Context(), middleware.UserID(c), c.Param("id"), audio, file.Header.Get("Content-Type"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "retake transcribed", result)
}

// Pronounce scores one exercise attempt against its reference sentence.
// Multipart audio in, assessment out; the engine owns scoring, Go proves
// ownership and enforces the cap.
//
// @Summary      Score a pronunciation attempt
// @Description  Attempt audio plus the reference sentence; returns the score with per-word errors. Nothing is committed.
// @Tags         conversations
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path string true "conversation id"
// @Param        audio formData file true "recorded attempt"
// @Param        expected_text formData string true "reference sentence"
// @Param        lang formData string false "language code (default en)"
// @Success      200 {object} common.ApiResponse{data=responses.PronounceResult}
// @Failure      400 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Failure      413 {object} common.ApiResponse
// @Failure      422 {object} common.ApiResponse
// @Failure      502 {object} common.ApiResponse
// @Failure      503 {object} common.ApiResponse
// @Router       /conversations/{id}/pronounce [post]
func (h *ConversationController) Pronounce(c *gin.Context) {
	file, err := c.FormFile("audio")
	if err != nil {
		common.Fail(c, common.BadRequest("audio is required"))
		return
	}
	if file.Size > int64(services.MaxPronounceBytes) {
		common.Fail(c, common.New(http.StatusRequestEntityTooLarge, "audio too large"))
		return
	}
	f, err := file.Open()
	if err != nil {
		common.Fail(c, common.BadRequest("audio is required"))
		return
	}
	defer f.Close()
	audio, err := io.ReadAll(io.LimitReader(f, int64(services.MaxPronounceBytes)+1))
	if err != nil {
		common.Fail(c, common.BadRequest("audio is required"))
		return
	}
	expectedText := c.PostForm("expected_text")
	if expectedText == "" {
		common.Fail(c, common.BadRequest("expected_text is required"))
		return
	}
	lang := c.PostForm("lang")
	if lang == "" {
		lang = "en"
	}
	result, err := h.service.Pronounce(c.Request.Context(), middleware.UserID(c), c.Param("id"), expectedText, lang, audio, file.Header.Get("Content-Type"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "attempt scored", result)
}
