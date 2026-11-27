package stories

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/config"
	"engflex-api/internal/database/repositories"
	paymentservices "engflex-api/internal/modules/payments/services"
	"engflex-api/internal/modules/stories/controllers"
	"engflex-api/internal/modules/stories/services"
	"engflex-api/internal/server/middleware"
	"engflex-api/internal/storage"
)

// RegisterRoutes mounts the stories endpoints under /api/v1.
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB, minioCfg config.MinioConfig) {
	repo := repositories.NewStoryRepository(db)
	premiumSvc := paymentservices.NewSubscriptionService(
		repositories.NewSubscriptionRepository(db), repositories.NewPlanRepository(db))
	svc := services.NewStoryService(repo, premiumSvc, repositories.NewTransactionManager(db))
	// A malformed/absent MinIO endpoint (an empty Config in a unit test, or
	// a deployment that genuinely has no object store) shouldn't take the
	// rest of the stories module down with it — only /stories/upload needs
	// this client, so a nil client there degrades to a 503, not a crash.
	storageClient, err := storage.NewClient(minioCfg)
	if err != nil {
		slog.Warn("stories: minio client unavailable, uploads disabled", "error", err)
		storageClient = nil
	}
	ctl := controllers.NewStoryController(svc, storageClient)

	g := rg.Group("/stories")
	g.GET("", middleware.RequireAuth(), ctl.List)
	g.GET("/topics", middleware.RequireAuth(), ctl.ListTopics)
	g.GET("/:id", middleware.RequireAuth(), ctl.GetByID)
	g.GET("/by-slug/:slug", middleware.RequireAuth(), ctl.GetBySlug)
	g.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.Create)
	g.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.Update)
	g.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.Delete)
	g.POST("/upload", middleware.RequireAuth(), middleware.RequireAdmin(), ctl.Upload)
}
