package controllers

import (
	"github.com/gin-gonic/gin"

	"engflex-api/internal/common"
	"engflex-api/internal/modules/scenarios/dtos/requests"
	"engflex-api/internal/modules/scenarios/dtos/responses"
	"engflex-api/internal/modules/scenarios/services"
	"engflex-api/internal/utils"
)

// PersonaController exposes the persona catalog over HTTP.
type PersonaController struct {
	service *services.PersonaService
}

// NewPersonaController wires the controller to the service.
func NewPersonaController(service *services.PersonaService) *PersonaController {
	return &PersonaController{service: service}
}

// List godoc
// @Summary      List personas
// @Tags         personas
// @Security     BearerAuth
// @Produce      json
// @Param        page query int false "page"
// @Param        pageSize query int false "page size"
// @Success      200 {object} common.ApiResponse{data=[]responses.Persona}
// @Router       /personas [get]
func (h *PersonaController) List(c *gin.Context) {
	page, pageSize := utils.ParsePagination(c)
	items, total, err := h.service.List(c.Request.Context(), requests.ListPersonas{
		ListParams: common.ListParams{Page: page, PageSize: pageSize},
	})
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dtos []responses.Persona
	_ = utils.MapSlice(&dtos, items)
	common.Paginated(c, "personas", dtos, page, pageSize, total)
}

// Create godoc
// @Summary      Create a persona (admin)
// @Tags         personas
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body requests.CreatePersona true "persona payload"
// @Success      201 {object} common.ApiResponse{data=responses.Persona}
// @Router       /personas [post]
func (h *PersonaController) Create(c *gin.Context) {
	var req requests.CreatePersona
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Persona
	_ = utils.Map(&dto, m)
	common.Created(c, "persona created", dto)
}

// Update godoc
// @Summary      Update a persona (admin)
// @Tags         personas
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "persona id"
// @Param        request body requests.UpdatePersona true "persona payload"
// @Success      200 {object} common.ApiResponse{data=responses.Persona}
// @Router       /personas/{id} [put]
func (h *PersonaController) Update(c *gin.Context) {
	var req requests.UpdatePersona
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, common.BadRequest(err.Error()))
		return
	}
	m, err := h.service.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		common.Fail(c, err)
		return
	}
	var dto responses.Persona
	_ = utils.Map(&dto, m)
	common.OK(c, "persona updated", dto)
}

// Delete godoc
// @Summary      Delete a persona (admin)
// @Tags         personas
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "persona id"
// @Success      200 {object} common.ApiResponse
// @Router       /personas/{id} [delete]
func (h *PersonaController) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), c.Param("id")); err != nil {
		common.Fail(c, err)
		return
	}
	common.OK(c, "persona deleted", nil)
}
