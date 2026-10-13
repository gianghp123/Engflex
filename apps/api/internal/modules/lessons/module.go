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
	audioSvc := services.NewAudioSynthesisService(services.ExecRunner{})
	activityCtl := controllers.NewLessonActivityController(activitySvc, audioSvc, media)

	bookmarkSvc := services.NewLessonBookmarkService(repositories.NewLessonBookmarkRepository(db))
	bookmarkCtl := controllers.NewLessonBookmarkController(bookmarkSvc)

	l := rg.Group("/lessons")
	l.GET("", middleware.RequireAuth(), lessonCtl.List)
	l.GET("/:id", middleware.RequireAuth(), lessonCtl.GetByID)
	l.GET("/by-slug/:slug", middleware.RequireAuth(), lessonCtl.GetBySlug)
	l.POST("/:id/attempts", middleware.RequireAuth(), activityCtl.StartAttempt)
	l.GET("/:id/attempts/open", middleware.RequireAuth(), activityCtl.GetOpenAttempt)

	la := rg.Group("/lesson-activities")
	la.GET("", middleware.RequireAuth(), activityCtl.List)
	la.GET("/:id", middleware.RequireAuth(), activityCtl.GetByID)
	la.POST("/:id/check", middleware.RequireAuth(), activityCtl.CheckAnswer)
	strict := []gin.HandlerFunc{}
	if sessionLimiter != nil {
		strict = append(strict, sessionLimiter)
	}
	la.POST("/:id/pronounce", append(strict, middleware.RequireAuth(), activityCtl.Pronounce)...)
	la.POST("/:id/score", append(strict, middleware.RequireAuth(), activityCtl.ScoreWriting)...)
	la.POST("/:id/synthesize-audio", middleware.RequireAuth(), middleware.RequireAdmin(), activityCtl.SynthesizeAudio)

	lb := rg.Group("/lesson-bookmarks")
	lb.GET("", middleware.RequireAuth(), bookmarkCtl.List)
	lb.POST("", middleware.RequireAuth(), bookmarkCtl.Save)
	lb.DELETE("/:lessonId", middleware.RequireAuth(), bookmarkCtl.Unsave)
}
