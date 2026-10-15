package controllers

import (
	"time"

	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/vocabulary/dtos/requests"
	"engflex-api/internal/modules/vocabulary/dtos/responses"
	"engflex-api/internal/modules/vocabulary/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// UserVocabularyController exposes per-user word state. Reads go through the
// hub (VocabularyItemService.Hub: saved items with state); writes go through
// the state service.
type UserVocabularyController struct {
	state services.UserVocabularyService
	hub   services.VocabularyItemService
}

// NewUserVocabularyController wires the controller to both services.
func NewUserVocabularyController(state services.UserVocabularyService, hub services.VocabularyItemService) *UserVocabularyController {
	return &UserVocabularyController{state: state, hub: hub}
}

// List godoc
// @Summary      List my saved vocabulary
// @Tags         user-vocabulary
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=[]responses.VocabularyItem}
// @Router       /user-vocabulary [get]
func (h *UserVocabularyController) List(c *gin.Context) {
	var req requests.ListVocabulary
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	page, pageSize := utils.ParsePagination(c)
	f := repositories.VocabularyItemFilter{
		Search: req.Search, CEFR: string(req.CEFR),
		Source: string(req.Source), Sort: string(req.Sort),
		Limit: pageSize, Offset: utils.Offset(page, pageSize),
	}
	items, total, err := h.hub.Hub(c.Request.Context(), middleware.UserID(c), f)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.VocabularyItem
	_ = utils.MapSlice(&dtos, items)
	common.Paginated(c, "user vocabulary", dtos, page, pageSize, total)
}

// Stats godoc
// @Summary      Get my vocabulary stats
// @Tags         user-vocabulary
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} common.ApiResponse{data=responses.VocabularyStats}
// @Router       /user-vocabulary/stats [get]
func (h *UserVocabularyController) Stats(c *gin.Context) {
	now := time.Now().UTC()
	// The reader's local midnight, so "saved today" is their day. Absent, the
	// UTC day is the documented default.
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if raw := c.Query("dayStart"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			common.Fail(c, common.BadRequest("dayStart must be RFC3339"))
			return
		}
		dayStart = parsed
	}
	stats, err := h.hub.Stats(c.Request.Context(), middleware.UserID(c), now, dayStart)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyStats
	_ = utils.Map(&dto, stats)
	common.OK(c, "vocabulary stats", dto)
}

// Save godoc
// @Summary      Save a word to my vocabulary
// @Tags         user-vocabulary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.SaveUserVocabulary true "save payload"
// @Success      201 {object} common.ApiResponse{data=responses.UserVocabularyState}
// @Router       /user-vocabulary [post]
func (h *UserVocabularyController) Save(c *gin.Context) {
	var req requests.SaveUserVocabulary
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.state.Save(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.UserVocabularyState
	_ = utils.Map(&dto, m)
	common.Created(c, "word saved", dto)
}

// Review godoc
// @Summary      Grade a review for one saved word
// @Tags         user-vocabulary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        itemId path string true "item id"
// @Param        request body requests.ReviewUserVocabulary true "review payload"
// @Success      200 {object} common.ApiResponse{data=responses.UserVocabularyState}
// @Router       /user-vocabulary/{itemId}/review [post]
func (h *UserVocabularyController) Review(c *gin.Context) {
	var req requests.ReviewUserVocabulary
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.state.Review(c.Request.Context(), middleware.UserID(c), c.Param("itemId"), *req.Grade)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.UserVocabularyState
	_ = utils.Map(&dto, m)
	common.OK(c, "reviewed", dto)
}

// Patch godoc
// @Summary      Set mastery or note for one saved word
// @Tags         user-vocabulary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        itemId path string true "item id"
// @Param        request body requests.PatchUserVocabulary true "patch payload"
// @Success      200 {object} common.ApiResponse{data=responses.UserVocabularyState}
// @Router       /user-vocabulary/{itemId} [patch]
func (h *UserVocabularyController) Patch(c *gin.Context) {
	var req requests.PatchUserVocabulary
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.state.Patch(c.Request.Context(), middleware.UserID(c), c.Param("itemId"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.UserVocabularyState
	_ = utils.Map(&dto, m)
	common.OK(c, "vocabulary updated", dto)
}

// Unsave godoc
// @Summary      Remove a word from my vocabulary
// @Tags         user-vocabulary
// @Security     BearerAuth
// @Produce      json
// @Param        itemId path string true "item id"
// @Success      200 {object} common.ApiResponse
// @Router       /user-vocabulary/{itemId} [delete]
func (h *UserVocabularyController) Unsave(c *gin.Context) {
	if err := h.state.Unsave(c.Request.Context(), middleware.UserID(c), c.Param("itemId")); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "word removed", nil)
}

// AddWord godoc
// @Summary      Add a word to my vocabulary by content
// @Tags         user-vocabulary
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.AddUserVocabularyWord true "word payload"
// @Success      201 {object} common.ApiResponse{data=responses.VocabularyItem}
// @Router       /user-vocabulary/words [post]
func (h *UserVocabularyController) AddWord(c *gin.Context) {
	var req requests.AddUserVocabularyWord
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.hub.AddWord(c.Request.Context(), middleware.UserID(c), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.VocabularyItem
	_ = utils.Map(&dto, m)
	common.Created(c, "word added", dto)
}
