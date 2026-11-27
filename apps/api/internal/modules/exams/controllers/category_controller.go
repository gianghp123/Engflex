package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/exams/dtos/requests"
	"engflex-api/internal/modules/exams/dtos/responses"
	"engflex-api/internal/utils"
)

// CreateCategory godoc
// @Summary Create an exam category
// @Description subType is optional and only supported for IELTS (academic or general_training).
// @Tags exams
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body requests.CreateCategory true "Exam category"
// @Success 201 {object} common.ApiResponse{data=responses.CategoryReference}
// @Failure 400 {object} common.ApiResponse
// @Failure 401 {object} common.ApiResponse
// @Failure 409 {object} common.ApiResponse
// @Failure 500 {object} common.ApiResponse
// @Router /exam-categories [post]
func (h *ExamController) CreateCategory(c *gin.Context) {
	var req requests.CreateCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	category, err := h.catalog.CreateCategory(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.CategoryReference
	_ = utils.Map(&dto, category)
	common.Created(c, "exam category created", dto)
}
