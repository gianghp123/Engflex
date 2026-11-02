// Package conversations registers routes over dependencies supplied by composition.
package conversations

import (
	"engflex-api/internal/modules/conversations/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, ctl *controllers.ConversationController, cb *controllers.VoiceCallbackController, tcb *controllers.TurnCallbackController, sessionLimiter gin.HandlerFunc, internalSecret string) {
	g := rg.Group("/conversations")
	strict := []gin.HandlerFunc{}
	if sessionLimiter != nil {
		strict = append(strict, sessionLimiter)
	}
	strictAuth := append(append([]gin.HandlerFunc{}, strict...), middleware.RequireAuth())
	g.POST("", append(strictAuth, ctl.Create)...)
	g.GET("", middleware.RequireAuth(), ctl.List)
	g.GET("/:id", middleware.RequireAuth(), ctl.Get)
	g.POST("/:id/start", append(strictAuth, ctl.Start)...)
	g.POST("/:id/offer", append(strictAuth, ctl.Offer)...)
	g.PATCH("/:id/offer", append(strictAuth, ctl.Offer)...)
	g.POST("/:id/end", middleware.RequireAuth(), ctl.End)
	g.POST("/:id/turns/:position/analyze", append(strictAuth, ctl.AnalyzeTurn)...)
	// One command for the correction modal: the engine owns the reviewed turn.
	// Retake audio uploads here; gin buffers the multipart file in memory, so
	// the controller caps the declared size and the service re-checks the
	// actual bytes before proxying.
	g.POST("/:id/transcript", middleware.RequireAuth(), ctl.Transcript)
	g.POST("/:id/transcribe", middleware.RequireAuth(), ctl.Transcribe)
	g.POST("/:id/pronounce", middleware.RequireAuth(), ctl.Pronounce)

	internal := rg.Group("/internal/conversations", middleware.RequireInternalSecret(internalSecret))
	internal.POST("/:id/finalize", cb.Finalize)
	internal.POST("/:id/turns", tcb.IngestTurns)
}
