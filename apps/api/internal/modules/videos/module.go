// Package videos wires the video-exercise module: exercises, their caption
// transcripts, and the content-format taxonomy. All three tables are served
// by this one module because they are one feature.
package videos

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/videos/controllers"
	"engflex-api/internal/modules/videos/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts the video endpoints under /api/v1.
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	exerciseSvc := services.NewVideoExerciseService(repositories.NewVideoExerciseRepository(db))
	exerciseCtl := controllers.NewVideoExerciseController(exerciseSvc)

	transcriptSvc := services.NewVideoTranscriptService(repositories.NewVideoTranscriptRepository(db))
	transcriptCtl := controllers.NewVideoTranscriptController(transcriptSvc)

	categorySvc := services.NewVideoCategoryService(repositories.NewVideoCategoryRepository(db))
	categoryCtl := controllers.NewVideoCategoryController(categorySvc)

	// All three resources are 100% admin-curated (no user-owned variant like
	// vocabulary decks have), so mutations are admin-gated outright; reads
	// stay open to any authenticated user, filtered to published content by
	// the service for video-exercises (video-categories/transcripts are
	// pure taxonomy/captions with no draft concept of their own).
	e := rg.Group("/video-exercises")
	e.GET("", middleware.RequireAuth(), exerciseCtl.List)
	e.GET("/:id", middleware.RequireAuth(), exerciseCtl.GetByID)
	e.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), exerciseCtl.Create)
	e.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), exerciseCtl.Update)
	e.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), exerciseCtl.Delete)

	t := rg.Group("/video-transcripts")
	t.GET("", middleware.RequireAuth(), transcriptCtl.List)
	t.GET("/:id", middleware.RequireAuth(), transcriptCtl.GetByID)
	t.GET("/video-exercise/:video_exercise_id", middleware.RequireAuth(), transcriptCtl.GetByVideoExerciseID)
	t.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), transcriptCtl.Create)
	t.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), transcriptCtl.Update)
	t.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), transcriptCtl.Delete)

	c := rg.Group("/video-categories")
	c.GET("", middleware.RequireAuth(), categoryCtl.List)
	c.GET("/:id", middleware.RequireAuth(), categoryCtl.GetByID)
	c.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), categoryCtl.Create)
	c.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), categoryCtl.Update)
	c.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), categoryCtl.Delete)
}
