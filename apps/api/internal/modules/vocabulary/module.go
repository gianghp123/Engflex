// Package vocabulary wires the vocabulary module: the shared dictionary,
// user flashcard decks and their rows, the deck taxonomy, and per-user word
// state. All five tables are served by this one module because they are one
// feature.
package vocabulary

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"engflex-api/internal/database/repositories"
	"engflex-api/internal/modules/vocabulary/controllers"
	"engflex-api/internal/modules/vocabulary/services"
	"engflex-api/internal/server/middleware"
)

// RegisterRoutes mounts the vocabulary endpoints under /api/v1.
func RegisterRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	categorySvc := services.NewVocabularyCategoryService(repositories.NewVocabularyCategoryRepository(db))
	categoryCtl := controllers.NewVocabularyCategoryController(categorySvc)

	deckSvc := services.NewVocabularyDeckService(repositories.NewVocabularyDeckRepository(db), repositories.NewVocabularyDeckItemRepository(db))
	deckCtl := controllers.NewVocabularyDeckController(deckSvc)

	itemRepo := repositories.NewVocabularyItemRepository(db)
	userRepo := repositories.NewUserVocabularyRepository(db)

	itemSvc := services.NewVocabularyDeckItemService(repositories.NewVocabularyDeckItemRepository(db), repositories.NewVocabularyDeckRepository(db), itemRepo)
	itemCtl := controllers.NewVocabularyDeckItemController(itemSvc)

	dictSvc := services.NewVocabularyItemService(itemRepo, userRepo, repositories.NewVocabularyDeckRepository(db), repositories.NewVocabularyDeckItemRepository(db))
	dictCtl := controllers.NewVocabularyItemController(dictSvc)

	stateSvc := services.NewUserVocabularyService(userRepo, itemRepo)
	stateCtl := controllers.NewUserVocabularyController(stateSvc, dictSvc)

	c := rg.Group("/vocabulary-categories")
	c.GET("", middleware.RequireAuth(), categoryCtl.List)
	c.GET("/:id", middleware.RequireAuth(), categoryCtl.GetByID)
	c.POST("", middleware.RequireAuth(), middleware.RequireAdmin(), categoryCtl.Create)
	c.PUT("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), categoryCtl.Update)
	c.DELETE("/:id", middleware.RequireAuth(), middleware.RequireAdmin(), categoryCtl.Delete)

	d := rg.Group("/vocabulary-decks")
	d.GET("", middleware.RequireAuth(), deckCtl.List)
	d.GET("/:id", middleware.RequireAuth(), deckCtl.GetByID)
	d.POST("", middleware.RequireAuth(), deckCtl.Create)
	d.PUT("/:id", middleware.RequireAuth(), deckCtl.Update)
	d.DELETE("/:id", middleware.RequireAuth(), deckCtl.Delete)

	i := rg.Group("/vocabulary-deck-items")
	i.GET("", middleware.RequireAuth(), itemCtl.List)
	i.GET("/:id", middleware.RequireAuth(), itemCtl.GetByID)
	i.POST("", middleware.RequireAuth(), itemCtl.Create)
	i.PUT("/:id", middleware.RequireAuth(), itemCtl.Update)
	i.DELETE("/:id", middleware.RequireAuth(), itemCtl.Delete)

	v := rg.Group("/vocabulary-items")
	v.GET("", middleware.RequireAuth(), dictCtl.List)
	v.GET("/:id", middleware.RequireAuth(), dictCtl.GetByID)
	v.POST("", middleware.RequireAuth(), dictCtl.Create)

	u := rg.Group("/user-vocabulary")
	u.GET("", middleware.RequireAuth(), stateCtl.List)
	u.GET("/stats", middleware.RequireAuth(), stateCtl.Stats)
	u.POST("", middleware.RequireAuth(), stateCtl.Save)
	u.POST("/words", middleware.RequireAuth(), stateCtl.AddWord)
	u.POST("/:itemId/review", middleware.RequireAuth(), stateCtl.Review)
	u.PATCH("/:itemId", middleware.RequireAuth(), stateCtl.Patch)
	u.DELETE("/:itemId", middleware.RequireAuth(), stateCtl.Unsave)
}
