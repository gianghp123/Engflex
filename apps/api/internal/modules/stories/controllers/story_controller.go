package controllers

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/logger"
	"engflex-api/internal/modules/stories/dtos/requests"
	"engflex-api/internal/modules/stories/dtos/responses"
	"engflex-api/internal/modules/stories/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/storage"
	"engflex-api/internal/utils"
)

// maxUploadBytes caps one admin-uploaded story asset (audio or thumbnail).
// Far above the 2MB pronunciation-clip cap elsewhere: a full narration
// track runs minutes, not seconds.
const maxUploadBytes = 50 * 1024 * 1024

type StoryController struct {
	service services.StoryService
	storage *storage.Client
}

func NewStoryController(service services.StoryService, storageClient *storage.Client) *StoryController {
	return &StoryController{service: service, storage: storageClient}
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
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	page, pageSize := utils.ParsePagination(c)
	isAdmin := middleware.Role(c) == enums.UserRoleAdmin
	items, total, err := h.service.ListStories(c.Request.Context(), req, pageSize, utils.Offset(page, pageSize), isAdmin)
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
	m, err := h.service.GetByID(c.Request.Context(), id, middleware.UserID(c), middleware.Role(c) == enums.UserRoleAdmin)
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
	m, err := h.service.GetBySlug(c.Request.Context(), slug, middleware.UserID(c), middleware.Role(c) == enums.UserRoleAdmin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.StoryDetail
	_ = utils.Map(&dto, m)
	common.OK(c, "story", dto)
}

// Create godoc
// @Summary      Create story
// @Tags         stories
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateStory true "Story payload"
// @Success      201 {object} common.ApiResponse{data=responses.StoryDetail}
// @Router       /stories [post]
func (h *StoryController) Create(c *gin.Context) {
	var req requests.CreateStory
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.StoryDetail
	_ = utils.Map(&dto, m)
	common.Created(c, "story created", dto)
}

// Update godoc
// @Summary      Update story
// @Tags         stories
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "story id"
// @Param        request body requests.UpdateStory true "Story payload"
// @Success      200 {object} common.ApiResponse{data=responses.StoryDetail}
// @Router       /stories/{id} [put]
func (h *StoryController) Update(c *gin.Context) {
	id := c.Param("id")
	var req requests.UpdateStory
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.StoryDetail
	_ = utils.Map(&dto, m)
	common.OK(c, "story updated", dto)
}

// Delete godoc
// @Summary      Delete story
// @Tags         stories
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "story id"
// @Success      200 {object} common.ApiResponse
// @Router       /stories/{id} [delete]
func (h *StoryController) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "story deleted successfully", nil)
}

// Upload godoc
// @Summary      Upload a story asset (audio or thumbnail)
// @Tags         stories
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "audio or image file"
// @Success      200 {object} common.ApiResponse{data=object{url=string}}
// @Router       /stories/upload [post]
func (h *StoryController) Upload(c *gin.Context) {
	if h.storage == nil {
		common.Fail(c, common.New(http.StatusServiceUnavailable, "uploads are not configured"))
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		common.Fail(c, common.BadRequest("file is required"))
		return
	}
	if file.Size > maxUploadBytes {
		common.Fail(c, common.New(http.StatusRequestEntityTooLarge, "file too large"))
		return
	}
	f, err := file.Open()
	if err != nil {
		common.Fail(c, common.BadRequest("file is required"))
		return
	}
	defer f.Close()

	ext := strings.ToLower(path.Ext(file.Filename))
	key := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if err := h.storage.Upload(c.Request.Context(), "stories", key, f, file.Size, contentType); err != nil {
		logger.Report(c.Request.Context(), "story asset upload failed", common.Internal())
		common.Fail(c, common.Internal())
		return
	}
	common.OK(c, "uploaded", gin.H{"url": "/api/media/stories/" + key})
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
