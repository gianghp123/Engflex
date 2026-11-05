package exams

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/exams/controllers"
	"engflex-api/internal/modules/exams/services"
	"engflex-api/internal/server/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	categories := repositories.NewExamCategoryRepository(db)
	tests := repositories.NewExamTestRepository(db)
	attempts := repositories.NewAttemptRepository(db)
	ctl := controllers.NewExamController(services.NewCatalogService(categories, tests), services.NewAttemptService(attempts, tests, repositories.NewTransactionManager(db)))
	rg.GET("/exam-categories", middleware.RequireAuth(), ctl.ListCategories)
	rg.GET("/exam-categories/:categoryId/tests", middleware.RequireAuth(), ctl.ListTests)
	rg.GET("/exam-tests/:id", middleware.RequireAuth(), ctl.GetTest)
	rg.POST("/exam-tests/:id/attempts", middleware.RequireAuth(), ctl.Start)
	rg.GET("/exam-attempts", middleware.RequireAuth(), ctl.ListAttempts)
	rg.GET("/exam-attempts/statistics", middleware.RequireAuth(), ctl.Statistics)
	rg.POST("/exam-attempts/:id/cancel", middleware.RequireAuth(), ctl.CancelAttempt)
	rg.PATCH("/exam-attempts/:id", middleware.RequireAuth(), ctl.PatchAttempt)
	rg.GET("/exam-attempts/:id", middleware.RequireAuth(), ctl.GetAttempt)
}
