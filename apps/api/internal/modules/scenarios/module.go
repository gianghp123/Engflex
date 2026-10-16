// Package scenarios wires the scenarios module.
package scenarios

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/scenarios/controllers"
	"engflex-api/internal/modules/scenarios/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts the catalog endpoints under /api/v1: scenarios,
// scenario topics, and personas (one feature, three route groups).
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	svc := services.NewScenarioService(
		repositories.NewScenarioRepository(db), repositories.NewScenarioTopicRepository(db))
	ctl := controllers.NewScenarioController(svc)

	personaSvc := services.NewPersonaService(repositories.NewPersonaRepository(db))
	personaCtl := controllers.NewPersonaController(personaSvc)

	// All three resources are 100% admin-curated, so mutations are
	// admin-gated outright; reads stay open to any authenticated user, with
	// scenarios filtered to published content by the service (topics/
	// personas are taxonomy/catalog with no draft concept of their own).
	g := rg.Group("/scenarios")
	g.GET("", middleware.RequireAuth(), ctl.List)
	g.GET("/:id", middleware.RequireAuth(), ctl.GetByID)
	g.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.Create)
	g.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.Update)
	g.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.Delete)

	t := rg.Group("/scenario-topics")
	t.GET("", middleware.RequireAuth(), ctl.TopicsWithPreview)
	t.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.CreateTopic)
	t.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.UpdateTopic)
	t.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.DeleteTopic)

	p := rg.Group("/personas")
	p.GET("", middleware.RequireAuth(), personaCtl.List)
	p.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), personaCtl.Create)
	p.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), personaCtl.Update)
	p.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), personaCtl.Delete)
}
