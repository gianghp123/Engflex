// Package lessons wires the curriculum module: units on the path (sections
// derived from CEFR), ordered activities, and the caller's bookmarks.
// Lesson attempts are recorded here (1 attempt = 1 lesson); progress
// rollups stay outside.
package lessons

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/config"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/llm"
	"engflex-api/internal/modules/lessons/controllers"
	"engflex-api/internal/modules/lessons/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/voice"
	"engflex-api/internal/voice/pronunciation"
)

// RegisterRoutes mounts the lesson endpoints under /api/v1.
// sessionLimiter is an optional IP-keyed throttle (nil = disabled) applied
// BEFORE auth on the engine-costly pronounce route (conversations parity:
// limiter precedes auth because the userID is unavailable pre-auth).
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, voiceCfg config.VoiceConfig, llmCfg config.LLMConfig, media config.MediaConfig, sessionLimiter gin.HandlerFunc) {
	lessonSvc := services.NewLessonService(
		repositories.NewLessonSectionRepository(db),
		repositories.NewLessonRepository(db),
	)
	lessonCtl := controllers.NewLessonController(lessonSvc, media)

	activitySvc := services.NewLessonActivityService(
		repositories.NewLessonActivityRepository(db),
		pronunciation.NewScorer(voice.NewHTTPPronounceEngine(voiceCfg.PronounceURL, voiceCfg.PronounceTimeout)),
		llm.NewHTTPLLMClient(llmCfg),
		repositories.NewAttemptRepository(db),
	)
	activityCtl := controllers.NewLessonActivityController(activitySvc, media)

	bookmarkSvc := services.NewLessonBookmarkService(repositories.NewLessonBookmarkRepository(db))
	bookmarkCtl := controllers.NewLessonBookmarkController(bookmarkSvc)

	// Lessons/activities/sections are 100% admin-curated, so mutations are
	// admin-gated outright; reads stay open to any authenticated user,
	// filtered to published content by the service (sections have no draft
	// concept of their own, same as video/vocabulary categories).
	l := rg.Group("/lessons")
	l.GET("", middleware.RequireAuth(), lessonCtl.List)
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

	lb := rg.Group("/lesson-bookmarks")
	lb.GET("", middleware.RequireAuth(), bookmarkCtl.List)
	lb.POST("", middleware.RequireAuth(), bookmarkCtl.Save)
	lb.DELETE("/:lessonId", middleware.RequireAuth(), bookmarkCtl.Unsave)

	// Admin-only: full (unstripped) activity config, needed to prefill the
	// edit form — the learner-shaped GET /lesson-activities/:id strips
	// answer keys and voice ids.
	admin := rg.Group("/admin")
	admin.GET("/lesson-activities/:id", middleware.RequireAuth(), middleware.RequireAdmin(), activityCtl.AdminGetByID)
}
