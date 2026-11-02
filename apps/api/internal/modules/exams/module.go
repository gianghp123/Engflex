// Package exams registers routes over dependencies supplied by composition.
package exams

import (
	"engflex-api/internal/modules/exams/controllers"
	"engflex-api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(rg *gin.RouterGroup, ctl *controllers.ExamController) {
	rg.GET("/exam-categories", middleware.RequireAuth(), ctl.ListCategories)
	rg.GET("/exam-categories/:categoryId/tests", middleware.RequireAuth(), ctl.ListTests)
	rg.GET("/exam-tests/:id", middleware.RequireAuth(), ctl.GetTest)
	rg.POST("/exam-tests/:id/attempts", middleware.RequireAuth(), ctl.Start)
	rg.GET("/exam-attempts", middleware.RequireAuth(), ctl.ListAttempts)
	rg.GET("/exam-attempts/stats", middleware.RequireAuth(), ctl.Stats)
	rg.POST("/exam-attempts/:id/cancel", middleware.RequireAuth(), ctl.CancelAttempt)
	rg.PATCH("/exam-attempts/:id", middleware.RequireAuth(), ctl.PatchAttempt)
	rg.GET("/exam-attempts/:id", middleware.RequireAuth(), ctl.GetAttempt)
}
