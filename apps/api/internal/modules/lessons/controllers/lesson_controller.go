package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/config"
	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/modules/lessons/dtos/requests"
	"engflex-api/internal/modules/lessons/dtos/responses"
	"engflex-api/internal/modules/lessons/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// LessonController exposes unit browsing.
type LessonController struct {
	lessons services.LessonService
	media   config.MediaConfig
}

// NewLessonController wires the controller to the service.
func NewLessonController(lessons services.LessonService, media config.MediaConfig) *LessonController {
	return &LessonController{lessons: lessons, media: media}
}

// List godoc
// @Summary      List units grouped by CEFR section (whole path, unpaginated)
// @Tags         lessons
// @Security     BearerAuth
// @Produce      json
// @Param        level query string false "CEFR filter"
// @Success      200 {object} common.ApiResponse{data=[]responses.UnitSection}
// @Router       /lessons [get]
func (h *LessonController) List(c *gin.Context) {
	var req requests.ListLessons
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	// Section-first read: the repository returns sections with units
	// attached, mirroring the DTO shape, so this is one MapSlice.
	secs, err := h.lessons.ListSections(c.Request.Context(), req, middleware.Role(c) == enums.UserRoleAdmin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.UnitSection
	_ = utils.MapSlice(&dtos, secs)
	common.OK(c, "lessons", dtos)
}

// GetByID godoc
// @Summary      Get one unit with ordered activities
// @Tags         lessons
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "lesson id"
// @Success      200 {object} common.ApiResponse{data=responses.LessonDetail}
// @Router       /lessons/{id} [get]
func (h *LessonController) GetByID(c *gin.Context) {
	m, err := h.lessons.GetDetail(c.Request.Context(), c.Param("id"), middleware.Role(c) == enums.UserRoleAdmin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.LessonDetail
	_ = utils.Map(&dto, m)
	for i := range dto.Activities {
		if i < len(m.Activities) {
			ApplyPayload(&dto.Activities[i], m.Activities[i], h.media)
		}
	}
	common.OK(c, "lesson", dto)
}

// Create godoc
// @Summary      Create a lesson (admin)
// @Tags         lessons
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateLesson true "lesson payload"
// @Success      201 {object} common.ApiResponse{data=responses.Lesson}
// @Router       /lessons [post]
func (h *LessonController) Create(c *gin.Context) {
	var req requests.CreateLesson
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.lessons.Create(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Lesson
	_ = utils.Map(&dto, m)
	common.Created(c, "lesson created", dto)
}

// Update godoc
// @Summary      Update a lesson (admin)
// @Tags         lessons
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "lesson id"
// @Param        request body requests.UpdateLesson true "lesson payload"
// @Success      200 {object} common.ApiResponse{data=responses.Lesson}
// @Router       /lessons/{id} [put]
func (h *LessonController) Update(c *gin.Context) {
	var req requests.UpdateLesson
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.lessons.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Lesson
	_ = utils.Map(&dto, m)
	common.OK(c, "lesson updated", dto)
}

// Delete godoc
// @Summary      Delete a lesson (admin)
// @Tags         lessons
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "lesson id"
// @Success      200 {object} common.ApiResponse
// @Router       /lessons/{id} [delete]
func (h *LessonController) Delete(c *gin.Context) {
	if err := h.lessons.Delete(c.Request.Context(), c.Param("id")); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "lesson deleted", nil)
}

// ListSections godoc
// @Summary      List every lesson section, unfiltered (admin)
// @Tags         lesson-sections
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=[]responses.Section}
// @Router       /lesson-sections [get]
func (h *LessonController) ListSections(c *gin.Context) {
	secs, err := h.lessons.ListAllSections(c.Request.Context())
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.Section
	_ = utils.MapSlice(&dtos, secs)
	common.OK(c, "lesson sections", dtos)
}

// CreateSection godoc
// @Summary      Create a lesson section (admin)
// @Tags         lesson-sections
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateLessonSection true "section payload"
// @Success      201 {object} common.ApiResponse{data=responses.Section}
// @Router       /lesson-sections [post]
func (h *LessonController) CreateSection(c *gin.Context) {
	var req requests.CreateLessonSection
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.lessons.CreateSection(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Section
	_ = utils.Map(&dto, m)
	common.Created(c, "lesson section created", dto)
}

// UpdateSection godoc
// @Summary      Update a lesson section (admin)
// @Tags         lesson-sections
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "section id"
// @Param        request body requests.UpdateLessonSection true "section payload"
// @Success      200 {object} common.ApiResponse{data=responses.Section}
// @Router       /lesson-sections/{id} [put]
func (h *LessonController) UpdateSection(c *gin.Context) {
	var req requests.UpdateLessonSection
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.lessons.UpdateSection(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Section
	_ = utils.Map(&dto, m)
	common.OK(c, "lesson section updated", dto)
}

// DeleteSection godoc
// @Summary      Delete a lesson section (admin)
// @Tags         lesson-sections
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "section id"
// @Success      200 {object} common.ApiResponse
// @Router       /lesson-sections/{id} [delete]
func (h *LessonController) DeleteSection(c *gin.Context) {
	if err := h.lessons.DeleteSection(c.Request.Context(), c.Param("id")); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "lesson section deleted", nil)
}

// GetBySlug godoc
// @Summary      Get one unit by slug with ordered activities
// @Tags         lessons
// @Security     BearerAuth
// @Produce      json
// @Param        slug path string true "lesson slug"
// @Success      200 {object} common.ApiResponse{data=responses.LessonDetail}
// @Router       /lessons/by-slug/{slug} [get]
func (h *LessonController) GetBySlug(c *gin.Context) {
	m, err := h.lessons.GetDetailBySlug(c.Request.Context(), c.Param("slug"), middleware.Role(c) == enums.UserRoleAdmin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.LessonDetail
	_ = utils.Map(&dto, m)
	for i := range dto.Activities {
		if i < len(m.Activities) {
			ApplyPayload(&dto.Activities[i], m.Activities[i], h.media)
		}
	}
	common.OK(c, "lesson", dto)
}
