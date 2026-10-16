package controllers

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/common/enums"
	"engflex-api/internal/modules/scenarios/dtos/requests"
	"engflex-api/internal/modules/scenarios/dtos/responses"
	"engflex-api/internal/modules/scenarios/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/utils"
)

// ScenarioController exposes scenario browsing.
type ScenarioController struct {
	service *services.ScenarioService
}

// NewScenarioController wires the controller to the service.
func NewScenarioController(service *services.ScenarioService) *ScenarioController {
	return &ScenarioController{service: service}
}

// List godoc
// @Summary      List scenarios
// @Tags         scenarios
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "page"
// @Param        pageSize query int false "page size"
// @Param        topicId query string false "topic filter"
// @Success      200 {object} common.ApiResponse{data=[]responses.Scenario}
// @Router       /scenarios [get]
func (h *ScenarioController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	var req requests.ListScenarios
	if err := c.ShouldBindQuery(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	req.Page, req.PageSize = page, pageSize
	items, total, err := h.service.List(c.Request.Context(), req, middleware.Role(c) == enums.UserRoleAdmin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	dtos := make([]responses.Scenario, 0, len(items))
	_ = utils.MapSlice(&dtos, items)
	common.Paginated(c, "scenarios", dtos, page, pageSize, total)
}

// TopicsWithPreview godoc
// @Summary      List scenario topics with top-k previews
// @Tags         scenario-topics
// @Security     BearerAuth
// @Produce      json
// @Param        previewK query int false "preview cards per topic (default 3, max 6)"
// @Param        difficulty query string false "B1+, B2 or C1 filter"
// @Param        search query string false "title/objective substring"
// @Success      200 {object} common.ApiResponse{data=[]responses.ScenarioTopic}
// @Router       /scenario-topics [get]
func (h *ScenarioController) TopicsWithPreview(c *gin.Context) {
	previewK := 3
	if raw := c.Query("previewK"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			previewK = n
		}
	}
	var difficulty *string
	if d := c.Query("difficulty"); d != "" {
		switch d {
		case "B1+", "B2", "C1":
			difficulty = &d
		default:
			common.Fail(c, common.BadRequest("difficulty must be one of B1+, B2, C1"))
			return
		}
	}
	search := strings.TrimSpace(c.Query("search"))
	topics, err := h.service.ListTopicsWithPreview(c.Request.Context(), previewK, difficulty, search)
	if err != nil {
		common.Fail(c, err)
		return
	}
	dtos := make([]responses.ScenarioTopic, 0, len(topics))
	_ = utils.MapSlice(&dtos, topics)
	// Plain array, NOT Paginated: the web reads it with api<T>. Each banner
	// carries its own scenarios, so the browser renders what it got.
	common.OK(c, "scenario topics", dtos)
}

// GetByID godoc
// @Summary      Get one scenario with detail content
// @Tags         scenarios
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "scenario id"
// @Success      200 {object} common.ApiResponse{data=responses.Scenario}
// @Router       /scenarios/{id} [get]
func (h *ScenarioController) GetByID(c *gin.Context) {
	m, err := h.service.GetDetail(c.Request.Context(), c.Param("id"), middleware.Role(c) == enums.UserRoleAdmin)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Scenario
	_ = utils.Map(&dto, m)
	common.OK(c, "scenario", dto)
}

// Create godoc
// @Summary      Create a scenario (admin)
// @Tags         scenarios
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateScenario true "scenario payload"
// @Success      201 {object} common.ApiResponse{data=responses.Scenario}
// @Router       /scenarios [post]
func (h *ScenarioController) Create(c *gin.Context) {
	var req requests.CreateScenario
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Scenario
	_ = utils.Map(&dto, m)
	common.Created(c, "scenario created", dto)
}

// Update godoc
// @Summary      Update a scenario (admin)
// @Tags         scenarios
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "scenario id"
// @Param        request body requests.UpdateScenario true "scenario payload"
// @Success      200 {object} common.ApiResponse{data=responses.Scenario}
// @Router       /scenarios/{id} [put]
func (h *ScenarioController) Update(c *gin.Context) {
	var req requests.UpdateScenario
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Scenario
	_ = utils.Map(&dto, m)
	common.OK(c, "scenario updated", dto)
}

// Delete godoc
// @Summary      Delete a scenario (admin)
// @Tags         scenarios
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "scenario id"
// @Success      200 {object} common.ApiResponse
// @Router       /scenarios/{id} [delete]
func (h *ScenarioController) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "scenario deleted", nil)
}

// CreateTopic godoc
// @Summary      Create a scenario topic (admin)
// @Tags         scenario-topics
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreateScenarioTopic true "topic payload"
// @Success      201 {object} common.ApiResponse{data=responses.ScenarioTopic}
// @Router       /scenario-topics [post]
func (h *ScenarioController) CreateTopic(c *gin.Context) {
	var req requests.CreateScenarioTopic
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.CreateTopic(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.ScenarioTopic
	_ = utils.Map(&dto, m)
	common.Created(c, "scenario topic created", dto)
}

// UpdateTopic godoc
// @Summary      Update a scenario topic (admin)
// @Tags         scenario-topics
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "topic id"
// @Param        request body requests.UpdateScenarioTopic true "topic payload"
// @Success      200 {object} common.ApiResponse{data=responses.ScenarioTopic}
// @Router       /scenario-topics/{id} [put]
func (h *ScenarioController) UpdateTopic(c *gin.Context) {
	var req requests.UpdateScenarioTopic
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.UpdateTopic(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.ScenarioTopic
	_ = utils.Map(&dto, m)
	common.OK(c, "scenario topic updated", dto)
}

// DeleteTopic godoc
// @Summary      Delete a scenario topic (admin)
// @Tags         scenario-topics
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "topic id"
// @Success      200 {object} common.ApiResponse
// @Router       /scenario-topics/{id} [delete]
func (h *ScenarioController) DeleteTopic(c *gin.Context) {
	if err := h.service.DeleteTopic(c.Request.Context(), c.Param("id")); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "scenario topic deleted", nil)
}
