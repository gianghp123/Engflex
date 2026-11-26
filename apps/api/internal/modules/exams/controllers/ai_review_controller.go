// Package controllers holds the exams module's HTTP handlers.
package controllers

import (
	"engflex-api/internal/common"
	"engflex-api/internal/modules/access/dtos/responses"
	accessservices "engflex-api/internal/modules/access/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"

	"github.com/gin-gonic/gin"
)

// AIReviewController serves the paid exam AI endpoints.
type AIReviewController struct {
	reviews *accessservices.AIReviewService
}

// NewAIReviewController builds the controller over the review service.
func NewAIReviewController(reviews *accessservices.AIReviewService) *AIReviewController {
	return &AIReviewController{reviews: reviews}
}

// GenerateExamReview godoc
// @Summary      Generate the AI mistake review for a completed attempt (1 credit)
// @Tags         exams
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "attempt ID"
// @Success      200 {object} common.ApiResponse{data=responses.AIReviewOutput}
// @Failure      400 {object} common.ApiResponse
// @Failure      401 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Failure      422 {object} common.ApiResponse
// @Router       /exam-attempts/{id}/ai-review [post]
func (h *AIReviewController) GenerateExamReview(c *gin.Context) {
	m, err := h.reviews.GenerateExamReview(c.Request.Context(), middleware.UserID(c), c.Param("id"), middleware.IsAdmin(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.AIReviewOutput
	if err := utils.Map(&dto, m); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.OK(c, "review generated", dto)
}

// GetExamReview godoc
// @Summary      Read the cached AI review (free)
// @Tags         exams
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "attempt ID"
// @Success      200 {object} common.ApiResponse{data=responses.AIReviewOutput}
// @Failure      401 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Router       /exam-attempts/{id}/ai-review [get]
func (h *AIReviewController) GetExamReview(c *gin.Context) {
	m, err := h.reviews.GetExamReview(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.AIReviewOutput
	if err := utils.Map(&dto, m); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.OK(c, "ok", dto)
}

// GenerateStudyPlan godoc
// @Summary      Generate the AI study plan for a completed attempt (2 credits)
// @Tags         exams
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "attempt ID"
// @Success      200 {object} common.ApiResponse{data=responses.AIReviewOutput}
// @Failure      400 {object} common.ApiResponse
// @Failure      401 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Failure      422 {object} common.ApiResponse
// @Router       /exam-attempts/{id}/study-plan [post]
func (h *AIReviewController) GenerateStudyPlan(c *gin.Context) {
	m, err := h.reviews.GenerateStudyPlan(c.Request.Context(), middleware.UserID(c), c.Param("id"), middleware.IsAdmin(c))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.AIReviewOutput
	if err := utils.Map(&dto, m); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.OK(c, "plan generated", dto)
}

// GetStudyPlan godoc
// @Summary      Read the cached study plan (free)
// @Tags         exams
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "attempt ID"
// @Success      200 {object} common.ApiResponse{data=responses.AIReviewOutput}
// @Failure      401 {object} common.ApiResponse
// @Failure      404 {object} common.ApiResponse
// @Router       /exam-attempts/{id}/study-plan [get]
func (h *AIReviewController) GetStudyPlan(c *gin.Context) {
	m, err := h.reviews.GetStudyPlan(c.Request.Context(), middleware.UserID(c), c.Param("id"))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.AIReviewOutput
	if err := utils.Map(&dto, m); err != nil {
		common.Fail(c, common.Internal())
		return
	}
	common.OK(c, "ok", dto)
}
