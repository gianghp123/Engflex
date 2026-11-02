// Package videos registers routes over dependencies supplied by composition.
package videos

import (
	"engflex-api/internal/modules/videos/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, exerciseCtl *controllers.VideoExerciseController, attemptCtl *controllers.VideoAttemptController, transcriptCtl *controllers.VideoTranscriptController, categoryCtl *controllers.VideoCategoryController) {
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
	e.POST("/:id/attempts", middleware.RequireAuth(), attemptCtl.StartAttempt)
	e.GET("/:id/attempts/open", middleware.RequireAuth(), attemptCtl.GetOpenAttempt)
	e.PUT("/:id/attempts/:attemptId/progress", middleware.RequireAuth(), attemptCtl.SaveProgress)
	e.POST("/:id/attempts/:attemptId/complete", middleware.RequireAuth(), attemptCtl.CompleteAttempt)

	t := rg.Group("/video-transcripts")
	t.GET("", middleware.RequireAuth(), middleware.RequireAdmin(), transcriptCtl.List)
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
	c.GET("/preview", middleware.RequireAuth(), categoryCtl.ListWithPreview)
}
