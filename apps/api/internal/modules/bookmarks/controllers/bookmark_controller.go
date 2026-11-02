package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/bookmarks/dtos/requests"
	"engflex-api/internal/modules/bookmarks/dtos/responses"
	"engflex-api/internal/modules/bookmarks/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

type BookmarkController struct {
	bookmarks services.BookmarkService
}

func NewBookmarkController(bookmarks services.BookmarkService) *BookmarkController {
	return &BookmarkController{bookmarks: bookmarks}
}

// List godoc
// @Summary      List my bookmarks
// @Tags         bookmarks
// @Security     BearerAuth
// @Produce      json
// @Param        type query string false "content type filter (lesson, video, story)"
// @Success      200 {object} common.ApiResponse{data=[]responses.Bookmark}
// @Router       /bookmarks [get]
func (h *BookmarkController) List(c *gin.Context) {
	var req requests.ListBookmarks
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	items, err := h.bookmarks.ListByUser(c.Request.Context(), middleware.UserID(c), string(req.ContentType))
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.Bookmark
	_ = utils.MapSlice(&dtos, items)
	common.OK(c, "bookmarks", dtos)
}

// Save godoc
// @Summary      Bookmark a content item (lesson, video, story)
// @Tags         bookmarks
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.SaveBookmark true "bookmark payload"
// @Success      201 {object} common.ApiResponse{data=responses.Bookmark}
// @Router       /bookmarks [post]
func (h *BookmarkController) Save(c *gin.Context) {
	var req requests.SaveBookmark
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.InvalidRequest(err))
		return
	}
	b, err := h.bookmarks.Save(c.Request.Context(), middleware.UserID(c), req.ContentType, req.ContentID)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Bookmark
	_ = utils.Map(&dto, b)
	common.Created(c, "bookmarked", dto)
}

// Unsave godoc
// @Summary      Remove a bookmark
// @Tags         bookmarks
// @Security     BearerAuth
// @Produce      json
// @Param        contentType path string true "content type (lesson, video, story)"
// @Param        contentId path string true "content id"
// @Success      200 {object} common.ApiResponse
// @Router       /bookmarks/{contentType}/{contentId} [delete]
func (h *BookmarkController) Unsave(c *gin.Context) {
	contentType := c.Param("contentType")
	contentID := c.Param("contentId")
	if err := h.bookmarks.Unsave(c.Request.Context(), middleware.UserID(c), contentType, contentID); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "unbookmarked", nil)
}
