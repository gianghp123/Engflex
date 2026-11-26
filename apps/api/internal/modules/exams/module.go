package exams

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/config"
	"engflex-api/internal/database/repositories"
	"engflex-api/internal/llm"
	"engflex-api/internal/modules/access"
	accessservices "engflex-api/internal/modules/access/services"
	"engflex-api/internal/modules/exams/controllers"
	"engflex-api/internal/modules/exams/services"
	"engflex-api/internal/server/middleware"
)

func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, llmCfg config.LLMConfig, accessSvc *access.Services) {
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

	aiReviews := accessservices.NewAIReviewService(
		attempts, tests, repositories.NewAIOutputRepository(db),
		llm.NewHTTPLLMClient(llmCfg), accessPolicy(accessSvc), accessUsage(accessSvc), accessEnforce(accessSvc),
		repositories.NewTransactionManager(db),
	)
	aiCtl := controllers.NewAIReviewController(aiReviews)
	rg.POST("/exam-attempts/:id/ai-review", middleware.RequireAuth(), aiCtl.GenerateExamReview)
	rg.GET("/exam-attempts/:id/ai-review", middleware.RequireAuth(), aiCtl.GetExamReview)
	rg.POST("/exam-attempts/:id/study-plan", middleware.RequireAuth(), aiCtl.GenerateStudyPlan)
	rg.GET("/exam-attempts/:id/study-plan", middleware.RequireAuth(), aiCtl.GetStudyPlan)
}

// accessPolicy returns the policy decider, tolerating a nil bundle in
// route-table tests that never serve traffic through the AI routes.
func accessPolicy(accessSvc *access.Services) accessservices.PolicyDecider {
	if accessSvc == nil {
		return nil
	}
	return accessSvc.Policy
}

// accessUsage returns the usage meter, tolerating a nil bundle as above.
func accessUsage(accessSvc *access.Services) accessservices.UsageMeter {
	if accessSvc == nil {
		return nil
	}
	return accessSvc.Usage
}

// accessEnforce returns the enforce flag, defaulting to shadow mode on nil.
func accessEnforce(accessSvc *access.Services) bool {
	if accessSvc == nil {
		return false
	}
	return accessSvc.Enforce
}
