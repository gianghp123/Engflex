package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/modules/exams/dtos/responses"
	"engflex-api/internal/modules/exams/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

type ExamController struct {
	catalog  *services.CatalogService
	attempts *services.AttemptService
}

func NewExamController(catalog *services.CatalogService, attempts *services.AttemptService) *ExamController {
	return &ExamController{catalog: catalog, attempts: attempts}
}

// ListCategories godoc
// @Summary List exam categories
// @Tags exams
// @Security BearerAuth
// @Produce json
// @Param type query string false "exam type"
// @Success 200 {object} common.ApiResponse{data=[]responses.CategoryCatalog}
// @Router /exam-categories [get]
func (h *ExamController) ListCategories(c *gin.Context) {
	items, err := h.catalog.ListCategories(c.Request.Context(), enums.ExamType(c.Query("type")))
	if err != nil {
		common.Fail(c, err)
		return
	}
	dtos := make([]responses.CategoryCatalog, 0, len(items))
	if err := utils.MapSlice(&dtos, items); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.OK(c, "exam categories", dtos)
}

// ListTests godoc
// @Summary List tests in a category
// @Tags exams
// @Security BearerAuth
// @Produce json
// @Param categoryId path string true "category ID"
// @Param page query int false "page"
// @Param pageSize query int false "page size"
// @Success 200 {object} common.ApiResponse{data=[]responses.TestMetadata}
// @Router /exam-categories/{categoryId}/tests [get]
func (h *ExamController) ListTests(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	items, total, err := h.catalog.ListTests(c.Request.Context(), c.Param("categoryId"), page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	dtos := make([]responses.TestMetadata, 0, len(items))
	if err := utils.MapSlice(&dtos, items); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.Paginated(c, "exam tests", dtos, page, pageSize, total)
}

// GetTest godoc
// @Summary Get exam test metadata
// @Description Returns category and part counts only, without questions or answer-bearing content.
// @Tags exams
// @Security BearerAuth
// @Produce json
// @Param id path string true "test UUID"
// @Success 200 {object} common.ApiResponse{data=responses.TestDetail}
// @Failure 400 {object} common.ApiResponse
// @Failure 401 {object} common.ApiResponse
// @Failure 404 {object} common.ApiResponse
// @Router /exam-tests/{id} [get]
func (h *ExamController) GetTest(c *gin.Context) {
	test, err := h.catalog.GetTest(c.Request.Context(), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.TestDetail
	if err := utils.Map(&dto, test); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.OK(c, "exam test", dto)
}

// Start godoc
// @Summary Start an exam attempt
// @Description Each POST creates a new UUID. Exam requires parts 1-7 and has a 120-minute deadline. Practice selects a non-empty subset and has no deadline. Do not retry POST automatically; recover the attempt ID from history.
// @Tags exams
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "test ID"
// @Param body body requests.StartAttempt true "mode and selected parts"
// @Success 201 {object} common.ApiResponse{data=responses.ExamAttempt}
// @Router /exam-tests/{id}/attempts [post]
func (h *ExamController) Start(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	var req requests.StartAttempt
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	attempt, err := h.attempts.Start(c.Request.Context(), middleware.UserID(c), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.ExamAttempt
	if err := utils.Map(&dto, attempt); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.Created(c, "exam attempt started", dto)
}

// ListAttempts godoc
// @Summary List my exam attempts
// @Tags exams
// @Security BearerAuth
// @Produce json
// @Param page query int false "page"
// @Param pageSize query int false "page size"
// @Success 200 {object} common.ApiResponse{data=[]responses.ExamAttemptSummary}
// @Router /exam-attempts [get]
func (h *ExamController) ListAttempts(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	page, pageSize := utils.ParsePagination(c)
	items, total, err := h.attempts.List(c.Request.Context(), middleware.UserID(c), page, pageSize)
	if err != nil {
		common.Fail(c, err)
		return
	}
	dtos := make([]responses.ExamAttemptSummary, 0, len(items))
	if err := utils.MapSlice(&dtos, items); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.Paginated(c, "exam attempts", dtos, page, pageSize, total)
}

// Stats godoc
// @Summary Get my TOEIC exam statistics
// @Description Counts completed exam-mode runs, including legacy unscored runs. Score metrics only include the current estimated scoring policy; practice is excluded.
// @Tags exams
// @Security BearerAuth
// @Produce json
// @Success 200 {object} common.ApiResponse{data=responses.ExamAttemptStats}
// @Failure 401 {object} common.ApiResponse
// @Failure 500 {object} common.ApiResponse
// @Router /exam-attempts/stats [get]
func (h *ExamController) Stats(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	stats, err := h.attempts.Stats(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.ExamAttemptStats
	if err := utils.Map(&dto, stats); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.OK(c, "exam statistics", dto)
}

// CancelAttempt godoc
// @Summary Cancel my exam attempt
// @Description Preserves saved answers. Expired exams are finalized instead. Closed attempts return their existing state.
// @Tags exams
// @Security BearerAuth
// @Produce json
// @Param id path string true "attempt UUID"
// @Success 200 {object} common.ApiResponse{data=responses.ExamAttemptSummary}
// @Failure 400 {object} common.ApiResponse
// @Failure 401 {object} common.ApiResponse
// @Failure 404 {object} common.ApiResponse
// @Router /exam-attempts/{id}/cancel [post]
func (h *ExamController) CancelAttempt(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	attempt, err := h.attempts.Cancel(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.ExamAttemptSummary
	if err := utils.Map(&dto, attempt); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.OK(c, "exam attempt cancellation processed", dto)
}

// GetAttempt godoc
// @Summary Get my exam attempt
// @Description Owner-only. In-progress payload omits answer keys, translations and transcripts. Completed payload includes review content. An expired attempt is finalized using saved answers before returning.
// @Tags exams
// @Security BearerAuth
// @Produce json
// @Param id path string true "attempt ID"
// @Success 200 {object} common.ApiResponse{data=responses.ReviewAttempt}
// @Failure 401 {object} common.ApiResponse
// @Failure 404 {object} common.ApiResponse
// @Router /exam-attempts/{id} [get]
func (h *ExamController) GetAttempt(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	attempt, err := h.attempts.Get(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	if attempt.Status == enums.AttemptStatusCompleted {
		var dto responses.ReviewAttempt
		if err := utils.Map(&dto, attempt); err != nil {
			common.Fail(c, common.Internal())
			return
		}
		common.OK(c, "exam attempt", dto)
		return
	}
	var dto responses.ExamAttempt
	if err := utils.Map(&dto, attempt); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.OK(c, "exam attempt", dto)
}

// PatchAttempt godoc
// @Summary Save answers or submit and grade my attempt
// @Description revision is required. Each answer replaces selectedKey and flagged. Use selectedKey:null to clear. submit:true applies the final batch and grades atomically. Completed submit retries return the previous snapshot. Stale revisions return 409; GET again. Late saves return 409 after finalization; late submit grades saved answers only. Estimated section scores require 100 questions per skill; score is the total when both skills are complete. Legacy result.scaledScore remains null.
// @Tags exams
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "attempt UUID"
// @Param body body requests.PatchAttempt true "Save or submit"
// @Success 200 {object} common.ApiResponse{data=responses.ReviewAttempt}
// @Failure 400 {object} common.ApiResponse
// @Failure 401 {object} common.ApiResponse
// @Failure 404 {object} common.ApiResponse
// @Failure 409 {object} common.ApiResponse
// @Router /exam-attempts/{id} [patch]
func (h *ExamController) PatchAttempt(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	var req requests.PatchAttempt
	decoder := json.NewDecoder(io.LimitReader(c.Request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		common.Fail(c, common.BadRequest("invalid patch body"))
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		common.Fail(c, common.BadRequest("invalid trailing body"))
		return
	}
	m, err := h.attempts.Patch(c.Request.Context(), middleware.UserID(c), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	if m.Status == enums.AttemptStatusCompleted {
		var dto responses.ReviewAttempt
		_ = utils.Map(&dto, m)
		common.OK(c, "exam attempt", dto)
		return
	}
	var dto responses.ExamAttempt
	_ = utils.Map(&dto, m)
	common.OK(c, "exam attempt", dto)
}
