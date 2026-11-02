package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/stories/dtos/requests"
	"engflex-api/internal/modules/stories/dtos/responses"
	"engflex-api/internal/modules/stories/services"
	"engflex-api/internal/utils"
)

type StoryController struct {
	service services.StoryService
}

func NewStoryController(service services.StoryService) *StoryController {
	return &StoryController{service: service}
}

// List godoc
// @Summary      List stories
// @Tags         stories
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "page number"
// @Param        pageSize query int false "page size"
// @Param        cefr query string false "CEFR level (A1, A2, B1, B2, C1, C2)"
// @Param        topic query string false "topic"
// @Param        search query string false "search term"
// @Success      200 {object} common.ApiResponse{data=[]responses.StoryItem}
// @Router       /stories [get]
func (h *StoryController) List(c *gin.Context) {
	var req requests.ListStories
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	page, pageSize := utils.ParsePagination(c)
	items, total, err := h.service.ListStories(c.Request.Context(), req, pageSize, utils.Offset(page, pageSize))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.StoryItem
	_ = utils.MapSlice(&dtos, items)
	common.Paginated(c, "stories", dtos, page, pageSize, total)
}

// GetByID godoc
// @Summary      Get story detail by ID
// @Tags         stories
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "story id"
// @Success      200 {object} common.ApiResponse{data=responses.StoryDetail}
// @Router       /stories/{id} [get]
func (h *StoryController) GetByID(c *gin.Context) {
	id := c.Param("id")
	m, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.StoryDetail
	_ = utils.Map(&dto, m)
	common.OK(c, "story", dto)
}

// GetBySlug godoc
// @Summary      Get story detail by slug
// @Tags         stories
// @Security     BearerAuth
// @Produce      json
// @Param        slug path string true "story slug"
// @Success      200 {object} common.ApiResponse{data=responses.StoryDetail}
// @Router       /stories/by-slug/{slug} [get]
func (h *StoryController) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	m, err := h.service.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.StoryDetail
	_ = utils.Map(&dto, m)
	common.OK(c, "story", dto)
}

// ListTopics godoc
// @Summary      List story topics
// @Tags         stories
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=[]string}
// @Router       /stories/topics [get]
func (h *StoryController) ListTopics(c *gin.Context) {
	topics, err := h.service.ListTopics(c.Request.Context())
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "story topics retrieved", topics)
}
