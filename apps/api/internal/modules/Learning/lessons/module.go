// Package lessons registers routes over dependencies supplied by composition.
package lessons

import (
	"engflex-api/internal/modules/lessons/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, lessonCtl *controllers.LessonController, activityCtl *controllers.LessonActivityController, sessionLimiter gin.HandlerFunc) {
	// Lessons/activities/sections are 100% admin-curated, so mutations are
	// admin-gated outright; reads stay open to any authenticated user,
	// filtered to published content by the service (sections have no draft
	// concept of their own, same as video/vocabulary categories).
	l := rg.Group("/lessons")
	l.GET("", middleware.RequireAuth(), lessonCtl.List)
	l.GET("/progress", middleware.RequireAuth(), activityCtl.ListProgress)
	l.GET("/:id", middleware.RequireAuth(), lessonCtl.GetByID)
	l.GET("/by-slug/:slug", middleware.RequireAuth(), lessonCtl.GetBySlug)
	l.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), lessonCtl.Create)
	l.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), lessonCtl.Update)
	l.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), lessonCtl.Delete)
	l.POST("/:id/attempts", middleware.RequireAuth(), activityCtl.StartAttempt)
	l.GET("/:id/attempts/open", middleware.RequireAuth(), activityCtl.GetOpenAttempt)

	ls := rg.Group("/lesson-sections")
	ls.GET("", middleware.RequireAuth(), lessonCtl.ListSections)
	ls.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), lessonCtl.CreateSection)
	ls.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), lessonCtl.UpdateSection)
	ls.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), lessonCtl.DeleteSection)

	la := rg.Group("/lesson-activities")
	la.GET("", middleware.RequireAuth(), activityCtl.List)
	la.GET("/:id", middleware.RequireAuth(), activityCtl.GetByID)
	la.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), activityCtl.Create)
	la.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), activityCtl.Update)
	la.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), activityCtl.Delete)
	la.POST("/:id/check", middleware.RequireAuth(), activityCtl.CheckAnswer)
	strict := []gin.HandlerFunc{}
	if sessionLimiter != nil {
		strict = append(strict, sessionLimiter)
	}
	la.POST("/:id/pronounce", append(strict, middleware.RequireAuth(), activityCtl.Pronounce)...)
	la.POST("/:id/score", append(strict, middleware.RequireAuth(), activityCtl.ScoreWriting)...)

	// Admin-only: full (unstripped) activity config, needed to prefill the
	// edit form — the learner-shaped GET /lesson-activities/:id strips
	// answer keys and voice ids.
	admin := rg.Group("/admin")
	admin.GET("/lesson-activities/:id", middleware.RequireAuth(), middleware.RequireAdmin(), activityCtl.AdminGetByID)
}
